package main

import (
	"backend/config"
	"backend/handler"
	"backend/middleware"
	"backend/model"
	"backend/service"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"
	_ "github.com/jackc/pgx/v5/stdlib"
)

func main() {
	// Load configuration
	cfg := config.Load()

	// Connect to database
	db, err := connectDB(cfg)
	if err != nil {
		log.Fatalf("❌ Failed to connect to database: %v", err)
	}
	defer db.Close()
	log.Println("✓ Connected to PostgreSQL database")

	// Initialize services with database
	authService := service.NewAuthService(cfg, db)
	userService := service.NewUserService(authService, db)
	queueService := service.NewQueueService(db)
	matchService := service.NewMatchService(userService, db)

	// Initialize handlers
	authHandler := handler.NewAuthHandler(authService)
	userHandler := handler.NewUserHandler(userService)
	queueHandler := handler.NewQueueHandler(queueService)
	matchHandler := handler.NewMatchHandler(matchService)

	// Initialize rate limiters
	authRateLimiter := middleware.StrictRateLimit()
	apiRateLimiter := middleware.DefaultRateLimit()

	// Create router
	r := chi.NewRouter()

	// Global middleware
	r.Use(chiMiddleware.Logger)
	r.Use(chiMiddleware.Recoverer)
	r.Use(chiMiddleware.RequestID)
	r.Use(chiMiddleware.RealIP)
	r.Use(middleware.NewCORS(cfg))

	// Health check
	r.Get("/api/health", func(w http.ResponseWriter, r *http.Request) {
		// Check DB health
		dbStatus := "ok"
		if err := db.Ping(); err != nil {
			dbStatus = "error"
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(model.NewSuccessResponse(map[string]interface{}{
			"status":   "ok",
			"version":  "1.0.0",
			"database": dbStatus,
			"time":     time.Now().Format(time.RFC3339),
		}))
	})

	// Public routes (with strict rate limiting)
	r.Group(func(r chi.Router) {
		r.Use(middleware.RateLimit(authRateLimiter))

		r.Post("/api/register", authHandler.Register)
		r.Post("/api/login", authHandler.Login)
		r.Post("/api/refresh", authHandler.Refresh)
		r.Post("/api/forgot-password", authHandler.ForgotPassword)
		r.Post("/api/reset-password", authHandler.ResetPassword)
	})

	// Protected routes
	r.Group(func(r chi.Router) {
		r.Use(middleware.RateLimit(apiRateLimiter))
		r.Use(middleware.Auth(authService))

		// Auth
		r.Post("/api/logout", authHandler.Logout)

		// Profile
		r.Get("/api/profile", userHandler.GetProfile)
		r.Put("/api/profile", userHandler.UpdateProfile)
		r.Get("/api/profile/stats", userHandler.GetStats)

		// Queue
		r.Get("/api/queue", queueHandler.GetStatus)
		r.Post("/api/queue/join", queueHandler.Join)
		r.Post("/api/queue/leave", queueHandler.Leave)

		// Matches
		r.Get("/api/matches", matchHandler.GetHistory)
		r.Get("/api/matches/active", matchHandler.GetActive)
	})

	// Organizer/Admin routes
	r.Group(func(r chi.Router) {
		r.Use(middleware.RateLimit(apiRateLimiter))
		r.Use(middleware.Auth(authService))
		r.Use(middleware.RequireRole(model.RoleOrganizer, model.RoleAdmin))

		r.Post("/api/queue/call", queueHandler.CallNext)
		r.Post("/api/matches", matchHandler.Create)
		r.Put("/api/matches/result", matchHandler.RecordResult)

		// Admin: Get all completed matches
		r.Get("/api/admin/matches/completed", func(w http.ResponseWriter, r *http.Request) {
			matches, err := matchService.GetAllCompleted(100)
			if err != nil {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(model.NewErrorResponse(500, "Failed to fetch matches", err.Error()))
				return
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(model.NewSuccessResponse(matches))
		})

		// Admin: Get all users
		r.Get("/api/admin/users", func(w http.ResponseWriter, r *http.Request) {
			users, err := authService.GetAllUsers()
			if err != nil {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(model.NewErrorResponse(500, "Failed to fetch users", err.Error()))
				return
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(model.NewSuccessResponse(users))
		})

		// Admin: Update player settings
		r.Put("/api/admin/users/{userID}", func(w http.ResponseWriter, r *http.Request) {
			userIDStr := chi.URLParam(r, "userID")
			userID, err := strconv.ParseInt(userIDStr, 10, 64)
			if err != nil {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusBadRequest)
				json.NewEncoder(w).Encode(model.NewErrorResponse(400, "Invalid user ID", ""))
				return
			}

			var req model.UpdatePlayerAdminRequest
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusBadRequest)
				json.NewEncoder(w).Encode(model.NewErrorResponse(400, "Invalid request body", err.Error()))
				return
			}

			if err := authService.UpdatePlayerAdmin(userID, req); err != nil {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusBadRequest)
				json.NewEncoder(w).Encode(model.NewErrorResponse(400, "Failed to update player", err.Error()))
				return
			}

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(model.NewSuccessResponse(map[string]string{"message": "Player updated successfully"}))
		})
	})

	// Queue status (optional auth for personalized info)
	r.Group(func(r chi.Router) {
		r.Use(middleware.OptionalAuth(authService))
		r.Get("/api/queue/status", queueHandler.GetStatus)
	})

	// Create server
	server := &http.Server{
		Addr:         ":" + cfg.Server.Port,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in goroutine
	go func() {
		log.Printf("🏸 SmashQueue server starting on port %s", cfg.Server.Port)
		log.Printf("📍 Health check: http://localhost:%s/api/health", cfg.Server.Port)

		var err error
		if cfg.Server.CertFile != "" && cfg.Server.KeyFile != "" {
			log.Println("🔒 Starting with HTTPS/TLS")
			err = server.ListenAndServeTLS(cfg.Server.CertFile, cfg.Server.KeyFile)
		} else {
			log.Println("⚠️  Starting without TLS (development mode)")
			err = server.ListenAndServe()
		}

		if err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited")
}

func connectDB(cfg *config.Config) (*sql.DB, error) {
	// Use PostgreSQL URI format instead
	connStr := fmt.Sprintf(
		"postgresql://%s:%s@%s:%s/%s?sslmode=%s",
		cfg.Database.User,
		cfg.Database.Password,
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.Name,
		cfg.Database.SSLMode,
	)

	db, err := sql.Open("pgx", connStr)
	if err != nil {
		return nil, err
	}

	// Configure connection pool
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(time.Hour)
	db.SetConnMaxIdleTime(30 * time.Minute)

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return nil, err
	}

	return db, nil
}
