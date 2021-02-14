package product

import (
	"encoding/json"
	"strconv"

	"github.com/mtekmir/warehouse-service/internal/article"
	"github.com/mtekmir/warehouse-service/internal/errors"
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

// UnmarshalJSON implements json.Unmarshaler.
func (prd *Product) UnmarshalJSON(data []byte) error {
	type Alias Product
	p := struct {
		*Alias
	}{Alias: (*Alias)(prd)}
	var op errors.Op = "product.unmarshalJSON"

	if err := json.Unmarshal(data, &p); err != nil {
		return errors.E(op, err)
	}

	if p.Barcode == "" {
		return errors.E(op, "Barcode must not be empty", errors.Invalid)
	}

	if p.Name == "" {
		return errors.E(op, "Product name must not be empty", errors.Invalid)
	}

	if len(p.Articles) == 0 {
		return errors.E(op, "Product must contain at least one article", errors.Invalid)
	}

	return nil
}

// Article represents an article that is needed for the assembly of the product.
type Article struct {
	ID     article.ID    `json:"-"`
	ArtID  article.ArtID `json:"art_id"`
	Name   string        `json:"name"`
	Amount int           `json:"amount_of"`
}

// UnmarshalJSON implements json.Unmarshaler.
func (a *Article) UnmarshalJSON(data []byte) error {
	var op errors.Op = "article.unmarshalJSON"

	type Alias Article
	j := &struct {
		ReqAmount string `json:"amount_of"`
		*Alias
	}{
		Alias: (*Alias)(a),
	}

	if err := json.Unmarshal(data, &j); err != nil {
		return errors.E(op, err)
	}

	s, err := strconv.Atoi(j.ReqAmount)
	if err != nil {
		return errors.E(op, errors.Invalid, err)
	}

	if s == 0 {
		return errors.E(op, errors.Invalid, "Amount must be bigger than 0")
	}

	a.Amount = s

	return nil
}

// ArticleRow is used while creating relationship between article and products in db.
type ArticleRow struct {
	ID        article.ID
	ProductID ID
	Amount    int
}

// ArticleStock conveys the stock information of an article that is required to assemble a product.
type ArticleStock struct {
	ID             article.ID    `json:"-"`
	ArtID          article.ArtID `json:"art_id"`
	Name           string        `json:"name"`
	Stock          int           `json:"stock"`
	RequiredAmount int           `json:"reqired_amount"`
}

// StockInfo conveys the stock information of a product and the required parts.
type StockInfo struct {
	ID           ID              `json:"id"`
	Barcode      Barcode         `json:"barcode"`
	Name         string          `json:"name"`
	AvailableQty int             `json:"available_quantity"`
	Articles     []*ArticleStock `json:"contain_articles"`
}
