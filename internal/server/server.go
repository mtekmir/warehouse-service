package server

import (
	"fmt"
	"net/http"
	"regexp"
	"time"

	"github.com/mtekmir/warehouse-service/internal/article"
	"github.com/mtekmir/warehouse-service/internal/product"
	"github.com/sirupsen/logrus"
)

// Server is an abstraction that holds the dependencies for the http server
// and handles routing.
type Server struct {
	productService *product.Service
	articleService *article.Service
	log            *logrus.Logger
}

var productPath = regexp.MustCompile(`/products/([0-9]+)`)

const (
	importProductsPath = "/products/import"
	getProductsPath    = "/products"
	importArticlesPath = "/articles/import"
)

func (s *Server) router(w http.ResponseWriter, r *http.Request) {
	switch {

	case r.Method == http.MethodGet && r.URL.Path == getProductsPath:
		handler(s.handleGetProducts).ServeHTTP(s.log, w, r)

	case r.Method == http.MethodGet && productPath.MatchString(r.URL.Path):
		handler(s.handleGetProduct).ServeHTTP(s.log, w, r)

	case r.Method == http.MethodPost && productPath.MatchString(r.URL.Path):
		handler(s.handleRemoveProduct).ServeHTTP(s.log, w, r)

	case r.Method == http.MethodPost && r.URL.Path == importProductsPath:
		handler(s.handleImportProducts).ServeHTTP(s.log, w, r)

	case r.Method == http.MethodPost && r.URL.Path == importArticlesPath:
		handler(s.handleImportArticles).ServeHTTP(s.log, w, r)

	}
}

// Start starts the server. Server sets up the routes and starts listening.
func (s *Server) Start(port string, wTimeout, rTimeout, idleTimeout time.Duration) error {
	http.Handle("/", applyMiddlewares(http.HandlerFunc(s.router), noPanicMiddleware(s.log), corsMiddleware("*")))

	srv := http.Server{
		Addr:         fmt.Sprintf(":%s", port),
		WriteTimeout: wTimeout,
		ReadTimeout:  rTimeout,
		IdleTimeout:  idleTimeout,
	}

	return srv.ListenAndServe()
}

// NewServer returns a new server instance with required dependencies.
func NewServer(l *logrus.Logger, ps *product.Service, as *article.Service) *Server {
	return &Server{
		log:            l,
		productService: ps,
		articleService: as,
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
