package test

import (
	"context"

	"github.com/mtekmir/warehouse-service/internal/article"
	"github.com/mtekmir/warehouse-service/internal/product"
)

// MockArticleService is mock impl of article service.
type MockArticleService struct {
	Calls map[string]interface{}
}

func (m *MockArticleService) Import(_ context.Context, rows []*article.Article) ([]*article.Article, error) {
	m.Calls["Import"] = rows
	return []*article.Article{}, nil
}

func (m *MockArticleService) FindAll(_ context.Context) ([]*article.Article, error) {
	m.Calls["FindAll"] = true
	return []*article.Article{}, nil
}

func NewMockArticleService() *MockArticleService {
	return &MockArticleService{
		Calls: make(map[string]interface{}),
	}
}

type MockProductService struct {
	Calls map[string][]interface{}
}

func (m *MockProductService) Import(ctx context.Context, rows []*product.Product) error {
	m.Calls["Import"] = []interface{}{rows}
	return nil
}

func (m *MockProductService) Remove(ctx context.Context, ID product.ID, qty int) (*product.StockInfo, error) {
	m.Calls["Remove"] = []interface{}{ID, qty}
	return &product.StockInfo{}, nil
}

func (m *MockProductService) Find(ctx context.Context, ID product.ID) (*product.StockInfo, error) {
	m.Calls["Find"] = []interface{}{ID}
	return &product.StockInfo{}, nil
}

func (m *MockProductService) FindAll(ctx context.Context, ff *product.Filters) ([]*product.StockInfo, error) {
	m.Calls["FindAll"] = []interface{}{ff}
	return []*product.StockInfo{}, nil
}

func NewMockProductService() *MockProductService {
	return &MockProductService{
		Calls: make(map[string][]interface{}),
	}
}
