package server

import (
	"encoding/json"
	"net/http"

	"github.com/mtekmir/warehouse-service/internal/article"
	"github.com/mtekmir/warehouse-service/internal/errors"
)

type inv struct {
	Inventory []*article.Article `json:"inventory"`
}

func (s *Server) handleImportArticles(w http.ResponseWriter, r *http.Request) error {
	var op errors.Op = "reqHandlers.handleImportArticles"

	var b inv
	if err := json.NewDecoder(r.Body).Decode(&b); err != nil {
		return errors.E(op, errors.Invalid, "Unable to unmarshal json. Invalid format", err)
	}

	res, err := s.ArticleService.Import(r.Context(), b.Inventory)
	if err != nil {
		return errors.E(op, err)
	}

	return json.NewEncoder(w).Encode(res)
}

func (s *Server) handleGetArticles(w http.ResponseWriter, r *http.Request) error {
	var op errors.Op = "reqHandlers.handleGetArticles"

	var res inv
	arts, err := s.ArticleService.FindAll(r.Context())
	if err != nil {
		return errors.E(op, err)
	}

	res.Inventory = arts
	return json.NewEncoder(w).Encode(res)
}
