package product_test

import (
	"fmt"
	"testing"

	"github.com/mtekmir/warehouse-service/internal/article"
	"github.com/mtekmir/warehouse-service/internal/postgres"
	"github.com/mtekmir/warehouse-service/internal/product"
	"github.com/mtekmir/warehouse-service/test"
)

func TestImport(t *testing.T) {
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
		{ID: 1, Name: "Article_1_1", Barcode: "Art_Barcode_1_1", Stock: 5},
		{ID: 2, Name: "Article_1_2", Barcode: "Art_Barcode_1_2", Stock: 5},
	}

	foundArts, err := ar.FindAllByBarcode(db, nil)
	if err != nil {
		t.Errorf("Unable to find all articles. %v", err)
	}

	test.CompareArticleSlices(t, expectedArts, foundArts)

	expectedPP := []*product.Product{
		{ID: 1, Name: "Name_1", Barcode: "Barcode_1", Articles: []*product.Article{
			{ID: 1, Name: "Article_1_1", Barcode: "Art_Barcode_1_1", Amount: 5},
		}},
		{ID: 2, Name: "Name_2", Barcode: "Barcode_2", Articles: []*product.Article{
			{ID: 2, Name: "Article_1_2", Barcode: "Art_Barcode_1_2", Amount: 5},
		}},	
	}

	foundPP, err := pr.FindAll(db, nil)
	if err != nil {
		t.Errorf("Unable to find products. %v", err)
	}

	test.Compare(t, "product", expectedPP, foundPP)
}

func createArticles(n int) []*product.Product {
	aa := make([]*product.Product, 0, n)
	for i := 0; i < n; i++ {
		aa = append(aa, &product.Product{
			Name:    fmt.Sprintf("Name_%d", i+1),
			Barcode: product.Barcode(fmt.Sprintf("Barcode_%d", i+1)),
			Articles: []*product.Article{
				{Name: fmt.Sprintf("Article_1_%d", i+1), Barcode: article.Barcode(fmt.Sprintf("Art_Barcode_1_%d", i+1)), Amount: 5},
			},
		})
	}
	return aa
}
