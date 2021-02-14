package article

import (
	"context"
	"database/sql"

	"github.com/mtekmir/warehouse-service/internal/errors"
	"github.com/sirupsen/logrus"
)

// Executor provides an interface for required db methods.
type Executor interface {
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
}

// Repo provides methods for managing articles in a db.
type Repo interface {
	FindAll(context.Context, Executor, *[]ArtID) ([]*Article, error)
	BatchInsert(context.Context, Executor, []*Article) ([]*Article, error)
	AdjustQuantities(context.Context, Executor, QtyAdjustmentKind, []*QtyAdjustment) error
	Import(context.Context, Executor, []*Article) ([]*Article, error)
}

// Service exposes methods on articles.
type Service struct {
	log  *logrus.Logger
	db   *sql.DB
	repo Repo
}

// Import imports the articles into the DB. New rows will be created for the non-existing
// articles and quantities of existing articles will be updated. Returns the new articles and
// updated articles.
func (s *Service) Import(ctx context.Context, rows []*Article) ([]*Article, error) {
	var op errors.Op = "articleService.import"
	s.log.Printf("Importing %d articles", len(rows))

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, errors.E(op, err)
	}

	arts, err := s.repo.Import(ctx, tx, rows)
	if err != nil {
		tx.Rollback()
		return nil, errors.E(op, err)
	}

	if err := tx.Commit(); err != nil {
		return nil, errors.E(op, err)
	}

	return arts, nil
}

// FindAll returns all the articles in db.
func (s *Service) FindAll(ctx context.Context) ([]*Article, error) {
	var op errors.Op = "articleService.findAll"

	arts, err := s.repo.FindAll(ctx, s.db, nil)
	if err != nil {
		return nil, errors.E(op, err)
	}

	return arts, nil
}

// NewService creates a new service with required dependencies.
func NewService(l *logrus.Logger, db *sql.DB, r Repo) *Service {
	return &Service{
		log:  l,
		db:   db,
		repo: r,
	}
}
