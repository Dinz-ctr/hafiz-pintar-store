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

func writeJSON(
	w http.ResponseWriter,
	statusCode int,
	response APIResponse,
) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("JSON response error: %v", err)
	}
}

func main() {
	// =========================================
	// 1. APPLICATION CONFIGURATION
	// =========================================

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		log.Fatal("DATABASE_URL environment variable is required")
	}

	// =========================================
	// 2. POSTGRESQL CONFIGURATION
	// =========================================

	config, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		log.Fatalf("Invalid database configuration: %v", err)
	}

	config.MaxConns = 10
	config.MinConns = 1
	config.MaxConnLifetime = time.Hour
	config.MaxConnIdleTime = 30 * time.Minute

	// =========================================
	// 3. INITIALIZE DATABASE CONNECTION
	// =========================================

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

	// =========================================
	// 4. INITIALIZE SQLC
	// =========================================

	queries := dbsqlc.New(pool)

	log.Println("SQLC database queries initialized successfully")

	// =========================================
	// 5. INITIALIZE PRODUCT LAYERS
	// =========================================

	productRepo := repository.NewProductRepository(queries, pool)
	productService := service.NewProductService(productRepo)
	productHandler := handler.NewProductHandler(productService)

	log.Println("Product layers initialized successfully")

	// =========================================
	// 6. INITIALIZE CATEGORY LAYERS
	// =========================================

	categoryRepo := repository.NewCategoryRepository(queries)
	categoryService := service.NewCategoryService(categoryRepo)
	categoryHandler := handler.NewCategoryHandler(categoryService)

	log.Println("Category layers initialized successfully")

	// =========================================
	// 7. INITIALIZE CUSTOMER LAYERS
	// =========================================

	customerRepo := repository.NewCustomerRepository(queries)
	customerService := service.NewCustomerService(customerRepo)
	customerHandler := handler.NewCustomerHandler(customerService)

	log.Println("Customer layers initialized successfully")

	// =========================================
	// 8. INITIALIZE CUSTOMER ADDRESS LAYERS
	// =========================================

	customerAddressRepo := repository.NewCustomerAddressRepository(
		queries,
		pool,
	)

	customerAddressService := service.NewCustomerAddressService(
		customerAddressRepo,
	)

	customerAddressHandler := handler.NewCustomerAddressHandler(
		customerAddressService,
	)

	log.Println("Customer address layers initialized successfully")

	// =========================================
	// 9. INITIALIZE CART LAYERS
	// =========================================

	cartRepo := repository.NewCartRepository(queries, pool)
	cartService := service.NewCartService(cartRepo)
	cartHandler := handler.NewCartHandler(cartService)

	log.Println("Cart layers initialized successfully")

	// =========================================
	// 10. INITIALIZE HTTP ROUTER
	// =========================================

	mux := http.NewServeMux()

	// =========================================
	// HEALTH CHECK ENDPOINTS
	// =========================================

	// Liveness endpoint.
	mux.HandleFunc("GET /health", func(
		w http.ResponseWriter,
		r *http.Request,
	) {
		writeJSON(w, http.StatusOK, APIResponse{
			Status:  "ok",
			Service: "hafiz-pintar-api",
			Message: "API is running",
		})
	})

	// Readiness endpoint.
	mux.HandleFunc("GET /ready", func(
		w http.ResponseWriter,
		r *http.Request,
	) {
		ctx, cancel := context.WithTimeout(
			r.Context(),
			3*time.Second,
		)
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

	// =========================================
	// PRODUCT ENDPOINTS
	// =========================================

	mux.HandleFunc(
		"GET /api/v1/products",
		productHandler.List,
	)

	mux.HandleFunc(
		"GET /api/v1/products/{slug}",
		productHandler.GetBySlug,
	)

	mux.HandleFunc(
		"POST /api/v1/products",
		productHandler.Create,
	)

	mux.HandleFunc(
		"GET /api/v1/products/{slug}/variants",
		productHandler.GetVariantsBySlug,
	)

	mux.HandleFunc(
		"GET /api/v1/products/{slug}/images",
		productHandler.GetImagesBySlug,
	)

	log.Println("Product routes registered successfully")

	// =========================================
	// CATEGORY ENDPOINTS
	// =========================================

	mux.HandleFunc(
		"GET /api/v1/categories",
		categoryHandler.List,
	)

	mux.HandleFunc(
		"GET /api/v1/categories/{slug}",
		categoryHandler.GetBySlug,
	)

	log.Println("Category routes registered successfully")

	// =========================================
	// CUSTOMER AUTH ENDPOINTS
	// =========================================

	mux.HandleFunc(
		"POST /api/v1/auth/register",
		customerHandler.Register,
	)

	mux.HandleFunc(
		"POST /api/v1/auth/login",
		customerHandler.Login,
	)

	log.Println("Customer authentication routes registered successfully")

	// =========================================
	// CUSTOMER PROFILE ENDPOINTS
	// =========================================

	mux.Handle(
		"GET /api/v1/customers/me",
		handler.JWTAuth(
			http.HandlerFunc(customerHandler.Me),
		),
	)

	log.Println("Customer profile routes registered successfully")

	// =========================================
	// CUSTOMER ADDRESS ENDPOINTS
	// =========================================

	mux.Handle(
		"GET /api/v1/customer-addresses",
		handler.JWTAuth(
			http.HandlerFunc(customerAddressHandler.List),
		),
	)

	mux.Handle(
		"GET /api/v1/customer-addresses/{id}",
		handler.JWTAuth(
			http.HandlerFunc(customerAddressHandler.GetByID),
		),
	)

	mux.Handle(
		"POST /api/v1/customer-addresses",
		handler.JWTAuth(
			http.HandlerFunc(customerAddressHandler.Create),
		),
	)

	mux.Handle(
		"PUT /api/v1/customer-addresses/{id}",
		handler.JWTAuth(
			http.HandlerFunc(customerAddressHandler.Update),
		),
	)

	mux.Handle(
		"DELETE /api/v1/customer-addresses/{id}",
		handler.JWTAuth(
			http.HandlerFunc(customerAddressHandler.Delete),
		),
	)

	mux.Handle(
		"PATCH /api/v1/customer-addresses/{id}/default",
		handler.JWTAuth(
			http.HandlerFunc(customerAddressHandler.SetDefault),
		),
	)

	log.Println("Customer address routes registered successfully")

	// =========================================
	// CART ENDPOINTS
	// =========================================

	mux.Handle(
		"GET /api/v1/cart",
		handler.JWTAuth(
			http.HandlerFunc(cartHandler.Get),
		),
	)

	mux.Handle(
		"POST /api/v1/cart/items",
		handler.JWTAuth(
			http.HandlerFunc(cartHandler.AddItem),
		),
	)

	mux.Handle(
		"PATCH /api/v1/cart/items/{id}",
		handler.JWTAuth(
			http.HandlerFunc(cartHandler.UpdateQuantity),
		),
	)

	mux.Handle(
		"DELETE /api/v1/cart/items/{id}",
		handler.JWTAuth(
			http.HandlerFunc(cartHandler.RemoveItem),
		),
	)

	mux.Handle(
		"DELETE /api/v1/cart/items",
		handler.JWTAuth(
			http.HandlerFunc(cartHandler.Clear),
		),
	)

	log.Println("Cart routes registered successfully")

	// =========================================
	// 11. INITIALIZE HTTP SERVER
	// =========================================

	server := &http.Server{
		Addr:              ":" + port,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	// =========================================
	// 12. START HTTP SERVER
	// =========================================

	go func() {
		log.Printf("Hafiz Pintar API running on port %s", port)

		if err := server.ListenAndServe(); err != nil &&
			err != http.ErrServerClosed {
			log.Printf("HTTP server error: %v", err)
		}
	}()

	// =========================================
	// 13. GRACEFUL SHUTDOWN
	// =========================================

	stop := make(chan os.Signal, 1)

	signal.Notify(
		stop,
		os.Interrupt,
		syscall.SIGTERM,
	)

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
