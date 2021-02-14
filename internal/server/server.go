package server

import (
	"context"
	"fmt"
	"net/http"
	"regexp"
	"time"

	"github.com/mtekmir/warehouse-service/internal/article"
	"github.com/mtekmir/warehouse-service/internal/product"
	"github.com/sirupsen/logrus"
)

type productService interface {
	Import(ctx context.Context, rows []*product.Product) error
	Remove(ctx context.Context, ID product.ID, qty int) (*product.StockInfo, error)
	Find(ctx context.Context, ID product.ID) (*product.StockInfo, error)
	FindAll(ctx context.Context, ff *product.Filters) ([]*product.StockInfo, error)
}

type articleService interface {
	Import(ctx context.Context, rows []*article.Article) ([]*article.Article, error)
	FindAll(ctx context.Context) ([]*article.Article, error)
}

// Server is an abstraction that holds the dependencies for the http server
// and handles routing.
type Server struct {
	ProductService productService
	ArticleService articleService
	Log            *logrus.Logger
}

var productPath = regexp.MustCompile(`/products/([0-9]+)`)

const (
	importProductsPath = "/products/import"
	getProductsPath    = "/products"
	importArticlesPath = "/articles/import"
	getArticlesPath    = "/articles"
)

// Router is a request multiplexer.
func (s *Server) Router(w http.ResponseWriter, r *http.Request) {
	switch {

	case r.Method == http.MethodGet && r.URL.Path == getProductsPath:
		handler(s.handleGetProducts).ServeHTTP(s.Log, w, r)

	case r.Method == http.MethodGet && productPath.MatchString(r.URL.Path):
		handler(s.handleGetProduct).ServeHTTP(s.Log, w, r)

	case r.Method == http.MethodPost && productPath.MatchString(r.URL.Path):
		handler(s.handleRemoveProduct).ServeHTTP(s.Log, w, r)

	case r.Method == http.MethodPost && r.URL.Path == importProductsPath:
		handler(s.handleImportProducts).ServeHTTP(s.Log, w, r)

	case r.Method == http.MethodPost && r.URL.Path == importArticlesPath:
		handler(s.handleImportArticles).ServeHTTP(s.Log, w, r)

	case r.Method == http.MethodGet && r.URL.Path == getArticlesPath:
		handler(s.handleGetArticles).ServeHTTP(s.Log, w, r)

	}
}

// Start starts the server. Server sets up the routes and starts listening.
func (s *Server) Start(port string, wTimeout, rTimeout, idleTimeout time.Duration) error {
	http.Handle("/", applyMiddlewares(http.HandlerFunc(s.Router), noPanicMiddleware(s.Log), corsMiddleware("*")))

	srv := http.Server{
		Addr:         fmt.Sprintf(":%s", port),
		WriteTimeout: wTimeout,
		ReadTimeout:  rTimeout,
		IdleTimeout:  idleTimeout,
	}

	return srv.ListenAndServe()
}

// NewServer returns a new server instance with required dependencies.
func NewServer(l *logrus.Logger, ps productService, as articleService) *Server {
	return &Server{
		Log:            l,
		ProductService: ps,
		ArticleService: as,
	}
}

type handler func(w http.ResponseWriter, r *http.Request) error

// Error describes an error that will be sent to the user.
type Error interface {
	Error() string
	Code() int
	Body() []byte
}

func (h handler) ServeHTTP(l *logrus.Logger, w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	if err := h(w, r); err != nil {
		e, ok := err.(Error)
		if !ok {
			l.Printf("An unexpected error occurred: %v\n", err)
			w.WriteHeader(500)
			w.Write([]byte(`{"message": "Something went wrong."}`))
			return
		}
		l.Print(e.Error())
		w.WriteHeader(e.Code())
		w.Write(e.Body())
	}
}
