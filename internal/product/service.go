package product

import (
	"database/sql"

	"github.com/mtekmir/warehouse-service/internal/article"
	"github.com/mtekmir/warehouse-service/internal/errors"
)

// Execer provides an interface for required db methods.
type Execer interface {
	QueryRow(query string, args ...interface{}) *sql.Row
	Exec(query string, args ...interface{}) (sql.Result, error)
	Query(query string, args ...interface{}) (*sql.Rows, error)
}

// Repo provides methods for managing products in a db.
type Repo interface {
	FindAll(Execer, *[]Barcode) ([]*Product, error)
	BatchInsert(Execer, []*Product) ([]*Product, error)
	InsertProductArticles(Execer, []*ArticleRow) error
	ExistingProductsMap(Execer, []*Barcode) (map[Barcode]ID, error)
}

// Service exposes methods on products.
type Service struct {
	db          *sql.DB
	productRepo Repo
	articleRepo article.Repo
}

// Import products.
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

	// Create non-existing products
	ppToCreate := make([]*Product, 0, len(rows)-len(existingM))
	for _, r := range rows {
		if _, ok := existingM[r.Barcode]; !ok {
			ppToCreate = append(ppToCreate, r)
		}
	}

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
