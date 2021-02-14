package product

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/mtekmir/warehouse-service/internal/article"
	"github.com/mtekmir/warehouse-service/internal/errors"
	"github.com/sirupsen/logrus"
)

// Executor provides an interface for required db methods.
type Executor interface {
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
}

// Filters are used to filter get products queries.
type Filters struct {
	BB *[]Barcode
	ID *ID
}

// Repo provides methods for managing products in a db.
type Repo interface {
	FindAll(context.Context, Executor, *Filters) ([]*StockInfo, error)
	BatchInsert(context.Context, Executor, []*Product) ([]*Product, error)
	InsertProductArticles(context.Context, Executor, []*ArticleRow) error
	ExistingProductsMap(context.Context, Executor, []*Barcode) (map[Barcode]ID, error)
}

// Service exposes methods on products.
type Service struct {
	log         *logrus.Logger
	db          *sql.DB
	productRepo Repo
	articleRepo article.Repo
}

// FindAll returns a slice of products with stock information. If barcodes slice is null
// all products will be returned.
func (s *Service) FindAll(ctx context.Context, ff *Filters) ([]*StockInfo, error) {
	var op errors.Op = "productService.findAll"

	pp, err := s.productRepo.FindAll(ctx, s.db, ff)
	if err != nil {
		return nil, errors.E(op, err)
	}

	return pp, nil
}

// Find returns a product with stock information. If not found an error is returned.
func (s *Service) Find(ctx context.Context, ID ID) (*StockInfo, error) {
	var op errors.Op = "productService.find"

	pp, err := s.productRepo.FindAll(ctx, s.db, &Filters{ID: &ID})
	if err != nil {
		return nil, errors.E(op, err)
	}

	if len(pp) == 0 {
		return nil, errors.E(op, errors.NotFound, "Product not found")
	}

	return pp[0], nil
}

// Remove subtracts the quantities of the articles of the product from the repository and returns
// the updated stock information of the product.
func (s *Service) Remove(ctx context.Context, ID ID, qty int) (*StockInfo, error) {
	var op errors.Op = "productService.remove"

	p, err := s.Find(ctx, ID)
	if err != nil {
		return nil, errors.E(op, err)
	}

	if p.AvailableQty < qty {
		return nil, errors.E(op, errors.Invalid, fmt.Sprintf("Insufficient stock quantity. Max available quantity: %d", p.AvailableQty))
	}

	qtyAdjs := []*article.QtyAdjustment{}
	for _, art := range p.Articles {
		qtyAdjs = append(qtyAdjs, &article.QtyAdjustment{ID: art.ID, Qty: qty * art.RequiredAmount})
	}

	if err := s.articleRepo.AdjustQuantities(ctx, s.db, article.QtyAdjustmentSubtract, qtyAdjs); err != nil {
		return nil, errors.E(op, err)
	}

	p.AvailableQty -= qty
	for _, art := range p.Articles {
		art.Stock -= art.RequiredAmount * qty
	}

	return p, nil
}

// Import products. Handles duplicate products. Imports the articles as well.
// If the product exists, it only updates the quantities of the articles.
// If it's a new product, it adds the product and associates the articles with it.
func (s *Service) Import(ctx context.Context, rows []*Product) error {
	var op errors.Op = "productService.import"
	s.log.Printf("Importing %d products", len(rows))

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return errors.E(op, err)
	}

	// Find existing products
	barcodes := make([]*Barcode, 0, len(rows))     // For finding existing products
	arts := make([]*article.Article, 0, len(rows)) // For importing articles

	for _, r := range rows {
		barcodes = append(barcodes, &r.Barcode)
		for _, a := range r.Articles {
			arts = append(arts, &article.Article{ID: a.ID, ArtID: a.ArtID, Name: a.Name, Stock: a.Amount})
		}
	}

	existingM, err := s.productRepo.ExistingProductsMap(ctx, tx, barcodes)
	if err != nil {
		tx.Rollback()
		return errors.E(op, err)
	}

	// Import articles
	insertedArts, err := s.articleRepo.Import(ctx, tx, arts)
	if err != nil {
		tx.Rollback()
		return errors.E(op, err)
	}
	artIDtoID := make(map[article.ArtID]article.ID, len(insertedArts))
	for _, art := range insertedArts {
		artIDtoID[art.ArtID] = art.ID
	}

	// Filter out non-existing products
	ppToCreate := make([]*Product, 0, len(rows)-len(existingM))
	for _, r := range rows {
		if _, ok := existingM[r.Barcode]; !ok {
			ppToCreate = append(ppToCreate, r)
		}
	}

	// Create non-existing products
	if len(ppToCreate) > 0 {

		created, err := s.productRepo.BatchInsert(ctx, tx, ppToCreate)
		if err != nil {
			tx.Rollback()
			return errors.E(op, err)
		}

		createdBarcodeToID := make(map[Barcode]ID, len(created))
		for _, p := range created {
			createdBarcodeToID[p.Barcode] = p.ID
		}

		pArts := make([]*ArticleRow, 0, len(rows))
		for _, p := range rows {
			if pID, ok := createdBarcodeToID[p.Barcode]; ok {
				for _, a := range p.Articles {
					pArts = append(pArts, &ArticleRow{ID: artIDtoID[a.ArtID], ProductID: pID, Amount: a.Amount})
				}
			}
		}

		if len(pArts) > 0 {
			if err := s.productRepo.InsertProductArticles(ctx, tx, pArts); err != nil {
				tx.Rollback()
				return errors.E(op, err)
			}
		}

	}

	if err := tx.Commit(); err != nil {
		return errors.E(op, err)
	}

	return nil
}

// NewService creates a new service with required dependencies.
func NewService(l *logrus.Logger, db *sql.DB, pr Repo, ar article.Repo) *Service {
	return &Service{
		log:         l,
		db:          db,
		productRepo: pr,
		articleRepo: ar,
	}
}
