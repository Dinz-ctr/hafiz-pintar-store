package service

import (
	"context"
	"errors"
	"strings"

	"hafiz-pintar-store/backend/api/internal/repository"

	"github.com/jackc/pgx/v5"
)

// =========================================
// ORDER ERRORS
// =========================================

var (
	ErrInvalidCustomer   = errors.New("invalid customer")
	ErrInvalidOrderID    = errors.New("invalid order ID")
	ErrOrderNotFound     = errors.New("order not found")
	ErrInvalidPagination = errors.New("invalid pagination")
)

// =========================================
// ORDER SERVICE
// =========================================

type OrderService struct {
	orderRepo *repository.OrderRepository
}

func NewOrderService(
	orderRepo *repository.OrderRepository,
) *OrderService {
	return &OrderService{
		orderRepo: orderRepo,
	}
}

// =========================================
// CHECKOUT
// =========================================

func (s *OrderService) Checkout(
	ctx context.Context,
	customerID int64,
	addressID int64,
	customerNote string,
) (repository.CheckoutResult, error) {

	if customerID <= 0 {
		return repository.CheckoutResult{}, ErrInvalidCustomer
	}

	if addressID <= 0 {
		return repository.CheckoutResult{}, repository.ErrCheckoutAddress
	}

	customerNote = strings.TrimSpace(customerNote)

	return s.orderRepo.Checkout(
		ctx,
		customerID,
		addressID,
		customerNote,
	)
}

// =========================================
// GET CUSTOMER ORDER HISTORY
// =========================================

func (s *OrderService) GetCustomerOrders(
	ctx context.Context,
	customerID int64,
	page int32,
	limit int32,
) ([]repository.OrderSummary, error) {

	if customerID <= 0 {
		return nil, ErrInvalidCustomer
	}

	if page <= 0 {
		page = 1
	}

	if limit <= 0 {
		limit = 10
	}

	// Batasi jumlah data per halaman.
	if limit > 100 {
		limit = 100
	}

	offset := (page - 1) * limit

	orders, err := s.orderRepo.GetCustomerOrders(
		ctx,
		customerID,
		limit,
		offset,
	)

	if err != nil {
		return nil, err
	}

	return orders, nil
}

// =========================================
// GET CUSTOMER ORDER DETAIL
// =========================================

func (s *OrderService) GetCustomerOrderDetail(
	ctx context.Context,
	customerID int64,
	orderID int64,
) (repository.OrderDetail, error) {

	if customerID <= 0 {
		return repository.OrderDetail{}, ErrInvalidCustomer
	}

	if orderID <= 0 {
		return repository.OrderDetail{}, ErrInvalidOrderID
	}

	order, err := s.orderRepo.GetCustomerOrderDetail(
		ctx,
		customerID,
		orderID,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return repository.OrderDetail{}, ErrOrderNotFound
	}

	if err != nil {
		return repository.OrderDetail{}, err
	}

	return order, nil
}
