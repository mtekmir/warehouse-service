package article

import (
	"encoding/json"
	"strconv"

	"github.com/mtekmir/warehouse-service/internal/errors"
)

// ID is the internal ID of an article.
type ID int

// ArtID is the external ID of an article.
type ArtID string

// Article represents a part of a product.
type Article struct {
	ID    ID     `json:"-"`
	ArtID ArtID  `json:"art_id"`
	Name  string `json:"name"`
	Stock int    `json:"stock"`
}

// UnmarshalJSON implements json.Unmarshaler.
func (a *Article) UnmarshalJSON(data []byte) error {
	var op errors.Op = "article.unmarshalJSON"

	type Alias Article
	j := &struct {
		Stock string `json:"stock"`
		*Alias
	}{
		Alias: (*Alias)(a),
	}

	if err := json.Unmarshal(data, &j); err != nil {
		return errors.E(op, err)
	}

	s, err := strconv.Atoi(j.Stock)
	if err != nil {
		return errors.E(op, err)
	}

	a.Stock = s

	return nil
}

// QtyAdjustmentKind is the type of qty adjustment of articleRepo.AdjustQuantities method
type QtyAdjustmentKind int

// QtyAdjustmentType is the type of qty adjustment of articleRepo.AdjustQuantities method
const (
	QtyAdjustmentAdd QtyAdjustmentKind = iota
	QtyAdjustmentSubtract
	QtyAdjustmentReplace
)

// QtyAdjustment describes a qty adjustment for an article.
type QtyAdjustment struct {
	ID  ID
	Qty int
}
