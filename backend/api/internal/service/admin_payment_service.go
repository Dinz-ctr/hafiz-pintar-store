package service

import (
	"context"
	"errors"
	"strings"

	"hafiz-pintar-store/backend/api/internal/repository"
)

var (
	ErrInvalidAdminPaymentID = errors.New(
		"invalid payment ID",
	)

	ErrInvalidAdminPaymentStatus = errors.New(
		"invalid payment status",
	)

	ErrInvalidAdminPaymentTransition = errors.New(
		"payment status transition is not allowed",
	)
)

type AdminPaymentService struct {
	repo *repository.AdminPaymentRepository
}

func NewAdminPaymentService(
	repo *repository.AdminPaymentRepository,
) *AdminPaymentService {
	return &AdminPaymentService{
		repo: repo,
	}
}

func (s *AdminPaymentService) List(
	ctx context.Context,
	status string,
	page, limit int32,
) ([]repository.AdminPayment, int64, error) {

	status = strings.TrimSpace(
		strings.ToLower(status),
	)

	switch status {
	case "", "pending", "paid", "failed", "refunded":
	default:
		return nil, 0, ErrInvalidAdminPaymentStatus
	}

	if page < 1 {
		page = 1
	}

	if limit < 1 {
		limit = 10
	}

	if limit > 100 {
		limit = 100
	}

	offset := (page - 1) * limit

	return s.repo.List(
		ctx,
		status,
		limit,
		offset,
	)
}

func (s *AdminPaymentService) GetByID(
	ctx context.Context,
	id int64,
) (repository.AdminPayment, error) {

	if id <= 0 {
		return repository.AdminPayment{}, ErrInvalidAdminPaymentID
	}

	return s.repo.GetByID(ctx, id)
}

func (s *AdminPaymentService) UpdateStatus(
	ctx context.Context,
	id int64,
	status string,
) (repository.AdminPayment, error) {

	if id <= 0 {
		return repository.AdminPayment{}, ErrInvalidAdminPaymentID
	}

	status = strings.TrimSpace(
		strings.ToLower(status),
	)

	switch status {
	case "paid", "failed", "refunded":
	default:
		return repository.AdminPayment{}, ErrInvalidAdminPaymentStatus
	}

	item, err := s.repo.UpdateStatus(
		ctx,
		id,
		status,
	)

	if errors.Is(err, repository.ErrAdminPaymentConflict) {
		return repository.AdminPayment{}, ErrInvalidAdminPaymentTransition
	}

	return item, err
}
