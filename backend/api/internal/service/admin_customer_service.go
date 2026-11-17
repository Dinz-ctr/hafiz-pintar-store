package service

import (
	"context"
	"errors"
	"strings"

	"hafiz-pintar-store/backend/api/internal/repository"
)

var ErrInvalidAdminCustomerID = errors.New("customer ID must be greater than zero")

type AdminCustomerService struct {
	repository *repository.AdminCustomerRepository
}

func NewAdminCustomerService(
	repo *repository.AdminCustomerRepository,
) *AdminCustomerService {
	return &AdminCustomerService{repository: repo}
}

func (s *AdminCustomerService) List(
	ctx context.Context,
	keyword string,
	page int32,
	limit int32,
) ([]repository.AdminCustomer, int64, int32, int32, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	offset := (page - 1) * limit
	if offset < 0 {
		offset = 0
	}

	customers, total, err := s.repository.List(
		ctx,
		strings.TrimSpace(keyword),
		limit,
		offset,
	)
	return customers, total, page, limit, err
}

func (s *AdminCustomerService) GetByID(
	ctx context.Context,
	id int64,
) (repository.AdminCustomer, error) {
	if id < 1 {
		return repository.AdminCustomer{}, ErrInvalidAdminCustomerID
	}

	return s.repository.GetByID(ctx, id)
}

func (s *AdminCustomerService) UpdateStatus(
	ctx context.Context,
	id int64,
	isActive bool,
) (repository.AdminCustomer, error) {
	if id < 1 {
		return repository.AdminCustomer{}, ErrInvalidAdminCustomerID
	}

	return s.repository.UpdateStatus(ctx, id, isActive)
}

func (s *AdminCustomerService) ListAddresses(
	ctx context.Context,
	id int64,
) ([]repository.AdminCustomerAddress, error) {
	if id < 1 {
		return nil, ErrInvalidAdminCustomerID
	}

	if _, err := s.repository.GetByID(ctx, id); err != nil {
		return nil, err
	}

	return s.repository.ListAddresses(ctx, id)
}

func (s *AdminCustomerService) ListOrders(
	ctx context.Context,
	id int64,
	page int32,
	limit int32,
) ([]repository.AdminCustomerOrder, int64, int32, int32, error) {
	if id < 1 {
		return nil, 0, 0, 0, ErrInvalidAdminCustomerID
	}

	if _, err := s.repository.GetByID(ctx, id); err != nil {
		return nil, 0, 0, 0, err
	}

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	offset := (page - 1) * limit
	if offset < 0 {
		offset = 0
	}

	orders, total, err := s.repository.ListOrders(
		ctx,
		id,
		limit,
		offset,
	)

	return orders, total, page, limit, err
}
