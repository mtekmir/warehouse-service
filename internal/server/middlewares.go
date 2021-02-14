package server

import (
	"net/http"
	"strings"

	"github.com/sirupsen/logrus"
)

func corsMiddleware(clientURL string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			allowedOrigin := ""
			if strings.TrimSpace(clientURL) == r.Header.Get("Origin") {
				allowedOrigin = r.Header.Get("Origin")
			}

			w.Header().Set("Access-Control-Allow-Origin", allowedOrigin)
			w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}
			next.ServeHTTP(w, r)
			return
		})
	}
}

func noPanicMiddleware(log *logrus.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				err := recover()
				if err != nil {
					log.Printf("An unexpected error occurred: %v\n", err)
					w.WriteHeader(500)
					w.Write([]byte(`{"message": "Something went wrong."}`))
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}

func applyMiddlewares(h http.Handler, mm ...func(http.Handler) http.Handler) http.Handler {
	for i := len(mm) - 1; i >= 0; i-- {
		h = mm[i](h)
	}
	return h
}
