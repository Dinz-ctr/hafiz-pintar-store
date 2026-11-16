package handler

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"hafiz-pintar-store/backend/api/internal/service"

	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"golang.org/x/crypto/bcrypt"
)

type CustomerHandler struct {
	service *service.CustomerService
}

func NewCustomerHandler(svc *service.CustomerService) *CustomerHandler {
	return &CustomerHandler{
		service: svc,
	}
}

type RegisterCustomerRequest struct {
	FullName string  `json:"full_name"`
	Email    *string `json:"email"`
	Phone    *string `json:"phone"`
	Password string  `json:"password"`
}

type LoginCustomerRequest struct {
	Identifier string `json:"identifier"`
	Password   string `json:"password"`
}

type CustomerResponse struct {
	ID        int64   `json:"id"`
	FullName  string  `json:"full_name"`
	Email     *string `json:"email"`
	Phone     *string `json:"phone"`
	IsActive  bool    `json:"is_active"`
	CreatedAt string  `json:"created_at"`
	UpdatedAt string  `json:"updated_at"`
}

func writeCustomerJSON(
	w http.ResponseWriter,
	status int,
	data any,
) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Printf("Customer JSON response error: %v", err)
	}
}

func customerError(
	w http.ResponseWriter,
	status int,
	message string,
) {
	writeCustomerJSON(w, status, map[string]string{
		"error": message,
	})
}

// Register mendaftarkan customer baru.
func (h *CustomerHandler) Register(
	w http.ResponseWriter,
	r *http.Request,
) {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	defer r.Body.Close()

	var input RegisterCustomerRequest

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&input); err != nil {
		customerError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		customerError(
			w,
			http.StatusBadRequest,
			"request body must contain only one JSON object",
		)
		return
	}

	input.FullName = strings.TrimSpace(input.FullName)
	input.Password = strings.TrimSpace(input.Password)

	if input.FullName == "" {
		customerError(w, http.StatusBadRequest, "full name is required")
		return
	}

	if utf8.RuneCountInString(input.FullName) > 150 {
		customerError(w, http.StatusBadRequest, "full name is too long")
		return
	}

	if input.Email != nil {
		email := strings.ToLower(strings.TrimSpace(*input.Email))
		if email == "" {
			input.Email = nil
		} else {
			input.Email = &email
		}
	}

	if input.Phone != nil {
		phone := strings.TrimSpace(*input.Phone)
		if phone == "" {
			input.Phone = nil
		} else {
			input.Phone = &phone
		}
	}

	if input.Email == nil && input.Phone == nil {
		customerError(
			w,
			http.StatusBadRequest,
			"email or phone is required",
		)
		return
	}

	if input.Password == "" {
		customerError(w, http.StatusBadRequest, "password is required")
		return
	}

	if len(input.Password) < 8 {
		customerError(
			w,
			http.StatusBadRequest,
			"password must contain at least 8 bytes",
		)
		return
	}

	if len([]byte(input.Password)) > 72 {
		customerError(
			w,
			http.StatusBadRequest,
			"password must not exceed 72 bytes",
		)
		return
	}

	passwordHash, err := bcrypt.GenerateFromPassword(
		[]byte(input.Password),
		bcrypt.DefaultCost,
	)
	if err != nil {
		log.Printf("Password hashing failed: %v", err)
		customerError(
			w,
			http.StatusInternalServerError,
			"failed to process password",
		)
		return
	}

	hashString := string(passwordHash)

	customer, err := h.service.Create(
		r.Context(),
		input.FullName,
		input.Email,
		input.Phone,
		&hashString,
	)
	if err != nil {
		var pgErr *pgconn.PgError

		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			switch pgErr.ConstraintName {
			case "customers_email_key":
				customerError(
					w,
					http.StatusConflict,
					"email is already registered",
				)
				return

			case "customers_phone_key":
				customerError(
					w,
					http.StatusConflict,
					"phone number is already registered",
				)
				return
			}
		}

		if errors.Is(err, service.ErrInvalidCustomerName) ||
			errors.Is(err, service.ErrInvalidCustomerContact) ||
			errors.Is(err, service.ErrInvalidCustomerPassword) {
			customerError(w, http.StatusBadRequest, err.Error())
			return
		}

		log.Printf("Customer registration failed: %v", err)
		customerError(
			w,
			http.StatusInternalServerError,
			"failed to register customer",
		)
		return
	}

	response := CustomerResponse{
		ID:        customer.ID,
		FullName:  customer.FullName,
		Email:     customer.Email,
		Phone:     customer.Phone,
		IsActive:  customer.IsActive,
		CreatedAt: customer.CreatedAt.Time.UTC().Format(time.RFC3339),
		UpdatedAt: customer.UpdatedAt.Time.UTC().Format(time.RFC3339),
	}

	writeCustomerJSON(w, http.StatusCreated, map[string]any{
		"message": "customer registered successfully",
		"data":    response,
	})
}

