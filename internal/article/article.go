package article

// ID of an article
type ID int

// Article represents a part of a product.
type Article struct {
	ID    ID     `json:"art_id"`
	Name  string `json:"name"`
	Stock int    `json:"stock"`
}
