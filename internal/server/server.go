package server

import (
	"log"
	"net/http"

	"github.com/mtekmir/warehouse-service/internal/article"
	"github.com/mtekmir/warehouse-service/internal/product"
)

// Server is an abstraction that holds the dependencies for the http server
// and handles routing.
type Server struct {
	productService *product.Service
	articleService *article.Service
}

const (
	importProductsPath = "/products/import"
	importArticlesPath = "/articles/import"
)

func (s *Server) router(w http.ResponseWriter, r *http.Request) {
	switch {

	case r.Method == http.MethodPost && r.URL.Path == importProductsPath:
		handler(s.handleImportProducts).ServeHTTP(w, r)

	case r.Method == http.MethodPost && r.URL.Path == importArticlesPath:
		handler(s.handleImportArticles).ServeHTTP(w, r)

	}
}

// Start starts the server. Server sets up the routes and starts listening.
func (s *Server) Start() error {
	http.Handle("/", applyMiddlewares(http.HandlerFunc(s.router), noPanicMiddleware, corsMiddleware("*")))

	return http.ListenAndServe(":8080", nil)
}

// NewServer returns a new server instance with required dependencies.
func NewServer(ps *product.Service, as *article.Service) *Server {
	return &Server{
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

func (h handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	if err := h(w, r); err != nil {
		e, ok := err.(Error)
		if !ok {
			log.Printf("An unexpected error occurred: %v\n", err)
			w.WriteHeader(500)
			w.Write([]byte(`{"message": "Something went wrong."}`))
			return
		}
		log.Print(e.Error())
		w.WriteHeader(e.Code())
		w.Write(e.Body())
	}
}