// Login melakukan autentikasi customer dan menerbitkan JWT.
func (h *CustomerHandler) Login(
	w http.ResponseWriter,
	r *http.Request,
) {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	defer r.Body.Close()

	var input LoginCustomerRequest

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&input); err != nil {
		customerError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		customerError(
			w,
			http.StatusBadRequest,
			"request body must contain only one JSON object",
		)
		return
	}

	input.Identifier = strings.TrimSpace(input.Identifier)

	if input.Identifier == "" || input.Password == "" {
		customerError(
			w,
			http.StatusBadRequest,
			"identifier and password are required",
		)
		return
	}

	if len([]byte(input.Password)) > 72 {
		customerError(
			w,
			http.StatusBadRequest,
			"invalid credentials",
		)
		return
	}

	// Cari customer berdasarkan email atau nomor HP.
	identifier := input.Identifier
	if strings.Contains(identifier, "@") {
		identifier = strings.ToLower(identifier)
	}

	customer, err := h.service.GetByEmailOrPhone(
		r.Context(),
		identifier,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			customerError(
				w,
				http.StatusUnauthorized,
				"invalid credentials",
			)
			return
		}

		log.Printf("Customer login lookup failed: %v", err)
		customerError(
			w,
			http.StatusInternalServerError,
			"failed to authenticate customer",
		)
		return
	}

	// Akun tidak aktif tidak boleh login.
	if !customer.IsActive {
		customerError(
			w,
			http.StatusForbidden,
			"account is inactive",
		)
		return
	}

	if customer.PasswordHash == nil ||
		bcrypt.CompareHashAndPassword(
			[]byte(*customer.PasswordHash),
			[]byte(input.Password),
		) != nil {
		customerError(
			w,
			http.StatusUnauthorized,
			"invalid credentials",
		)
		return
	}

	// Ambil JWT secret dari environment.
	jwtSecret := os.Getenv("JWT_SECRET")

	if len(jwtSecret) < 32 {
		log.Println("JWT_SECRET is missing or too short")
		customerError(
			w,
			http.StatusInternalServerError,
			"authentication service is not configured",
		)
		return
	}

	// Token berlaku selama 24 jam.
	now := time.Now()
	expiresAt := now.Add(24 * time.Hour)

	claims := jwt.RegisteredClaims{
		Subject:   strconv.FormatInt(customer.ID, 10),
		Issuer:    "hafiz-pintar-api",
		IssuedAt:  jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(expiresAt),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	signedToken, err := token.SignedString([]byte(jwtSecret))
	if err != nil {
		log.Printf("JWT signing failed: %v", err)
		customerError(
			w,
			http.StatusInternalServerError,
			"failed to generate access token",
		)
		return
	}

	response := CustomerResponse{
		ID:        customer.ID,
		FullName:  customer.FullName,
		Email:     customer.Email,
		Phone:     customer.Phone,
		IsActive:  customer.IsActive,
		CreatedAt: customer.CreatedAt.Time.UTC().Format(time.RFC3339),
		UpdatedAt: customer.UpdatedAt.Time.UTC().Format(time.RFC3339),
	}

	writeCustomerJSON(w, http.StatusOK, map[string]any{
		"message":      "login successful",
		"access_token": signedToken,
		"token_type":   "Bearer",
		"expires_at":   expiresAt.UTC().Format(time.RFC3339),
		"data":         response,
	})
}
