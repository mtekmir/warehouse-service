package server

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

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

	if err := s.productService.Import(r.Context(), b.Products); err != nil {
		return errors.E(op, err)
	}

	w.WriteHeader(200)
	return nil
}

func (s *Server) handleGetProducts(w http.ResponseWriter, r *http.Request) error {
	var op errors.Op = "reqHandlers.handleGetProducts"

	ff := &product.Filters{}
	for _, b := range strings.Split(r.URL.Query().Get("barcodes"), ",") {
		if b != "" {
			*ff.BB = append(*ff.BB, product.Barcode(b))
		}
	}

	res, err := s.productService.FindAll(r.Context(), ff)
	if err != nil {
		return errors.E(op, err)
	}

	return json.NewEncoder(w).Encode(res)
}

func (s *Server) handleGetProduct(w http.ResponseWriter, r *http.Request) error {
	var op errors.Op = "reqHandlers.handleGetProduct"

	ID, err := strconv.Atoi(string(productPath.FindSubmatch([]byte(r.URL.Path))[1]))
	if err != nil {
		return errors.E(op, err)
	}

	p, err := s.productService.Find(r.Context(), product.ID(ID))
	if err != nil {
		return errors.E(op, err)
	}

	return json.NewEncoder(w).Encode(p)
}

func (s *Server) handleRemoveProduct(w http.ResponseWriter, r *http.Request) error {
	var op errors.Op = "reqHandlers.handleRemoveProduct"

	ID, err := strconv.Atoi(string(productPath.FindSubmatch([]byte(r.URL.Path))[1]))
	if err != nil {
		return errors.E(op, err)
	}

	body := struct {
		Qty int `json:"qty"`
	}{}

	json.NewDecoder(r.Body).Decode(&body)

	p, err := s.productService.Remove(r.Context(), product.ID(ID), body.Qty)
	if err != nil {
		return errors.E(op, err)
	}

	return json.NewEncoder(w).Encode(p)
}
