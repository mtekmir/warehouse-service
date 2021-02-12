package server

import (
	"encoding/json"
	"net/http"

	"github.com/mtekmir/warehouse-service/internal/errors"
	"github.com/mtekmir/warehouse-service/internal/product"
)

func (s *Server) handleImportProducts(w http.ResponseWriter, r *http.Request) error {
	var op errors.Op = "reqHandlers.handleImportProducts"

	b := struct {
		Products []*product.Product `json:"products"`
	}{}

	if err := json.NewDecoder(r.Body).Decode(&b); err != nil {
		return errors.E(op, err)
	}

	if err := s.productService.Import(b.Products); err != nil {
		return errors.E(op, err)
	}

	w.WriteHeader(200)
	return nil
}
