package middleware

import (
	"net/http"

	"backend/config"

	"github.com/go-chi/cors"
)

// NewCORS creates a CORS middleware with the specified allowed origins
func NewCORS(cfg *config.Config) func(next http.Handler) http.Handler {
	return cors.Handler(cors.Options{
		AllowedOrigins:   cfg.CORS.AllowedOrigins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300, // Maximum value not ignored by any major browser
	})
}
