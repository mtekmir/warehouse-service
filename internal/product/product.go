package product

import (
	"github.com/mtekmir/warehouse-service/internal/article"
)

// ID of a product.
type ID int

// Product represents a product that is made of a set of articles.
type Product struct {
	ID       ID              `json:"id"`
	Name     string          `json:"name"`
	Articles article.Article `json:"contain_articles"`
}
