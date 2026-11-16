package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"hafiz-pintar-store/backend/api/internal/repository"
	"hafiz-pintar-store/backend/api/internal/service"

	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	// 1. Baca konfigurasi dari environment.
	databaseURL := os.Getenv("DATABASE_URL")
	fullName := strings.TrimSpace(os.Getenv("ADMIN_FULL_NAME"))
	email := strings.ToLower(strings.TrimSpace(os.Getenv("ADMIN_EMAIL")))
	password := os.Getenv("ADMIN_SEED_PASSWORD")

	if databaseURL == "" {
		log.Fatal("DATABASE_URL is required")
	}

	if fullName == "" || email == "" || len(password) < 8 {
		log.Fatal("ADMIN_FULL_NAME, ADMIN_EMAIL, and a password of at least 8 characters are required")
	}

	// 2. Koneksi database.
	ctx, cancel := context.WithTimeout(
		context.Background(),
		15*time.Second,
	)
	defer cancel()

	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		log.Fatalf("Database connection failed: %v", err)
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		log.Fatalf("Database ping failed: %v", err)
	}

	// 3. Inisialisasi repository dan service.
	adminRepo := repository.NewAdminRepository(pool)
	adminService := service.NewAdminService(adminRepo)

	// 4. Pastikan email belum digunakan.
	_, err = adminRepo.GetByEmail(ctx, email)

	if err == nil {
		log.Fatal("Admin email already exists. No account was created.")
	}

	if !errors.Is(err, repository.ErrAdminNotFound) {
		log.Fatalf("Failed to check existing admin: %v", err)
	}

	// 5. Buat admin. Service akan melakukan hashing bcrypt.
	admin, err := adminService.Create(
		ctx,
		fullName,
		email,
		password,
	)
	if err != nil {
		log.Fatalf("Failed to create admin: %v", err)
	}

	fmt.Println("Admin account created successfully.")
	fmt.Printf("Admin ID: %d\n", admin.ID)
	fmt.Printf("Admin name: %s\n", admin.FullName)
	fmt.Printf("Admin email: %s\n", admin.Email)
}
