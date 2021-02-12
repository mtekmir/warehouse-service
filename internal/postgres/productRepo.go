package postgres

import (
	"fmt"
	"strings"

	"github.com/mtekmir/warehouse-service/internal/errors"
	"github.com/mtekmir/warehouse-service/internal/product"
)

type productRepo struct{}

func (productRepo) ExistingProductsMap(db product.Execer, bb []*product.Barcode) (map[product.Barcode]product.ID, error) {
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

func (productRepo) FindAll(db product.Execer, bb *[]product.Barcode) ([]*product.Product, error) {
	var op errors.Op = "productRepo.findAllByBarcode"

	var whereInQuery string
	var values []interface{}
	if bb != nil {
		pHolders := make([]string, 0, len(*bb))
		for i, b := range *bb {
			pHolders = append(pHolders, fmt.Sprintf("$%d", i+1))
			values = append(values, b)
		}
		whereInQuery = fmt.Sprintf("WHERE p.barcode IN (%s)", strings.Join(pHolders, ","))
	}

	stmt := fmt.Sprintf(`
		SELECT p.id, p.barcode, p.name, a.id, a.art_id, a.name, pa.amount
		FROM products p
		JOIN product_articles pa ON p.id = pa.product_id
		JOIN articles a ON a.id = pa.article_id
		%s
`, whereInQuery)

	rows, err := db.Query(stmt, values...)
	if err != nil {
		return nil, errors.E(op, err)
	}
	defer rows.Close()

	pp := map[product.ID]*product.Product{}
	var order []product.ID

	for rows.Next() {
		var p product.Product
		var art product.Article
		err := rows.Scan(&p.ID, &p.Barcode, &p.Name, &art.ID, &art.ArtID, &art.Name, &art.Amount)
		if err != nil {
			return nil, errors.E(op, err)
		}
		if found, ok := pp[p.ID]; ok {
			found.Articles = append(found.Articles, &art)
			continue
		}
		p.Articles = []*product.Article{&art}
		pp[p.ID] = &p
		order = append(order, p.ID)
	}

	res := make([]*product.Product, 0, len(order))
	for _, id := range order {
		res = append(res, pp[id])
	}

	return res, nil
}

func (productRepo) BatchInsert(db product.Execer, pp []*product.Product) ([]*product.Product, error) {
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

func (productRepo) InsertProductArticles(db product.Execer, arts []*product.ArticleRow) error {
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
		fmt.Println("asdasd")
		return errors.E(op, err)
	}

	return nil
}

// NewProductRepo returns a new product repo.
func NewProductRepo() product.Repo {
	return productRepo{}
}
