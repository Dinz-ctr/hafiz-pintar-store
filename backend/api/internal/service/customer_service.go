package service

import (
	"context"
	"errors"
	"strings"

	"hafiz-pintar-store/backend/api/internal/database/sqlc"
	"hafiz-pintar-store/backend/api/internal/repository"
	
)

var (
	ErrInvalidCustomerName     = errors.New("customer name is required")
	ErrInvalidCustomerContact  = errors.New("email or phone is required")
	ErrInvalidCustomerPassword = errors.New("customer password hash is required")
	ErrInvalidCustomerID       = errors.New("customer ID must be greater than zero")
)

type CustomerService struct {
	repository *repository.CustomerRepository
}

func NewCustomerService(
	repo *repository.CustomerRepository,
) *CustomerService {
	return &CustomerService{
		repository: repo,
	}
}

// GetByEmail mengambil customer berdasarkan email.
func (s *CustomerService) GetByEmail(
	ctx context.Context,
	email *string,
) (sqlc.Customer, error) {
	email = normalizeOptionalString(email)

	if email == nil {
		return sqlc.Customer{}, ErrInvalidCustomerContact
	}

	*email = strings.ToLower(*email)

	return s.repository.GetByEmail(ctx, email)
}

// GetByPhone mengambil customer berdasarkan nomor telepon.
func (s *CustomerService) GetByPhone(
	ctx context.Context,
	phone *string,
) (sqlc.Customer, error) {
	phone = normalizeOptionalString(phone)

	if phone == nil {
		return sqlc.Customer{}, ErrInvalidCustomerContact
	}

	return s.repository.GetByPhone(ctx, phone)
}

// GetByEmailOrPhone mengambil customer berdasarkan email atau nomor telepon.
func (s *CustomerService) GetByEmailOrPhone(
	ctx context.Context,
	identifier string,
) (sqlc.Customer, error) {
	identifier = strings.TrimSpace(identifier)

	if identifier == "" {
		return sqlc.Customer{}, ErrInvalidCustomerContact
	}

	if strings.Contains(identifier, "@") {
		identifier = strings.ToLower(identifier)
	}

	return s.repository.GetByEmailOrPhone(ctx, &identifier)
}

// Create membuat customer baru.
func (s *CustomerService) Create(
	ctx context.Context,
	fullName string,
	email *string,
	phone *string,
	passwordHash *string,
) (sqlc.CreateCustomerRow, error) {
	fullName = strings.TrimSpace(fullName)
	email = normalizeOptionalString(email)
	phone = normalizeOptionalString(phone)

	if fullName == "" {
		return sqlc.CreateCustomerRow{}, ErrInvalidCustomerName
	}

	if email == nil && phone == nil {
		return sqlc.CreateCustomerRow{}, ErrInvalidCustomerContact
	}

	if email != nil {
		*email = strings.ToLower(*email)
	}

	passwordHash = normalizeOptionalString(passwordHash)

	if passwordHash == nil {
		return sqlc.CreateCustomerRow{}, ErrInvalidCustomerPassword
	}

	return s.repository.Create(
		ctx,
		fullName,
		email,
		phone,
		passwordHash,
	)
}

// GetByID mengambil customer berdasarkan ID.
func (s *CustomerService) GetByID(
	ctx context.Context,
	id int64,
) (sqlc.GetCustomerByIDRow, error) {
	if id < 1 {
		return sqlc.GetCustomerByIDRow{}, ErrInvalidCustomerID
	}

	return s.repository.GetByID(ctx, id)
}

// UpdateProfile memperbarui profil customer.
func (s *CustomerService) UpdateProfile(
	ctx context.Context,
	id int64,
	fullName string,
	email *string,
	phone *string,
) (sqlc.UpdateCustomerProfileRow, error) {
	if id < 1 {
		return sqlc.UpdateCustomerProfileRow{}, ErrInvalidCustomerID
	}

	fullName = strings.TrimSpace(fullName)
	email = normalizeOptionalString(email)
	phone = normalizeOptionalString(phone)

	if fullName == "" {
		return sqlc.UpdateCustomerProfileRow{}, ErrInvalidCustomerName
	}

	if email == nil && phone == nil {
		return sqlc.UpdateCustomerProfileRow{}, ErrInvalidCustomerContact
	}

	if email != nil {
		*email = strings.ToLower(*email)
	}

	return s.repository.UpdateProfile(
		ctx,
		id,
		fullName,
		email,
		phone,
	)
}

// UpdatePassword memperbarui password hash customer.
func (s *CustomerService) UpdatePassword(
	ctx context.Context,
	id int64,
	passwordHash *string,
) (int64, error) {
	if id < 1 {
		return 0, ErrInvalidCustomerID
	}

	passwordHash = normalizeOptionalString(passwordHash)

	if passwordHash == nil {
		return 0, ErrInvalidCustomerPassword
	}

	return s.repository.UpdatePassword(
		ctx,
		id,
		passwordHash,
	)
}

// Activate mengaktifkan kembali akun customer.
func (s *CustomerService) Activate(
	ctx context.Context,
	id int64,
) (int64, error) {
	if id < 1 {
		return 0, ErrInvalidCustomerID
	}

	return s.repository.Activate(ctx, id)
}

// Deactivate menonaktifkan akun customer.
func (s *CustomerService) Deactivate(
	ctx context.Context,
	id int64,
) (int64, error) {
	if id < 1 {
		return 0, ErrInvalidCustomerID
	}

	return s.repository.Deactivate(ctx, id)
}

// List mengambil daftar customer dengan pagination.
func (s *CustomerService) List(
	ctx context.Context,
	limit int32,
	offset int32,
) ([]sqlc.ListCustomersRow, error) {
	if limit < 1 || limit > 100 {
		limit = 20
	}

	if offset < 0 {
		offset = 0
	}

	return s.repository.List(ctx, limit, offset)
}

// Search mencari customer berdasarkan keyword.
func (s *CustomerService) Search(
	ctx context.Context,
	keyword string,
	limit int32,
	offset int32,
) ([]sqlc.SearchCustomersRow, error) {
	keyword = strings.TrimSpace(keyword)

	if limit < 1 || limit > 100 {
		limit = 20
	}

	if offset < 0 {
		offset = 0
	}

	return s.repository.Search(
		ctx,
		keyword,
		limit,
		offset,
	)
}

// Count menghitung jumlah customer.
func (s *CustomerService) Count(
	ctx context.Context,
) (int64, error) {
	return s.repository.Count(ctx)
}

// normalizeOptionalString membersihkan string opsional.
func normalizeOptionalString(value *string) *string {
	if value == nil {
		return nil
	}

	normalized := strings.TrimSpace(*value)

	if normalized == "" {
		return nil
	}

	return &normalized
}
