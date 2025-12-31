package middleware

import (
	"backend/handler"
	"backend/model"
	"backend/service"
	"context"
	"encoding/json"
	"net/http"
	"strings"
)

// Auth creates an authentication middleware
func Auth(authService *service.AuthService) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get Authorization header
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				respondUnauthorized(w, "Missing authorization header")
				return
			}

			// Parse Bearer token
			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
				respondUnauthorized(w, "Invalid authorization header format")
				return
			}

			token := parts[1]

			// Validate token
			payload, err := authService.ValidateToken(token)
			if err != nil {
				respondUnauthorized(w, "Invalid or expired token")
				return
			}

			// Add user info to context
			ctx := context.WithValue(r.Context(), handler.UserContextKey, payload)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// OptionalAuth is like Auth but allows unauthenticated requests
func OptionalAuth(authService *service.AuthService) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				next.ServeHTTP(w, r)
				return
			}

			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
				next.ServeHTTP(w, r)
				return
			}

			token := parts[1]
			payload, err := authService.ValidateToken(token)
			if err != nil {
				next.ServeHTTP(w, r)
				return
			}

			ctx := context.WithValue(r.Context(), handler.UserContextKey, payload)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// RequireRole creates a middleware that requires a specific role
func RequireRole(roles ...model.Role) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			payload, ok := r.Context().Value(handler.UserContextKey).(*model.TokenPayload)
			if !ok || payload == nil {
				respondUnauthorized(w, "Unauthorized")
				return
			}

			for _, role := range roles {
				if payload.Role == role {
					next.ServeHTTP(w, r)
					return
				}
			}

			respondForbidden(w, "Insufficient permissions")
		})
	}
}

func respondUnauthorized(w http.ResponseWriter, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	json.NewEncoder(w).Encode(model.NewErrorResponse(http.StatusUnauthorized, message, ""))
}

func respondForbidden(w http.ResponseWriter, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusForbidden)
	json.NewEncoder(w).Encode(model.NewErrorResponse(http.StatusForbidden, message, ""))
}
