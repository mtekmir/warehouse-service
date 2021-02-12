package server

import (
	"encoding/json"
	"net/http"

	"github.com/mtekmir/warehouse-service/internal/article"
	"github.com/mtekmir/warehouse-service/internal/errors"
)

func (s *Server) handleImportArticles(w http.ResponseWriter, r *http.Request) error {
	var op errors.Op = "reqHandlers.handleImportArticles"

	b := struct {
		Inventory []*article.Article `json:"inventory"`
	}{}
	if err := json.NewDecoder(r.Body).Decode(&b); err != nil {
		return errors.E(op, errors.Invalid, "Unable to unmarshal json. Invalid format", err)
	}

	res, err := s.articleService.Import(b.Inventory)
	if err != nil {
		return errors.E(op, err)
	}

	return json.NewEncoder(w).Encode(res)
}
