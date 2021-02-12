package postgres_test

import (
	"fmt"
	"testing"

	"github.com/mtekmir/warehouse-service/internal/article"
	"github.com/mtekmir/warehouse-service/internal/postgres"
	"github.com/mtekmir/warehouse-service/test"
)

func TestBatchInsert(t *testing.T) {
	db, dbTidy := test.SetupDB(t)
	defer dbTidy()

	test.CreateArticleTable(t, db)
	r := postgres.NewArticleRepo()

	aa := createArticles(3)

	arts, err := r.BatchInsert(db, aa)
	if err != nil {
		t.Errorf("Unable to batch insert articles. %v", err)
	}

	expectedArts := []*article.Article{
		{ID: 1, Name: "Name_1", Barcode: "Barcode_1", Stock: 1},
		{ID: 2, Name: "Name_2", Barcode: "Barcode_2", Stock: 2},
		{ID: 3, Name: "Name_3", Barcode: "Barcode_3", Stock: 3},
	}

	test.CompareArticleSlices(t, expectedArts, arts)
}

func TestFindAllByBarcode(t *testing.T) {
	db, dbTidy := test.SetupDB(t)
	defer dbTidy()

	test.CreateArticleTable(t, db)
	r := postgres.NewArticleRepo()

	aa := createArticles(6)

	_, err := r.BatchInsert(db, aa)
	if err != nil {
		t.Errorf("Unable to batch insert articles. %v", err)
	}

	found, err := r.FindAllByBarcode(db, &[]article.Barcode{"Barcode_1", "Barcode_2", "Barcode_3"})
	if err != nil {
		t.Errorf("Unable to find articles. %v", err)
	}

	expectedArts := []*article.Article{
		{ID: 1, Name: "Name_1", Barcode: "Barcode_1", Stock: 1},
		{ID: 2, Name: "Name_2", Barcode: "Barcode_2", Stock: 2},
		{ID: 3, Name: "Name_3", Barcode: "Barcode_3", Stock: 3},
	}

	test.CompareArticleSlices(t, expectedArts, found)
}

func TestAdjustQuantities(t *testing.T) {
	db, dbTidy := test.SetupDB(t)
	defer dbTidy()

	test.CreateArticleTable(t, db)
	r := postgres.NewArticleRepo()

	aa := createArticles(3)

	_, err := r.BatchInsert(db, aa)
	if err != nil {
		t.Errorf("Unable to batch insert articles. %v", err)
	}

	err = r.AdjustQuantities(db, article.QtyAdjustmentAdd, []*article.QtyAdjustment{{ID: 1, Qty: 10}, {ID: 2, Qty: 20}})
	if err != nil {
		t.Errorf("Unable to adjust quantities of articles. %v", err)
	}

	found, err := r.FindAllByBarcode(db, nil)
	if err != nil {
		t.Errorf("Unable to find articles. %v", err)
	}

	expectedArts := []*article.Article{
		{ID: 1, Name: "Name_1", Barcode: "Barcode_1", Stock: 11},
		{ID: 2, Name: "Name_2", Barcode: "Barcode_2", Stock: 22},
		{ID: 3, Name: "Name_3", Barcode: "Barcode_3", Stock: 3},
	}

	test.CompareArticleSlices(t, expectedArts, found)

	err = r.AdjustQuantities(db, article.QtyAdjustmentSubtract, []*article.QtyAdjustment{{ID: 2, Qty: 2}, {ID: 3, Qty: 3}})
	if err != nil {
		t.Errorf("Unable to adjust quantities of articles. %v", err)
	}

	found, err = r.FindAllByBarcode(db, nil)
	if err != nil {
		t.Errorf("Unable to find articles. %v", err)
	}

	expectedArts = []*article.Article{
		{ID: 1, Name: "Name_1", Barcode: "Barcode_1", Stock: 11},
		{ID: 2, Name: "Name_2", Barcode: "Barcode_2", Stock: 20},
		{ID: 3, Name: "Name_3", Barcode: "Barcode_3", Stock: 0},
	}

	test.CompareArticleSlices(t, expectedArts, found)

	err = r.AdjustQuantities(db, article.QtyAdjustmentReplace, []*article.QtyAdjustment{{ID: 1, Qty: 5}, {ID: 2, Qty: 10}})
	if err != nil {
		t.Errorf("Unable to adjust quantities of articles. %v", err)
	}

	found, err = r.FindAllByBarcode(db, nil)
	if err != nil {
		t.Errorf("Unable to find articles. %v", err)
	}

	expectedArts = []*article.Article{
		{ID: 1, Name: "Name_1", Barcode: "Barcode_1", Stock: 5},
		{ID: 2, Name: "Name_2", Barcode: "Barcode_2", Stock: 10},
		{ID: 3, Name: "Name_3", Barcode: "Barcode_3", Stock: 0},
	}

	test.CompareArticleSlices(t, expectedArts, found)
}

func TestImport(t *testing.T) {
	db, dbTidy := test.SetupDB(t)
	defer dbTidy()

	test.CreateArticleTable(t, db)
	r := postgres.NewArticleRepo()

	aa := createArticles(3)

	imported, err := r.Import(db, aa)
	if err != nil {
		t.Errorf("Unable to import articles. %v", err)
	}

	expectedArts := []*article.Article{
		{ID: 1, Name: "Name_1", Barcode: "Barcode_1", Stock: 1},
		{ID: 2, Name: "Name_2", Barcode: "Barcode_2", Stock: 2},
		{ID: 3, Name: "Name_3", Barcode: "Barcode_3", Stock: 3},
	}

	test.CompareArticleSlices(t, expectedArts, imported)

	imported, err = r.Import(db, aa)
	if err != nil {
		t.Errorf("Unable to import articles. %v", err)
	}

	expectedArts = []*article.Article{
		{ID: 1, Name: "Name_1", Barcode: "Barcode_1", Stock: 2},
		{ID: 2, Name: "Name_2", Barcode: "Barcode_2", Stock: 4},
		{ID: 3, Name: "Name_3", Barcode: "Barcode_3", Stock: 6},
	}

	test.CompareArticleSlices(t, expectedArts, imported)

	found, err := r.FindAllByBarcode(db, nil)
	if err != nil {
		t.Errorf("Unable to find all articles. %v", err)
	}

	test.CompareArticleSlices(t, expectedArts, found)
}

func createArticles(n int) []*article.Article {
	aa := make([]*article.Article, 0, n)
	for i := 0; i < n; i++ {
		aa = append(aa, &article.Article{
			Name:    fmt.Sprintf("Name_%d", i+1),
			Barcode: article.Barcode(fmt.Sprintf("Barcode_%d", i+1)),
			Stock:   i + 1,
		})
	}
	return aa
}
