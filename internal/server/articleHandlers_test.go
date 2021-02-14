package server_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mtekmir/warehouse-service/internal/article"
	"github.com/mtekmir/warehouse-service/internal/server"
	"github.com/mtekmir/warehouse-service/test"
)

func TestArticleRoutes(t *testing.T) {
	aSvc := test.NewMockArticleService()
	srv := server.Server{ArticleService: aSvc}

	ts := httptest.NewServer(http.HandlerFunc(srv.Router))
	defer ts.Close()

	res := testRequest(t, ts, "GET", "/articles", nil, []reqHeader{})
	if res.StatusCode != http.StatusOK {
		t.Errorf("Expected OK got %s", res.Status)
	}

	if _, ok := aSvc.Calls["FindAll"]; !ok {
		t.Error("Expected articleService.findall to be called")
	}

	body := `{ "inventory": [{"art_id": "19999", "name": "rear leg", "stock": "281"}] }`
	res2 := testRequest(t, ts, "POST", "/articles/import", body, []reqHeader{})
	if res2.StatusCode != http.StatusOK {
		t.Errorf("Expected OK got %s", res2.Status)
	}

	expectedB := []*article.Article{{ArtID: "19999", Name: "rear leg", Stock: 281}}
	test.Compare(t, "importCallArgs", expectedB, aSvc.Calls["Import"])
}
