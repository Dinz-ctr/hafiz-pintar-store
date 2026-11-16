package service

import (
	"context"
	"errors"
	"os"
	"strconv"
	"strings"
	"time"

	"hafiz-pintar-store/backend/api/internal/repository"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrAdminInvalidCredentials = errors.New("invalid admin credentials")
	ErrAdminInactive           = errors.New("admin account is inactive")
	ErrAdminInvalidID          = errors.New("invalid admin ID")
	ErrAdminTokenConfig        = errors.New("admin JWT secret is not configured")
	ErrAdminInvalidInput       = errors.New("invalid admin input")
)

type AdminService struct {
	repo *repository.AdminRepository
}

type AdminLoginResult struct {
	Admin       repository.Admin `json:"admin"`
	AccessToken string           `json:"access_token"`
	TokenType   string           `json:"token_type"`
	ExpiresIn   int64            `json:"expires_in"`
}

func NewAdminService(repo *repository.AdminRepository) *AdminService {
	return &AdminService{repo: repo}
}

// ADMIN LOGIN
func (s *AdminService) Login(
	ctx context.Context,
	email string,
	password string,
) (AdminLoginResult, error) {
	email = strings.ToLower(strings.TrimSpace(email))

	if email == "" || password == "" {
		return AdminLoginResult{}, ErrAdminInvalidCredentials
	}

	admin, err := s.repo.GetByEmail(ctx, email)
	if errors.Is(err, repository.ErrAdminNotFound) {
		return AdminLoginResult{}, ErrAdminInvalidCredentials
	}
	if err != nil {
		return AdminLoginResult{}, err
	}

	if !admin.IsActive {
		return AdminLoginResult{}, ErrAdminInactive
	}

	if err := bcrypt.CompareHashAndPassword(
		[]byte(admin.PasswordHash),
		[]byte(password),
	); err != nil {
		return AdminLoginResult{}, ErrAdminInvalidCredentials
	}

	secret := os.Getenv("ADMIN_JWT_SECRET")
	if len(secret) < 32 {
		return AdminLoginResult{}, ErrAdminTokenConfig
	}

	const tokenDuration = 24 * time.Hour
	now := time.Now()

	claims := jwt.RegisteredClaims{
		Subject:   strconv.FormatInt(admin.ID, 10),
		Issuer:    "hafiz-pintar-admin",
		IssuedAt:  jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(now.Add(tokenDuration)),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	signedToken, err := token.SignedString([]byte(secret))
	if err != nil {
		return AdminLoginResult{}, err
	}

	admin.PasswordHash = ""

	return AdminLoginResult{
		Admin:       admin,
		AccessToken: signedToken,
		TokenType:   "Bearer",
		ExpiresIn:   int64(tokenDuration.Seconds()),
	}, nil
}

// GET ADMIN PROFILE
func (s *AdminService) GetByID(
	ctx context.Context,
	id int64,
) (repository.Admin, error) {
	if id <= 0 {
		return repository.Admin{}, ErrAdminInvalidID
	}

	return s.repo.GetByID(ctx, id)
}

// CREATE ADMIN
func (s *AdminService) Create(
	ctx context.Context,
	fullName string,
	email string,
	password string,
) (repository.Admin, error) {
	fullName = strings.TrimSpace(fullName)
	email = strings.ToLower(strings.TrimSpace(email))

	if fullName == "" ||
		email == "" ||
		len(password) < 8 {
		return repository.Admin{}, ErrAdminInvalidInput
	}

	hash, err := bcrypt.GenerateFromPassword(
		[]byte(password),
		bcrypt.DefaultCost,
	)
	if err != nil {
		return repository.Admin{}, err
	}

	return s.repo.Create(
		ctx,
		fullName,
		email,
		string(hash),
	)
}
