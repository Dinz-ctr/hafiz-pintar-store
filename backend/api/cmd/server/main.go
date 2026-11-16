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

// =========================================
// API RESPONSE
// =========================================

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

// =========================================
// MAIN APPLICATION
// =========================================

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
	// 10. INITIALIZE ORDER LAYERS
	// =========================================

	orderRepo := repository.NewOrderRepository(pool)
	orderService := service.NewOrderService(orderRepo)
	orderHandler := handler.NewOrderHandler(orderService)

	log.Println("Order layers initialized successfully")

	// =========================================
	// 11. INITIALIZE PAYMENT LAYERS
	// =========================================

	paymentRepo := repository.NewPaymentRepository(queries, pool)
	paymentService := service.NewPaymentService(paymentRepo)
	paymentHandler := handler.NewPaymentHandler(paymentService)

	log.Println("Payment layers initialized successfully")

	// =========================================
	// 12. INITIALIZE ADMIN LAYERS
	// =========================================

	adminRepo := repository.NewAdminRepository(pool)
	adminService := service.NewAdminService(adminRepo)
	adminHandler := handler.NewAdminHandler(adminService)

	log.Println("Admin layers initialized successfully")

	// =========================================
	// 13. INITIALIZE ADMIN DASHBOARD LAYERS
	// =========================================

	adminDashboardRepo := repository.NewAdminDashboardRepository(pool)

	adminDashboardService := service.NewAdminDashboardService(
		adminDashboardRepo,
	)

	adminDashboardHandler := handler.NewAdminDashboardHandler(
		adminDashboardService,
	)

	log.Println("Admin dashboard layers initialized successfully")

	// =========================================
	// 14. INITIALIZE ADMIN PRODUCT LAYERS
	// =========================================

	adminProductRepo := repository.NewAdminProductRepository(pool)

	adminProductService := service.NewAdminProductService(
		adminProductRepo,
		productRepo,
	)

	adminProductHandler := handler.NewAdminProductHandler(
		adminProductService,
		productHandler,
	)

	log.Println("Admin product layers initialized successfully")

	// =========================================
	// 15. INITIALIZE HTTP ROUTER
	// =========================================

	mux := http.NewServeMux()

	// =========================================
	// 16. HEALTH CHECK ENDPOINTS
	// =========================================

	// GET /health
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

	// GET /ready
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
	// 17. PUBLIC PRODUCT ENDPOINTS
	// =========================================

	// GET /api/v1/products
	mux.HandleFunc(
		"GET /api/v1/products",
		productHandler.List,
	)

	// GET /api/v1/products/{slug}
	mux.HandleFunc(
		"GET /api/v1/products/{slug}",
		productHandler.GetBySlug,
	)

	// GET /api/v1/products/{slug}/variants
	mux.HandleFunc(
		"GET /api/v1/products/{slug}/variants",
		productHandler.GetVariantsBySlug,
	)

	// GET /api/v1/products/{slug}/images
	mux.HandleFunc(
		"GET /api/v1/products/{slug}/images",
		productHandler.GetImagesBySlug,
	)

	log.Println("Public product routes registered successfully")

	// =========================================
	// 18. CATEGORY ENDPOINTS
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
	// 19. CUSTOMER AUTHENTICATION ENDPOINTS
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
	// 20. CUSTOMER PROFILE ENDPOINTS
	// =========================================

	mux.Handle(
		"GET /api/v1/customers/me",
		handler.JWTAuth(
			http.HandlerFunc(customerHandler.Me),
		),
	)

	log.Println("Customer profile routes registered successfully")

	// =========================================
	// 21. CUSTOMER ADDRESS ENDPOINTS
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
	// 22. SHOPPING CART ENDPOINTS
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
	// 23. ORDER ENDPOINTS
	// =========================================

	mux.Handle(
		"POST /api/v1/orders/checkout",
		handler.JWTAuth(
			http.HandlerFunc(orderHandler.Checkout),
		),
	)

	mux.Handle(
		"GET /api/v1/orders",
		handler.JWTAuth(
			http.HandlerFunc(orderHandler.GetCustomerOrders),
		),
	)

	mux.Handle(
		"GET /api/v1/orders/{id}",
		handler.JWTAuth(
			http.HandlerFunc(orderHandler.GetCustomerOrderDetail),
		),
	)

	log.Println("Order routes registered successfully")

	// =========================================
	// 24. PAYMENT ENDPOINTS
	// =========================================

	mux.Handle(
		"POST /api/v1/payments",
		handler.JWTAuth(
			http.HandlerFunc(paymentHandler.Create),
		),
	)

	log.Println("Payment routes registered successfully")

	// =========================================
	// 25. ADMIN AUTHENTICATION ENDPOINTS
	// =========================================

	// POST /api/v1/admin/auth/login
	mux.HandleFunc(
		"POST /api/v1/admin/auth/login",
		adminHandler.Login,
	)

	// GET /api/v1/admin/me
	mux.Handle(
		"GET /api/v1/admin/me",
		handler.AdminJWTAuth(
			http.HandlerFunc(adminHandler.Me),
		),
	)

	log.Println("Admin authentication routes registered successfully")

	// =========================================
	// 26. ADMIN DASHBOARD ENDPOINTS
	// =========================================

	// GET /api/v1/admin/dashboard
	mux.Handle(
		"GET /api/v1/admin/dashboard",
		handler.AdminJWTAuth(
			http.HandlerFunc(adminDashboardHandler.GetDashboard),
		),
	)

	log.Println("Admin dashboard routes registered successfully")

	// =========================================
	// 27. ADMIN PRODUCT ENDPOINTS
	// =========================================

	// GET /api/v1/admin/products
	mux.Handle(
		"GET /api/v1/admin/products",
		handler.AdminJWTAuth(
			http.HandlerFunc(adminProductHandler.List),
		),
	)

	// GET /api/v1/admin/products/{id}
	mux.Handle(
		"GET /api/v1/admin/products/{id}",
		handler.AdminJWTAuth(
			http.HandlerFunc(adminProductHandler.GetByID),
		),
	)

	// POST /api/v1/admin/products
	mux.Handle(
		"POST /api/v1/admin/products",
		handler.AdminJWTAuth(
			http.HandlerFunc(adminProductHandler.Create),
		),
	)

	// PUT /api/v1/admin/products/{id}
	mux.Handle(
		"PUT /api/v1/admin/products/{id}",
		handler.AdminJWTAuth(
			http.HandlerFunc(adminProductHandler.UpdateDetails),
		),
	)

	// PATCH /api/v1/admin/products/{id}/status
	mux.Handle(
		"PATCH /api/v1/admin/products/{id}/status",
		handler.AdminJWTAuth(
			http.HandlerFunc(adminProductHandler.UpdateStatus),
		),
	)

	// DELETE /api/v1/admin/products/{id}
	// Archive product, bukan hard delete.
	mux.Handle(
		"DELETE /api/v1/admin/products/{id}",
		handler.AdminJWTAuth(
			http.HandlerFunc(adminProductHandler.Archive),
		),
	)

	// PUT /api/v1/admin/products/{id}/variants/{variantId}
	mux.Handle(
		"PUT /api/v1/admin/products/{id}/variants/{variantId}",
		handler.AdminJWTAuth(
			http.HandlerFunc(adminProductHandler.UpdateVariant),
		),
	)

	// DELETE /api/v1/admin/products/{id}/variants/{variantId}
	mux.Handle(
		"DELETE /api/v1/admin/products/{id}/variants/{variantId}",
		handler.AdminJWTAuth(
			http.HandlerFunc(adminProductHandler.DeleteVariant),
		),
	)

	log.Println("Admin product routes registered successfully")

	// =========================================
	// 28. INITIALIZE HTTP SERVER
	// =========================================

	server := &http.Server{
		Addr:              ":" + port,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	// =========================================
	// 29. START HTTP SERVER
	// =========================================

	go func() {
		log.Printf("Hafiz Pintar API running on port %s", port)

		if err := server.ListenAndServe(); err != nil &&
			err != http.ErrServerClosed {
			log.Printf("HTTP server error: %v", err)
		}
	}()

	// =========================================
	// 30. GRACEFUL SHUTDOWN
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
