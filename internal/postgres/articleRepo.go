package postgres

import (
	"fmt"
	"strings"

	"github.com/mtekmir/warehouse-service/internal/article"
	"github.com/mtekmir/warehouse-service/internal/errors"
)

type articleRepo struct{}

func (r articleRepo) Import(db article.Execer, rows []*article.Article) ([]*article.Article, error) {
	var op errors.Op = "articleRepo.import"

	//
	// Find existing articles
	Barcodes := make([]article.Barcode, 0, len(rows))
	for _, r := range rows {
		Barcodes = append(Barcodes, article.Barcode(r.Barcode))
	}
	existing, err := r.FindAllByBarcode(db, &Barcodes)
	if err != nil {
		return nil, errors.E(op, err)
	}

	existingM := make(map[article.Barcode]*article.Article, len(existing))
	for _, art := range existing {
		existingM[art.Barcode] = art
	}

	//
	// Create non existing ones
	artToCreate := make([]*article.Article, 0, len(rows)-len(existing))
	for _, r := range rows {
		if _, ok := existingM[r.Barcode]; !ok {
			artToCreate = append(artToCreate, &article.Article{Name: r.Name, Barcode: r.Barcode, Stock: r.Stock})
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
	// Update existing products
	adjustments := make([]*article.QtyAdjustment, 0, len(existing))
	for _, r := range rows {
		if art, ok := existingM[r.Barcode]; ok {
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
		values = append(values, art.Barcode, art.Name, art.Stock)
	}

	stmt := fmt.Sprintf("INSERT INTO articles (barcode, name, stock) VALUES %s RETURNING id, barcode, name, stock", strings.Join(pHolders, ", "))

	rows, err := db.Query(stmt, values...)
	if err != nil {
		return nil, errors.E(op, err)
	}
	defer rows.Close()

	inserted := make([]*article.Article, 0, len(arts))

	for rows.Next() {
		var art article.Article
		if err := rows.Scan(&art.ID, &art.Barcode, &art.Name, &art.Stock); err != nil {
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

func (articleRepo) FindAllByBarcode(db article.Execer, bb *[]article.Barcode) ([]*article.Article, error) {
	var op errors.Op = "articleRepo.findAllByBarcode"

	var barcodeQ string
	var values []interface{}
	if bb != nil {
		pHolders := make([]string, 0, len(*bb))
		for i, Barcode := range *bb {
			pHolders = append(pHolders, fmt.Sprintf("$%d", i+1))
			values = append(values, Barcode)
		}
		barcodeQ = fmt.Sprintf("WHERE barcode IN (%s)", strings.Join(pHolders, ","))
	}

	stmt := fmt.Sprintf(`
		SELECT id, name, barcode, stock
		FROM articles 
		%s
	`, barcodeQ)

	rows, err := db.Query(stmt, values...)
	if err != nil {
		return nil, errors.E(op, err)
	}
	defer rows.Close()

	articles := []*article.Article{}

	for rows.Next() {
		var art article.Article

		if err := rows.Scan(&art.ID, &art.Name, &art.Barcode, &art.Stock); err != nil {
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
