package postgres

import (
	"context"
	"fmt"
	"strings"

	"github.com/mtekmir/warehouse-service/internal/article"
	"github.com/mtekmir/warehouse-service/internal/errors"
)

type articleRepo struct{}

// Imports articles to db. Create and update operations run on separate goroutines.
func (r articleRepo) Import(ctx context.Context, db article.Executor, aa []*article.Article) ([]*article.Article, error) {
	var op errors.Op = "articleRepo.import"

	// Handle duplicates
	order := make([]article.ArtID, 0, len(aa))
	rowMap := make(map[article.ArtID][]*article.Article)
	for _, art := range aa {
		if existing, ok := rowMap[art.ArtID]; ok {
			existing = append(existing, art)
			continue
		}
		order = append(order, art.ArtID)
		rowMap[art.ArtID] = []*article.Article{art}
	}

	rows := make([]*article.Article, 0, len(rowMap))
	for _, id := range order {
		rr := rowMap[id]
		art := &article.Article{ArtID: id, Name: rr[0].Name}
		for _, r := range rr {
			art.Stock += r.Stock
		}
		rows = append(rows, art)
	}

	//
	// Find existing articles
	artIDs := make([]article.ArtID, 0, len(rows))
	for _, r := range rows {
		artIDs = append(artIDs, article.ArtID(r.ArtID))
	}
	existing, err := r.FindAll(ctx, db, &artIDs)
	if err != nil {
		return nil, errors.E(op, err)
	}

	existingM := make(map[article.ArtID]*article.Article, len(existing))
	for _, art := range existing {
		existingM[art.ArtID] = art
	}

	type result struct {
		Arts []*article.Article
		Err  error
	}

	//
	// Create non existing ones
	createdC := make(chan *result)
	go func() {
		artToCreate := make([]*article.Article, 0, len(rows)-len(existing))
		for _, r := range rows {
			if _, ok := existingM[r.ArtID]; !ok {
				artToCreate = append(artToCreate, &article.Article{Name: r.Name, ArtID: r.ArtID, Stock: r.Stock})
			}
		}

		if len(artToCreate) > 0 {
			arts, err := r.BatchInsert(ctx, db, artToCreate)
			if err != nil {
				createdC <- &result{Err: err}
				return
			}
			createdC <- &result{Arts: arts}
			return
		}

		createdC <- &result{}
	}()

	//
	// Update quantities of existing articles
	updatedC := make(chan *result)
	go func(db article.Executor) {
		updated := make([]*article.Article, 0, len(existing))

		adjustments := make([]*article.QtyAdjustment, 0, len(existing))
		for _, r := range rows {
			if art, ok := existingM[r.ArtID]; ok {
				adjustments = append(adjustments, &article.QtyAdjustment{ID: art.ID, Qty: r.Stock})
				// update quantities in existingM.
				art.Stock += r.Stock
				updated = append(updated, art)
			}
		}

		if len(adjustments) > 0 {
			if err := r.AdjustQuantities(ctx, db, article.QtyAdjustmentAdd, adjustments); err != nil {
				updatedC <- &result{Err: err}
				return
			}
			updatedC <- &result{Arts: updated}
			return
		}

		updatedC <- &result{}
	}(db)

	results := make([]*article.Article, 0, len(rows))

	for i := 0; i < 2; i++ {
		select {
		case <-ctx.Done():
			return nil, errors.E(op, ctx.Err())
		case res := <-createdC:
			if res.Err != nil {
				return nil, errors.E(op, res.Err)
			}
			if res.Arts != nil {
				results = append(results, res.Arts...)
			}
		case res := <-updatedC:
			if res.Err != nil {
				return nil, errors.E(op, res.Err)
			}
			if res.Arts != nil {
				results = append(results, res.Arts...)
			}
		}
	}

	return results, nil
}

