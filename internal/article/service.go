package article

import (
	"database/sql"

	"github.com/mtekmir/warehouse-service/internal/errors"
)

// Executor provides an interface for required db methods.
type Executor interface {
	QueryRow(query string, args ...interface{}) *sql.Row
	Exec(query string, args ...interface{}) (sql.Result, error)
	Query(query string, args ...interface{}) (*sql.Rows, error)
}

// Repo provides methods for managing articles in a db.
type Repo interface {
	FindAll(Executor, *[]ArtID) ([]*Article, error)
	BatchInsert(Executor, []*Article) ([]*Article, error)
	AdjustQuantities(Executor, QtyAdjustmentKind, []*QtyAdjustment) error
	Import(Executor, []*Article) ([]*Article, error)
}

type Service struct {
	db   *sql.DB
	repo Repo
}

func (s *Service) Import(rows []*Article) ([]*Article, error) {
	var op errors.Op = "articleService.import"

	tx, err := s.db.Begin()
	if err != nil {
		return nil, errors.E(op, err)
	}

	arts, err := s.repo.Import(tx, rows)
	if err != nil {
		tx.Rollback()
		return nil, errors.E(op, err)
	}

	if err := tx.Commit(); err != nil {
		return nil, errors.E(op, err)
	}

	return arts, nil
}

// NewService creates a new service with required dependencies.
func NewService(db *sql.DB, r Repo) *Service {
	return &Service{
		db:   db,
		repo: r,
	}
}
