package postgres

import (
	"fmt"
	"math"
	"strings"

	"github.com/mtekmir/warehouse-service/internal/errors"
	"github.com/mtekmir/warehouse-service/internal/product"
)

type productRepo struct{}

func (productRepo) ExistingProductsMap(db product.Executor, bb []*product.Barcode) (map[product.Barcode]product.ID, error) {
	var op errors.Op = "productRepo.existingProductsMap"

	pHolders := make([]string, 0, len(bb))
	values := make([]interface{}, 0, len(bb))
	for i, b := range bb {
		pHolders = append(pHolders, fmt.Sprintf("$%d", i+1))
		values = append(values, b)
	}

	stmt := fmt.Sprintf(`SELECT id, barcode FROM products WHERE barcode IN (%s)`, strings.Join(pHolders, ","))
	rows, err := db.Query(stmt, values...)
	if err != nil {
		return nil, errors.E(op, err)
	}
	defer rows.Close()

	m := make(map[product.Barcode]product.ID, len(bb))
	for rows.Next() {
		var p product.Product
		err := rows.Scan(&p.ID, &p.Barcode)
		if err != nil {
			return nil, errors.E(op, err)
		}
		m[p.Barcode] = p.ID
	}

	return m, nil
}

func (productRepo) FindAll(db product.Executor, ff *product.Filters) ([]*product.StockInfo, error) {
	var op errors.Op = "productRepo.findAll"

	filterQueries := make([]string, 0, 2)
	var values []interface{}

	if ff.BB != nil {
		pHolders := make([]string, 0, len(*ff.BB))
		for i, b := range *ff.BB {
			pHolders = append(pHolders, fmt.Sprintf("$%d", i+1))
			values = append(values, b)
		}
		filterQueries = append(filterQueries, fmt.Sprintf("p.barcode IN (%s)", strings.Join(pHolders, ",")))
	}

	if ff.ID != nil {
		filterQueries = append(filterQueries, fmt.Sprintf("p.id = %d", len(values)+1))
	}

	var filters string
	if len(filterQueries) > 0 {
		filters = fmt.Sprintf("WHERE %s", strings.Join(filterQueries, " AND "))
	}

	stmt := fmt.Sprintf(`
		SELECT p.id, p.barcode, p.name, 
		a.id, a.art_id, a.name, pa.amount, a.stock
		FROM products p
		JOIN product_articles pa ON p.id = pa.product_id
		JOIN articles a ON a.id = pa.article_id
		%s
	`, filters)

	rows, err := db.Query(stmt, values...)
	if err != nil {
		return nil, errors.E(op, err)
	}
	defer rows.Close()

	pp := map[product.ID]*product.StockInfo{}
	var order []product.ID

	for rows.Next() {
		var p product.StockInfo
		var art product.ArticleStock

		err := rows.Scan(&p.ID, &p.Barcode, &p.Name, &art.ID, &art.ArtID, &art.Name, &art.RequiredAmount, &art.Stock)
		if err != nil {
			return nil, errors.E(op, err)
		}

		if found, ok := pp[p.ID]; ok {
			found.Articles = append(found.Articles, &art)
			continue
		}
		p.Articles = []*product.ArticleStock{&art}
		pp[p.ID] = &p
		order = append(order, p.ID)
	}

	res := make([]*product.StockInfo, 0, len(order))
	for _, id := range order {
		p := pp[id]
		minStock := math.MaxInt64
		for _, art := range p.Articles {
			if art.Stock/art.RequiredAmount < minStock {
				minStock = art.Stock / art.RequiredAmount
			}
		}
		p.AvailableQty = minStock
		res = append(res, p)
	}

	return res, nil
}

func (productRepo) BatchInsert(db product.Executor, pp []*product.Product) ([]*product.Product, error) {
	var op errors.Op = "productRepo.batchInsert"

	values := make([]interface{}, 0, len(pp))
	pHolders := make([]string, 0, len(pp))
	for i, p := range pp {
		ph := make([]string, 0, 2)
		for j := 1; j < 3; j++ {
			ph = append(ph, fmt.Sprintf("$%d", i*2+j))
		}
		pHolders = append(pHolders, "("+strings.Join(ph, ", ")+")")
		values = append(values, p.Barcode, p.Name)
	}

	stmt := fmt.Sprintf("INSERT INTO products (barcode, name) VALUES %s RETURNING id, barcode, name", strings.Join(pHolders, ", "))

	rows, err := db.Query(stmt, values...)
	if err != nil {
		return nil, errors.E(op, err)
	}
	defer rows.Close()

	inserted := make([]*product.Product, 0, len(pp))

	for rows.Next() {
		var art product.Product
		if err := rows.Scan(&art.ID, &art.Barcode, &art.Name); err != nil {
			return nil, errors.E(op, err)
		}
		inserted = append(inserted, &art)
	}

	return inserted, nil
}

func (productRepo) InsertProductArticles(db product.Executor, arts []*product.ArticleRow) error {
	var op errors.Op = "productRepo.insertProductArticles"

	pHolders := make([]string, 0, len(arts))
	values := make([]interface{}, 0, len(arts)*3)
	for i, art := range arts {
		ph := make([]string, 0, 3)
		for j := 1; j < 4; j++ {
			ph = append(ph, fmt.Sprintf("$%d", i*3+j))
		}
		pHolders = append(pHolders, "("+strings.Join(ph, ", ")+")")
		values = append(values, art.Amount, art.ProductID, art.ID)
	}

	stmt := fmt.Sprintf(`
		INSERT INTO product_articles (amount, product_id, article_id) VALUES %s
	`, strings.Join(pHolders, ", "))

	_, err := db.Exec(stmt, values...)
	if err != nil {
		return errors.E(op, err)
	}

	return nil
}

// NewProductRepo returns a new product repo.
func NewProductRepo() product.Repo {
	return productRepo{}
}