// BatchInsert inserts an article slice into db. Does not handle duplicates.
func (articleRepo) BatchInsert(ctx context.Context, db article.Executor, arts []*article.Article) ([]*article.Article, error) {
	var op errors.Op = "articleRepo.batchInsert"

	values := make([]interface{}, 0, len(arts))
	pHolders := make([]string, 0, len(arts))
	for i, art := range arts {
		ph := make([]string, 0, 3)
		for j := 1; j < 4; j++ {
			ph = append(ph, fmt.Sprintf("$%d", i*3+j))
		}
		pHolders = append(pHolders, "("+strings.Join(ph, ", ")+")")
		values = append(values, art.ArtID, art.Name, art.Stock)
	}

	stmt := fmt.Sprintf("INSERT INTO articles(art_id, name, stock) VALUES %s RETURNING id, art_id, name, stock", strings.Join(pHolders, ", "))

	rows, err := db.QueryContext(ctx, stmt, values...)
	if err != nil {
		return nil, errors.E(op, err)
	}
	defer rows.Close()

	inserted := make([]*article.Article, 0, len(arts))

	for rows.Next() {
		var art article.Article
		if err := rows.Scan(&art.ID, &art.ArtID, &art.Name, &art.Stock); err != nil {
			return nil, errors.E(op, err)
		}
		inserted = append(inserted, &art)
	}

	return inserted, nil
}

// AdjustQuantities is for updating quantities of articles.
func (articleRepo) AdjustQuantities(ctx context.Context, db article.Executor, t article.QtyAdjustmentKind, changes []*article.QtyAdjustment) error {
	var op errors.Op = "articleRepo.adjustQuantities"

	pHolders := make([]string, 0, len(changes))
	values := make([]interface{}, 0, len(changes)*2)
	for i, c := range changes {
		pHolders = append(pHolders, fmt.Sprintf("($%d::int, $%d::int)", i*2+1, i*2+2))
		values = append(values, c.ID, c.Qty)
	}

	adjustment := "stock = v.q"
	switch t {
	case article.QtyAdjustmentAdd:
		adjustment = "stock = stock + v.q"
	case article.QtyAdjustmentSubtract:
		adjustment = "stock = stock - v.q"
	}

	stmt := fmt.Sprintf(`
		UPDATE articles a SET %s
		FROM (VALUES %s) as v(id, q)
		WHERE a.id = v.id
	`, adjustment, strings.Join(pHolders, ", "))

	res, err := db.ExecContext(ctx, stmt, values...)
	if err != nil {
		return errors.E(op, err)
	}
	count, err := res.RowsAffected()
	if err != nil {
		return errors.E(op, err)
	}
	if int(count) != len(changes) {
		return errors.E(op, "Updated rows don't match with articles length")
	}
	return nil
}

func (articleRepo) FindAll(ctx context.Context, db article.Executor, bb *[]article.ArtID) ([]*article.Article, error) {
	var op errors.Op = "articleRepo.findAll"

	var artIDQuery string
	var values []interface{}
	if bb != nil {
		pHolders := make([]string, 0, len(*bb))
		for i, artID := range *bb {
			pHolders = append(pHolders, fmt.Sprintf("$%d", i+1))
			values = append(values, artID)
		}
		artIDQuery = fmt.Sprintf("WHERE art_id IN (%s)", strings.Join(pHolders, ","))
	}

	stmt := fmt.Sprintf(`
		SELECT id, name, art_id, stock
		FROM articles 
		%s
		ORDER BY art_id
	`, artIDQuery)

	rows, err := db.QueryContext(ctx, stmt, values...)
	if err != nil {
		return nil, errors.E(op, err)
	}
	defer rows.Close()

	articles := []*article.Article{}

	for rows.Next() {
		var art article.Article

		if err := rows.Scan(&art.ID, &art.Name, &art.ArtID, &art.Stock); err != nil {
			return nil, errors.E(op, err)
		}
		articles = append(articles, &art)
	}

	return articles, nil
}

// NewArticleRepo returns a postgres repo for articles.
func NewArticleRepo() article.Repo {
	return articleRepo{}
}
