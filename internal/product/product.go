package product

import (
	"github.com/mtekmir/warehouse-service/internal/article"
)

// ID of a product.
type ID int

// Barcode of a product.
type Barcode string

// Product represents a product that is made of a set of articles.
type Product struct {
	ID       ID         `json:"id"`
	Barcode  Barcode    `json:"barcode"`
	Name     string     `json:"name"`
	Articles []*Article `json:"contain_articles"`
}

// Article represents an article that is needed for the assembly of the product.
type Article struct {
	ID      article.ID      `json:"art_id"`
	Barcode article.Barcode `json:"barcode"`
	Name    string          `json:"name"`
	Amount  int             `json:"amount_of"`
}

// ArticleRow is used while creating relationship between article and products in db.
type ArticleRow struct {
	ID        article.ID
	ProductID ID
	Amount    int
}
