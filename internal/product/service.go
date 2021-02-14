package product

import (
	"database/sql"
	"fmt"

	"github.com/mtekmir/warehouse-service/internal/article"
	"github.com/mtekmir/warehouse-service/internal/errors"
)

// Executor provides an interface for required db methods.
type Executor interface {
	QueryRow(query string, args ...interface{}) *sql.Row
	Exec(query string, args ...interface{}) (sql.Result, error)
	Query(query string, args ...interface{}) (*sql.Rows, error)
}

// Filters are used to filter get products queries.
type Filters struct {
	BB *[]Barcode
	ID *ID
}

// Repo provides methods for managing products in a db.
type Repo interface {
	FindAll(Executor, *Filters) ([]*StockInfo, error)
	BatchInsert(Executor, []*Product) ([]*Product, error)
	InsertProductArticles(Executor, []*ArticleRow) error
	ExistingProductsMap(Executor, []*Barcode) (map[Barcode]ID, error)
}

// Service exposes methods on products.
type Service struct {
	db          *sql.DB
	productRepo Repo
	articleRepo article.Repo
}

// FindAll returns a slice of products with stock information. If barcodes slice is null
// all products will be returned.
func (s *Service) FindAll(ff *Filters) ([]*StockInfo, error) {
	var op errors.Op = "productService.findAll"

	pp, err := s.productRepo.FindAll(s.db, ff)
	if err != nil {
		return nil, errors.E(op, err)
	}

	return pp, nil
}

// Find returns a product with stock information. If not found an error is returned.
func (s *Service) Find(ID ID) (*StockInfo, error) {
	var op errors.Op = "productService.find"

	pp, err := s.productRepo.FindAll(s.db, &Filters{ID: &ID})
	if err != nil {
		return nil, errors.E(op, err)
	}

	if len(pp) == 0 {
		return nil, errors.E(op, errors.NotFound, "Product not found")
	}

	return pp[0], nil
}

// Remove removes the articles of the product from the repository and returns
// the updated stock information of the product.
func (s *Service) Remove(ID ID, qty int) (*StockInfo, error) {
	var op errors.Op = "productService.remove"

	p, err := s.Find(ID)
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

	if err := s.articleRepo.AdjustQuantities(s.db, article.QtyAdjustmentSubtract, qtyAdjs); err != nil {
		return nil, errors.E(op, err)
	}

	p.AvailableQty -= qty
	for _, art := range p.Articles {
		art.Stock -= art.RequiredAmount * qty
	}

	return p, nil
}

// Import products. Handles duplicate products. Imports the articles as well.
func (s *Service) Import(rows []*Product) error {
	var op errors.Op = "productService.import"

	tx, err := s.db.Begin()
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

	existingM, err := s.productRepo.ExistingProductsMap(tx, barcodes)
	if err != nil {
		tx.Rollback()
		return errors.E(op, err)
	}

	// Import articles
	insertedArts, err := s.articleRepo.Import(tx, arts)
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

		created, err := s.productRepo.BatchInsert(tx, ppToCreate)
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
			if err := s.productRepo.InsertProductArticles(tx, pArts); err != nil {
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
func NewService(db *sql.DB, pr Repo, ar article.Repo) *Service {
	return &Service{
		db:          db,
		productRepo: pr,
		articleRepo: ar,
	}
}
