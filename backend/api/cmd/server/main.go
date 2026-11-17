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
	// 15. INITIALIZE ADMIN CATEGORY LAYERS
	// =========================================

	adminCategoryRepo := repository.NewAdminCategoryRepository(pool)

	adminCategoryService := service.NewAdminCategoryService(
		adminCategoryRepo,
	)

	adminCategoryHandler := handler.NewAdminCategoryHandler(
		adminCategoryService,
	)

	log.Println("Admin category layers initialized successfully")

	// =========================================
	// 16. INITIALIZE ADMIN CUSTOMER LAYERS
	// =========================================

	adminCustomerRepo := repository.NewAdminCustomerRepository(pool)

	adminCustomerService := service.NewAdminCustomerService(
		adminCustomerRepo,
	)

	adminCustomerHandler := handler.NewAdminCustomerHandler(
		adminCustomerService,
	)

	log.Println("Admin customer layers initialized successfully")

	// =========================================
	// INITIALIZE ADMIN ORDER LAYERS
	// =========================================

	adminOrderRepo := repository.NewAdminOrderRepository(pool)

	adminOrderService := service.NewAdminOrderService(
		adminOrderRepo,
	)

	adminOrderHandler := handler.NewAdminOrderHandler(
		adminOrderService,
	)

	log.Println("Admin order layers initialized successfully")

	// =========================================
	// ADMIN PAYMENT LAYERS
	// =========================================

	adminPaymentRepo := repository.NewAdminPaymentRepository(pool)

	adminPaymentService := service.NewAdminPaymentService(
		adminPaymentRepo,
	)

	adminPaymentHandler := handler.NewAdminPaymentHandler(
		adminPaymentService,
	)

	log.Println("Admin payment layers initialized successfully")

	// =========================================
	// 17. INITIALIZE HTTP ROUTER
	// =========================================

	mux := http.NewServeMux()

	// =========================================
	// 18. HEALTH CHECK ENDPOINTS
	// =========================================

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
	// 19. PUBLIC PRODUCT ENDPOINTS
	// =========================================

	mux.HandleFunc("GET /api/v1/products", productHandler.List)
	mux.HandleFunc("GET /api/v1/products/{slug}", productHandler.GetBySlug)
	mux.HandleFunc("GET /api/v1/products/{slug}/variants", productHandler.GetVariantsBySlug)
	mux.HandleFunc("GET /api/v1/products/{slug}/images", productHandler.GetImagesBySlug)

	log.Println("Public product routes registered successfully")

	// =========================================
	// 20. PUBLIC CATEGORY ENDPOINTS
	// =========================================

	mux.HandleFunc("GET /api/v1/categories", categoryHandler.List)
	mux.HandleFunc("GET /api/v1/categories/{slug}", categoryHandler.GetBySlug)

	log.Println("Category routes registered successfully")

	// =========================================
	// 21. CUSTOMER AUTHENTICATION ENDPOINTS
	// =========================================

	mux.HandleFunc("POST /api/v1/auth/register", customerHandler.Register)
	mux.HandleFunc("POST /api/v1/auth/login", customerHandler.Login)

	log.Println("Customer authentication routes registered successfully")

	// =========================================
	// 22. CUSTOMER PROFILE ENDPOINTS
	// =========================================

	mux.Handle(
		"GET /api/v1/customers/me",
		handler.JWTAuth(http.HandlerFunc(customerHandler.Me)),
	)

	log.Println("Customer profile routes registered successfully")

	// =========================================
	// 23. CUSTOMER ADDRESS ENDPOINTS
	// =========================================

	mux.Handle(
		"GET /api/v1/customer-addresses",
		handler.JWTAuth(http.HandlerFunc(customerAddressHandler.List)),
	)

	mux.Handle(
		"GET /api/v1/customer-addresses/{id}",
		handler.JWTAuth(http.HandlerFunc(customerAddressHandler.GetByID)),
	)

	mux.Handle(
		"POST /api/v1/customer-addresses",
		handler.JWTAuth(http.HandlerFunc(customerAddressHandler.Create)),
	)

	mux.Handle(
		"PUT /api/v1/customer-addresses/{id}",
		handler.JWTAuth(http.HandlerFunc(customerAddressHandler.Update)),
	)

	mux.Handle(
		"DELETE /api/v1/customer-addresses/{id}",
		handler.JWTAuth(http.HandlerFunc(customerAddressHandler.Delete)),
	)

	mux.Handle(
		"PATCH /api/v1/customer-addresses/{id}/default",
		handler.JWTAuth(http.HandlerFunc(customerAddressHandler.SetDefault)),
	)

	log.Println("Customer address routes registered successfully")

	// =========================================
	// 24. SHOPPING CART ENDPOINTS
	// =========================================

	mux.Handle(
		"GET /api/v1/cart",
		handler.JWTAuth(http.HandlerFunc(cartHandler.Get)),
	)

	mux.Handle(
		"POST /api/v1/cart/items",
		handler.JWTAuth(http.HandlerFunc(cartHandler.AddItem)),
	)

	mux.Handle(
		"PATCH /api/v1/cart/items/{id}",
		handler.JWTAuth(http.HandlerFunc(cartHandler.UpdateQuantity)),
	)

	mux.Handle(
		"DELETE /api/v1/cart/items/{id}",
		handler.JWTAuth(http.HandlerFunc(cartHandler.RemoveItem)),
	)

	mux.Handle(
		"DELETE /api/v1/cart/items",
		handler.JWTAuth(http.HandlerFunc(cartHandler.Clear)),
	)

	log.Println("Cart routes registered successfully")

	// =========================================
	// 25. ORDER ENDPOINTS
	// =========================================

	mux.Handle(
		"POST /api/v1/orders/checkout",
		handler.JWTAuth(http.HandlerFunc(orderHandler.Checkout)),
	)

	mux.Handle(
		"GET /api/v1/orders",
		handler.JWTAuth(http.HandlerFunc(orderHandler.GetCustomerOrders)),
	)

	mux.Handle(
		"GET /api/v1/orders/{id}",
		handler.JWTAuth(http.HandlerFunc(orderHandler.GetCustomerOrderDetail)),
	)

	log.Println("Order routes registered successfully")

	// =========================================
	// 26. PAYMENT ENDPOINTS
	// =========================================

	mux.Handle(
		"POST /api/v1/payments",
		handler.JWTAuth(http.HandlerFunc(paymentHandler.Create)),
	)

	mux.Handle(
		"GET /api/v1/payments/{id}",
		handler.JWTAuth(
			http.HandlerFunc(paymentHandler.GetByID),
		),
	)

	log.Println("Payment routes registered successfully")

	// =========================================
	// 27. ADMIN AUTHENTICATION ENDPOINTS
	// =========================================

	mux.HandleFunc("POST /api/v1/admin/auth/login", adminHandler.Login)

	mux.Handle(
		"GET /api/v1/admin/me",
		handler.AdminJWTAuth(http.HandlerFunc(adminHandler.Me)),
	)

	log.Println("Admin authentication routes registered successfully")

	// =========================================
	// 28. ADMIN DASHBOARD ENDPOINTS
	// =========================================

	mux.Handle(
		"GET /api/v1/admin/dashboard",
		handler.AdminJWTAuth(http.HandlerFunc(adminDashboardHandler.GetDashboard)),
	)

	log.Println("Admin dashboard routes registered successfully")

	// =========================================
	// 29. ADMIN PRODUCT ENDPOINTS
	// =========================================

	mux.Handle(
		"GET /api/v1/admin/products",
		handler.AdminJWTAuth(http.HandlerFunc(adminProductHandler.List)),
	)

	mux.Handle(
		"GET /api/v1/admin/products/{id}",
		handler.AdminJWTAuth(http.HandlerFunc(adminProductHandler.GetByID)),
	)

	mux.Handle(
		"POST /api/v1/admin/products",
		handler.AdminJWTAuth(http.HandlerFunc(adminProductHandler.Create)),
	)

	mux.Handle(
		"PUT /api/v1/admin/products/{id}",
		handler.AdminJWTAuth(http.HandlerFunc(adminProductHandler.UpdateDetails)),
	)

	mux.Handle(
		"PATCH /api/v1/admin/products/{id}/status",
		handler.AdminJWTAuth(http.HandlerFunc(adminProductHandler.UpdateStatus)),
	)

	// Archive product, bukan hard delete.
	mux.Handle(
		"DELETE /api/v1/admin/products/{id}",
		handler.AdminJWTAuth(http.HandlerFunc(adminProductHandler.Archive)),
	)

	mux.Handle(
		"PUT /api/v1/admin/products/{id}/variants/{variantId}",
		handler.AdminJWTAuth(http.HandlerFunc(adminProductHandler.UpdateVariant)),
	)

	mux.Handle(
		"DELETE /api/v1/admin/products/{id}/variants/{variantId}",
		handler.AdminJWTAuth(http.HandlerFunc(adminProductHandler.DeleteVariant)),
	)

	log.Println("Admin product routes registered successfully")

	// =========================================
	// 30. ADMIN CATEGORY ENDPOINTS
	// =========================================

	mux.Handle(
		"GET /api/v1/admin/categories",
		handler.AdminJWTAuth(http.HandlerFunc(adminCategoryHandler.List)),
	)

	mux.Handle(
		"GET /api/v1/admin/categories/{id}",
		handler.AdminJWTAuth(http.HandlerFunc(adminCategoryHandler.GetByID)),
	)

	mux.Handle(
		"POST /api/v1/admin/categories",
		handler.AdminJWTAuth(http.HandlerFunc(adminCategoryHandler.Create)),
	)

	mux.Handle(
		"PUT /api/v1/admin/categories/{id}",
		handler.AdminJWTAuth(http.HandlerFunc(adminCategoryHandler.UpdateDetails)),
	)

	mux.Handle(
		"PATCH /api/v1/admin/categories/{id}/status",
		handler.AdminJWTAuth(http.HandlerFunc(adminCategoryHandler.UpdateStatus)),
	)

	// Soft delete: mengubah is_active menjadi FALSE.
	mux.Handle(
		"DELETE /api/v1/admin/categories/{id}",
		handler.AdminJWTAuth(http.HandlerFunc(adminCategoryHandler.Archive)),
	)

	log.Println("Admin category routes registered successfully")

	// =========================================
	// 31. ADMIN CUSTOMER ENDPOINTS
	// =========================================

	mux.Handle(
		"GET /api/v1/admin/customers",
		handler.AdminJWTAuth(http.HandlerFunc(adminCustomerHandler.List)),
	)

	mux.Handle(
		"GET /api/v1/admin/customers/{id}",
		handler.AdminJWTAuth(http.HandlerFunc(adminCustomerHandler.GetByID)),
	)

	mux.Handle(
		"PATCH /api/v1/admin/customers/{id}/status",
		handler.AdminJWTAuth(http.HandlerFunc(adminCustomerHandler.UpdateStatus)),
	)

	mux.Handle(
		"GET /api/v1/admin/customers/{id}/addresses",
		handler.AdminJWTAuth(http.HandlerFunc(adminCustomerHandler.ListAddresses)),
	)

	mux.Handle(
		"GET /api/v1/admin/customers/{id}/orders",
		handler.AdminJWTAuth(http.HandlerFunc(adminCustomerHandler.ListOrders)),
	)

	log.Println("Admin customer routes registered successfully")

	// =========================================
	// ADMIN ORDER ENDPOINTS
	// =========================================

	mux.Handle(
		"GET /api/v1/admin/orders",
		handler.AdminJWTAuth(
			http.HandlerFunc(adminOrderHandler.List),
		),
	)

	mux.Handle(
		"GET /api/v1/admin/orders/{id}",
		handler.AdminJWTAuth(
			http.HandlerFunc(adminOrderHandler.GetByID),
		),
	)

	mux.Handle(
		"PATCH /api/v1/admin/orders/{id}/status",
		handler.AdminJWTAuth(
			http.HandlerFunc(adminOrderHandler.UpdateStatus),
		),
	)

	// Update kurir dan nomor resi, sekaligus mengubah status menjadi shipped.
	mux.Handle(
		"PATCH /api/v1/admin/orders/{id}/shipping",
		handler.AdminJWTAuth(
			http.HandlerFunc(adminOrderHandler.UpdateShipping),
		),
	)

	log.Println("Admin order routes registered successfully")

	// =========================================
	// ADMIN PAYMENT ENDPOINTS
	// =========================================

	mux.Handle(
		"GET /api/v1/admin/payments",
		handler.AdminJWTAuth(
			http.HandlerFunc(adminPaymentHandler.List),
		),
	)

	mux.Handle(
		"GET /api/v1/admin/payments/{id}",
		handler.AdminJWTAuth(
			http.HandlerFunc(adminPaymentHandler.GetByID),
		),
	)

	mux.Handle(
		"PATCH /api/v1/admin/payments/{id}/status",
		handler.AdminJWTAuth(
			http.HandlerFunc(adminPaymentHandler.UpdateStatus),
		),
	)

	log.Println("Admin payment routes registered successfully")

	// =========================================
	// 32. INITIALIZE HTTP SERVER
	// =========================================

	server := &http.Server{
		Addr:              ":" + port,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	// =========================================
	// 33. START HTTP SERVER
	// =========================================

	go func() {
		log.Printf("Hafiz Pintar API running on port %s", port)

		if err := server.ListenAndServe(); err != nil &&
			err != http.ErrServerClosed {
			log.Printf("HTTP server error: %v", err)
		}
	}()

	// =========================================
	// 34. GRACEFUL SHUTDOWN
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
