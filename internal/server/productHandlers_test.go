package server_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mtekmir/warehouse-service/internal/product"
	"github.com/mtekmir/warehouse-service/internal/server"
	"github.com/mtekmir/warehouse-service/test"
	"github.com/sirupsen/logrus"
)

func TestImportProducts(t *testing.T) {
	pSvc := test.NewMockProductService()
	srv := server.Server{ProductService: pSvc, Log: logrus.New()}

	ts := httptest.NewServer(http.HandlerFunc(srv.Router))
	defer ts.Close()

	body := `
	{ 
		"products": [
			{
			"name": "big chair",
			"barcode": "820438363",
			"contain_articles": [{"art_id": "1", "name": "big door", "amount_of": "433"	}]
			}
		] 
	}
	`
	res := testRequest(t, ts, "POST", "/products/import", body, []reqHeader{})
	if res.StatusCode != http.StatusOK {
		t.Errorf("Expected OK got %s", res.Status)
	}

	expectedB := []*product.Product{{Barcode: "820438363", Name: "big chair", Articles: []*product.Article{
		{ArtID: "1", Name: "big door", Amount: 433},
	}}}
	test.Compare(t, "importCallArgs", expectedB, pSvc.Calls["Import"][0])
}
