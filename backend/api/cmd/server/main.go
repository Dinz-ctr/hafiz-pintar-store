package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	dbsqlc "hafiz-pintar-store/backend/api/internal/database/sqlc"
	"hafiz-pintar-store/backend/api/internal/handler"
	"hafiz-pintar-store/backend/api/internal/repository"
	"hafiz-pintar-store/backend/api/internal/service"

	"github.com/jackc/pgx/v5/pgxpool"
)

type APIResponse struct {
	Status  string `json:"status"`
	Service string `json:"service"`
	Message string `json:"message"`
}

func writeJSON(w http.ResponseWriter, statusCode int, response APIResponse) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("JSON response error: %v", err)
	}
}

func main() {
	// 1. Application configuration.
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		log.Fatal("DATABASE_URL environment variable is required")
	}

	// 2. PostgreSQL configuration.
	config, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		log.Fatalf("Invalid database configuration: %v", err)
	}

	config.MaxConns = 10
	config.MinConns = 1
	config.MaxConnLifetime = time.Hour
	config.MaxConnIdleTime = 30 * time.Minute

	// 3. Initialize PostgreSQL connection pool.
	startupCtx, startupCancel := context.WithTimeout(
		context.Background(),
		10*time.Second,
	)
	defer startupCancel()

	pool, err := pgxpool.NewWithConfig(startupCtx, config)
	if err != nil {
		log.Fatalf("Unable to create database pool: %v", err)
	}
	defer pool.Close()

	if err := pool.Ping(startupCtx); err != nil {
		log.Fatalf("PostgreSQL connection failed: %v", err)
	}

	log.Println("PostgreSQL connection established successfully")

	// 4. Initialize SQLC.
	queries := dbsqlc.New(pool)
	log.Println("SQLC database queries initialized successfully")

	// 5. Initialize product layers.
	productRepo := repository.NewProductRepository(queries, pool)
	productService := service.NewProductService(productRepo)
	productHandler := handler.NewProductHandler(productService)

	log.Println("Product layers initialized successfully")

	// 6. Initialize HTTP router.
	mux := http.NewServeMux()

	// Liveness endpoint.
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, APIResponse{
			Status:  "ok",
			Service: "hafiz-pintar-api",
			Message: "API is running",
		})
	})

	// Readiness endpoint.
	mux.HandleFunc("GET /ready", func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
		defer cancel()

		_, err := queries.HealthCheck(ctx)
		if err != nil {
			log.Printf("Database readiness check failed: %v", err)

			writeJSON(w, http.StatusServiceUnavailable, APIResponse{
				Status:  "error",
				Service: "hafiz-pintar-api",
				Message: "Database is not available",
			})
			return
		}

		writeJSON(w, http.StatusOK, APIResponse{
			Status:  "ok",
			Service: "hafiz-pintar-api",
			Message: "API and PostgreSQL are ready",
		})
	})

	// Product endpoints.
	mux.HandleFunc("GET /api/v1/products", productHandler.List)
	mux.HandleFunc("GET /api/v1/products/{slug}", productHandler.GetBySlug)
	mux.HandleFunc("POST /api/v1/products", productHandler.Create)
	mux.HandleFunc("GET /api/v1/products/{slug}/variants", productHandler.GetVariantsBySlug)

	log.Println("Product routes registered successfully")

	// 7. Initialize HTTP server.
	server := &http.Server{
		Addr:              ":" + port,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	// 8. Start HTTP server.
	go func() {
		log.Printf("Hafiz Pintar API running on port %s", port)

		if err := server.ListenAndServe(); err != nil &&
			err != http.ErrServerClosed {
			log.Printf("HTTP server error: %v", err)
		}
	}()

	// 9. Graceful shutdown.
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	log.Println("Shutting down API server...")

	shutdownCtx, shutdownCancel := context.WithTimeout(
		context.Background(),
		10*time.Second,
	)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("HTTP server shutdown error: %v", err)
	}

	log.Println("API server stopped")
}
