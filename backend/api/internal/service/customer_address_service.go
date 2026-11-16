package service

import (
	"context"
	"errors"
	"strings"

	"hafiz-pintar-store/backend/api/internal/database/sqlc"
	"hafiz-pintar-store/backend/api/internal/repository"
)

var (
	ErrInvalidAddressCustomerID = errors.New("customer ID must be greater than zero")
	ErrInvalidAddressID         = errors.New("address ID must be greater than zero")
	ErrInvalidAddressLabel      = errors.New("address label is required")
	ErrInvalidRecipientName     = errors.New("recipient name is required")
	ErrInvalidRecipientPhone    = errors.New("recipient phone is required")
	ErrInvalidAddressLine       = errors.New("address line is required")
	ErrInvalidAddressCity       = errors.New("address city is required")
	ErrInvalidAddressProvince   = errors.New("address province is required")
)

type CustomerAddressService struct {
	repository *repository.CustomerAddressRepository
}

func NewCustomerAddressService(
	repo *repository.CustomerAddressRepository,
) *CustomerAddressService {
	return &CustomerAddressService{
		repository: repo,
	}
}

// List mengambil seluruh alamat milik customer.
func (s *CustomerAddressService) List(
	ctx context.Context,
	customerID int64,
) ([]sqlc.CustomerAddress, error) {
	if customerID < 1 {
		return nil, ErrInvalidAddressCustomerID
	}

	return s.repository.List(ctx, customerID)
}

// GetByID mengambil alamat berdasarkan ID dan customer.
func (s *CustomerAddressService) GetByID(
	ctx context.Context,
	id int64,
	customerID int64,
) (sqlc.CustomerAddress, error) {
	if id < 1 {
		return sqlc.CustomerAddress{}, ErrInvalidAddressID
	}

	if customerID < 1 {
		return sqlc.CustomerAddress{}, ErrInvalidAddressCustomerID
	}

	return s.repository.GetByID(ctx, id, customerID)
}

// GetDefault mengambil alamat utama customer.
func (s *CustomerAddressService) GetDefault(
	ctx context.Context,
	customerID int64,
) (sqlc.CustomerAddress, error) {
	if customerID < 1 {
		return sqlc.CustomerAddress{}, ErrInvalidAddressCustomerID
	}

	return s.repository.GetDefault(ctx, customerID)
}

// Create menambahkan alamat baru untuk customer.
func (s *CustomerAddressService) Create(
	ctx context.Context,
	input sqlc.CreateCustomerAddressParams,
) (sqlc.CustomerAddress, error) {
	if input.CustomerID < 1 {
		return sqlc.CustomerAddress{}, ErrInvalidAddressCustomerID
	}

	input.Label = strings.TrimSpace(input.Label)
	input.RecipientName = strings.TrimSpace(input.RecipientName)
	input.RecipientPhone = strings.TrimSpace(input.RecipientPhone)
	input.AddressLine = strings.TrimSpace(input.AddressLine)
	input.City = strings.TrimSpace(input.City)
	input.Province = strings.TrimSpace(input.Province)

	input.Subdistrict = normalizeOptionalString(input.Subdistrict)
	input.District = normalizeOptionalString(input.District)
	input.PostalCode = normalizeOptionalString(input.PostalCode)

	if input.Label == "" {
		return sqlc.CustomerAddress{}, ErrInvalidAddressLabel
	}

	if input.RecipientName == "" {
		return sqlc.CustomerAddress{}, ErrInvalidRecipientName
	}

	if input.RecipientPhone == "" {
		return sqlc.CustomerAddress{}, ErrInvalidRecipientPhone
	}

	if input.AddressLine == "" {
		return sqlc.CustomerAddress{}, ErrInvalidAddressLine
	}

	if input.City == "" {
		return sqlc.CustomerAddress{}, ErrInvalidAddressCity
	}

	if input.Province == "" {
		return sqlc.CustomerAddress{}, ErrInvalidAddressProvince
	}

	return s.repository.Create(ctx, input)
}

// Update memperbarui alamat milik customer.
func (s *CustomerAddressService) Update(
	ctx context.Context,
	input sqlc.UpdateCustomerAddressParams,
) (sqlc.CustomerAddress, error) {
	if input.ID < 1 {
		return sqlc.CustomerAddress{}, ErrInvalidAddressID
	}

	if input.CustomerID < 1 {
		return sqlc.CustomerAddress{}, ErrInvalidAddressCustomerID
	}

	input.Label = strings.TrimSpace(input.Label)
	input.RecipientName = strings.TrimSpace(input.RecipientName)
	input.RecipientPhone = strings.TrimSpace(input.RecipientPhone)
	input.AddressLine = strings.TrimSpace(input.AddressLine)
	input.City = strings.TrimSpace(input.City)
	input.Province = strings.TrimSpace(input.Province)

	input.Subdistrict = normalizeOptionalString(input.Subdistrict)
	input.District = normalizeOptionalString(input.District)
	input.PostalCode = normalizeOptionalString(input.PostalCode)

	if input.Label == "" {
		return sqlc.CustomerAddress{}, ErrInvalidAddressLabel
	}

	if input.RecipientName == "" {
		return sqlc.CustomerAddress{}, ErrInvalidRecipientName
	}

	if input.RecipientPhone == "" {
		return sqlc.CustomerAddress{}, ErrInvalidRecipientPhone
	}

	if input.AddressLine == "" {
		return sqlc.CustomerAddress{}, ErrInvalidAddressLine
	}

	if input.City == "" {
		return sqlc.CustomerAddress{}, ErrInvalidAddressCity
	}

	if input.Province == "" {
		return sqlc.CustomerAddress{}, ErrInvalidAddressProvince
	}

	return s.repository.Update(ctx, input)
}

// Delete menghapus alamat milik customer.
func (s *CustomerAddressService) Delete(
	ctx context.Context,
	id int64,
	customerID int64,
) (int64, error) {
	if id < 1 {
		return 0, ErrInvalidAddressID
	}

	if customerID < 1 {
		return 0, ErrInvalidAddressCustomerID
	}

	affected, err := s.repository.Delete(ctx, id, customerID)
	if err != nil {
		return 0, err
	}

	if affected == 0 {
		return 0, repository.ErrCustomerAddressNotFound
	}

	return affected, nil
}

// SetDefault menjadikan alamat sebagai alamat utama customer.
func (s *CustomerAddressService) SetDefault(
	ctx context.Context,
	id int64,
	customerID int64,
) error {
	if id < 1 {
		return ErrInvalidAddressID
	}

	if customerID < 1 {
		return ErrInvalidAddressCustomerID
	}

	return s.repository.SetDefault(ctx, id, customerID)
}

// Count menghitung jumlah alamat customer.
func (s *CustomerAddressService) Count(
	ctx context.Context,
	customerID int64,
) (int64, error) {
	if customerID < 1 {
		return 0, ErrInvalidAddressCustomerID
	}

	return s.repository.Count(ctx, customerID)
}
