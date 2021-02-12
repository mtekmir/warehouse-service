package article

// ID is the internal ID of an article.
type ID int

// Barcode is the type of the article.
type Barcode string

// Article represents a part of a product.
type Article struct {
	ID      ID      `json:"art_id"`
	Barcode Barcode `json:"barcode"`
	Name    string  `json:"name"`
	Stock   int     `json:"stock"`
}

// QtyAdjustmentKind is the type of qty adjustment of productRepo.AdjustQuantities method
type QtyAdjustmentKind int

// QtyAdjustmentType is the type of qty adjustment of productRepo.AdjustQuantities method
const (
	QtyAdjustmentAdd QtyAdjustmentKind = iota
	QtyAdjustmentSubtract
	QtyAdjustmentReplace
)

// QtyAdjustment describes a qty adjustment for a product.
type QtyAdjustment struct {
	ID  ID
	Qty int
}

// ImportSummary is the response type of article import.
type ImportSummary struct {
	Updated []*Article `json:"updated"`
	Created []*Article `json:"created"`
}
