package postgres

import (
	"fmt"
	"strings"

	"github.com/mtekmir/warehouse-service/internal/article"
	"github.com/mtekmir/warehouse-service/internal/errors"
)

type articleRepo struct{}

func (r articleRepo) Import(db article.Execer, aa []*article.Article) ([]*article.Article, error) {
	var op errors.Op = "articleRepo.import"

	// Handle duplicates
	rowMap := make(map[article.ArtID][]*article.Article)
	for _, art := range aa {
		if existing, ok := rowMap[art.ArtID]; ok {
			existing = append(existing, art)
			continue
		}
		rowMap[art.ArtID] = []*article.Article{art}
	}

	rows := make([]*article.Article, 0, len(rowMap))
	for b, rr := range rowMap {
		art := &article.Article{ArtID: b, Name: rr[0].Name}
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
	existing, err := r.FindAll(db, &artIDs)
	if err != nil {
		return nil, errors.E(op, err)
	}

	existingM := make(map[article.ArtID]*article.Article, len(existing))
	for _, art := range existing {
		existingM[art.ArtID] = art
	}

	//
	// Create non existing ones
	artToCreate := make([]*article.Article, 0, len(rows)-len(existing))
	for _, r := range rows {
		if _, ok := existingM[r.ArtID]; !ok {
			artToCreate = append(artToCreate, &article.Article{Name: r.Name, ArtID: r.ArtID, Stock: r.Stock})
		}
	}

	var created []*article.Article
	if len(artToCreate) > 0 {
		arts, err := r.BatchInsert(db, artToCreate)
		if err != nil {
			return nil, err
		}
		created = arts
	}

	//
	// Update existing articles
	adjustments := make([]*article.QtyAdjustment, 0, len(existing))
	for _, r := range rows {
		if art, ok := existingM[r.ArtID]; ok {
			adjustments = append(adjustments, &article.QtyAdjustment{ID: art.ID, Qty: r.Stock})
			// update quantities in existingM.
			art.Stock += r.Stock
		}
	}

	if len(adjustments) > 0 {
		if err := r.AdjustQuantities(db, article.QtyAdjustmentAdd, adjustments); err != nil {
			return nil, errors.E(op, err)
		}
	}

	return append(existing, created...), nil
}

func (articleRepo) BatchInsert(db article.Execer, arts []*article.Article) ([]*article.Article, error) {
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

	stmt := fmt.Sprintf("INSERT INTO articles (art_id, name, stock) VALUES %s RETURNING id, art_id, name, stock", strings.Join(pHolders, ", "))

	rows, err := db.Query(stmt, values...)
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

func (articleRepo) AdjustQuantities(db article.Execer, t article.QtyAdjustmentKind, changes []*article.QtyAdjustment) error {
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

	res, err := db.Exec(stmt, values...)
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

func (articleRepo) FindAll(db article.Execer, bb *[]article.ArtID) ([]*article.Article, error) {
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
	`, artIDQuery)

	rows, err := db.Query(stmt, values...)
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
