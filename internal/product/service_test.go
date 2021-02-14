package product_test

import (
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/mtekmir/warehouse-service/internal/article"
	"github.com/mtekmir/warehouse-service/internal/postgres"
	"github.com/mtekmir/warehouse-service/internal/product"
	"github.com/mtekmir/warehouse-service/test"
)

func TestImportProducts(t *testing.T) {
	db, dbTidy := test.SetupDB(t)
	defer dbTidy()

	test.CreateProductTables(t, db)

	ar := postgres.NewArticleRepo()
	pr := postgres.NewProductRepo()
	s := product.NewService(db, pr, ar)
	pp := createArticles(2)

	err := s.Import(pp)
	if err != nil {
		t.Errorf("Unable to import products. %v", err)
	}

	expectedArts := []*article.Article{
		{ID: 1, Name: "Article_1_1", ArtID: "Art_ArtID_1_1", Stock: 5},
		{ID: 2, Name: "Article_1_2", ArtID: "Art_ArtID_1_2", Stock: 5},
	}

	foundArts, err := ar.FindAll(db, nil)
	if err != nil {
		t.Errorf("Unable to find all articles. %v", err)
	}

	test.Compare(t, "article", expectedArts, foundArts)

	expectedPP := []*product.StockInfo{
		{ID: 1, Name: "Name_1", Barcode: "Barcode_1", AvailableQty: 1, Articles: []*product.ArticleStock{
			{ID: 1, Name: "Article_1_1", ArtID: "Art_ArtID_1_1", Stock: 5, RequiredAmount: 5},
		}},
		{ID: 2, Name: "Name_2", Barcode: "Barcode_2", AvailableQty: 1, Articles: []*product.ArticleStock{
			{ID: 2, Name: "Article_1_2", ArtID: "Art_ArtID_1_2", Stock: 5, RequiredAmount: 5},
		}},
	}

	foundPP, err := pr.FindAll(db, &product.Filters{})
	if err != nil {
		t.Errorf("Unable to find products. %v", err)
	}

	test.Compare(t, "product", expectedPP, foundPP)
}

func TestRemove(t *testing.T) {
	db, dbTidy := test.SetupDB(t)
	defer dbTidy()

	test.CreateProductTables(t, db)

	ar := postgres.NewArticleRepo()
	pr := postgres.NewProductRepo()
	s := product.NewService(db, pr, ar)

	prod := &product.Product{Barcode: "barcode", Name: "name", Articles: []*product.Article{
		{ArtID: "art_id1", Name: "name_1", Amount: 5},
		{ArtID: "art_id2", Name: "name_2", Amount: 3},
		{ArtID: "art_id3", Name: "name_3", Amount: 2},
	}}

	if err := s.Import([]*product.Product{prod}); err != nil {
		t.Errorf("Unable to import products. %v", err)
	}

	foundP, err := s.Find(1)
	if err != nil {
		t.Errorf("Unable to find products. %v", err)
	}

	expectedStockInfo := &product.StockInfo{
		ID: 1, Barcode: "barcode", Name: "name", AvailableQty: 1, Articles: []*product.ArticleStock{
			{ArtID: "art_id1", Name: "name_1", Stock: 5, RequiredAmount: 5},
			{ArtID: "art_id2", Name: "name_2", Stock: 3, RequiredAmount: 3},
			{ArtID: "art_id3", Name: "name_3", Stock: 2, RequiredAmount: 2},
		},
	}

	test.Compare(t, "stockInfo", expectedStockInfo, foundP, cmpopts.IgnoreFields(product.ArticleStock{}, "ID"))

	if _, err := s.Remove(1, 2); err == nil {
		t.Errorf("Should return an error when there is not enough stock")
	}

	stock, err := s.Remove(1, 1)
	if err != nil {
		t.Errorf("Unable to remove a product. %v", err)
	}

	expectedStockInfo = &product.StockInfo{
		ID: 1, Barcode: "barcode", Name: "name", AvailableQty: 0, Articles: []*product.ArticleStock{
			{ArtID: "art_id1", Name: "name_1", Stock: 0, RequiredAmount: 5},
			{ArtID: "art_id2", Name: "name_2", Stock: 0, RequiredAmount: 3},
			{ArtID: "art_id3", Name: "name_3", Stock: 0, RequiredAmount: 2},
		},
	}

	test.Compare(t, "stockInfo", expectedStockInfo, stock, cmpopts.IgnoreFields(product.ArticleStock{}, "ID"))

	foundP, err = s.Find(1)
	if err != nil {
		t.Errorf("Unable to find products. %v", err)
	}

	test.Compare(t, "stockInfo", expectedStockInfo, foundP, cmpopts.IgnoreFields(product.ArticleStock{}, "ID"))
}

func compareStockInfos(t *testing.T, expected, got *product.StockInfo) {
	test.Compare(t, "stockInfo", expected, got, cmpopts.IgnoreFields(product.ArticleStock{}, "ID"), cmpopts.SortSlices(func(s1, s2 *product.ArticleStock) bool {
		return s1.ArtID > s2.ArtID
	}))
}

func createArticles(n int) []*product.Product {
	aa := make([]*product.Product, 0, n)
	for i := 0; i < n; i++ {
		aa = append(aa, &product.Product{
			Name:    fmt.Sprintf("Name_%d", i+1),
			Barcode: product.Barcode(fmt.Sprintf("Barcode_%d", i+1)),
			Articles: []*product.Article{
				{Name: fmt.Sprintf("Article_1_%d", i+1), ArtID: article.ArtID(fmt.Sprintf("Art_ArtID_1_%d", i+1)), Amount: 5},
			},
		})
	}
	return aa
}
