package service

import (
	"context"
	"errors"
	"strings"

	dbsqlc "hafiz-pintar-store/backend/api/internal/database/sqlc"
	"hafiz-pintar-store/backend/api/internal/repository"

	"github.com/jackc/pgx/v5"
)

var (
	ErrInvalidPaymentOrderID  = errors.New("invalid order ID")
	ErrInvalidPaymentMethod   = errors.New("invalid payment method")
	ErrPaymentOrderNotFound   = errors.New("order not found")
	ErrPaymentAlreadyExists   = errors.New("payment already exists")
	ErrPaymentOrderNotPayable = errors.New("order is not payable")
	ErrInvalidPaymentID       = errors.New("invalid payment ID")
)

type PaymentService struct {
	repo *repository.PaymentRepository
}

func NewPaymentService(
	repo *repository.PaymentRepository,
) *PaymentService {
	return &PaymentService{
		repo: repo,
	}
}

// =========================================
// CREATE PAYMENT
// =========================================

func (s *PaymentService) Create(
	ctx context.Context,
	customerID int64,
	orderID int64,
	paymentMethod string,
) (dbsqlc.Payment, error) {

	if customerID <= 0 || orderID <= 0 {
		return dbsqlc.Payment{}, ErrInvalidPaymentOrderID
	}

	paymentMethod = strings.TrimSpace(paymentMethod)

	if paymentMethod == "" {
		return dbsqlc.Payment{}, ErrInvalidPaymentMethod
	}

	// Pastikan order dimiliki customer yang login.
	order, err := s.repo.GetCustomerOrder(
		ctx,
		orderID,
		customerID,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return dbsqlc.Payment{}, ErrPaymentOrderNotFound
	}

	if err != nil {
		return dbsqlc.Payment{}, err
	}

	// Hanya order pending yang dapat dibuatkan pembayaran.
	if order.Status != "pending" {
		return dbsqlc.Payment{}, ErrPaymentOrderNotPayable
	}

	// Cek apakah pembayaran untuk order sudah tersedia.
	_, err = s.repo.GetByOrderID(ctx, orderID)

	if err == nil {
		return dbsqlc.Payment{}, ErrPaymentAlreadyExists
	}

	if !errors.Is(err, pgx.ErrNoRows) {
		return dbsqlc.Payment{}, err
	}

	// Nominal pembayaran selalu berasal dari database.
	params := dbsqlc.CreatePaymentParams{
		OrderID:       order.ID,
		PaymentMethod: paymentMethod,
		Amount:        order.GrandTotal,
	}

	// Simpan payment.
	payment, err := s.repo.Create(ctx, params)

	// Tangani duplicate payment dari unique constraint database.
	if errors.Is(err, repository.ErrPaymentAlreadyExists) {
		return dbsqlc.Payment{}, ErrPaymentAlreadyExists
	}

	if err != nil {
		return dbsqlc.Payment{}, err
	}

	return payment, nil
}

// =========================================
// GET PAYMENT BY ID
// =========================================

func (s *PaymentService) GetByID(
	ctx context.Context,
	id int64,
) (dbsqlc.Payment, error) {

	if id <= 0 {
		return dbsqlc.Payment{}, ErrInvalidPaymentID
	}

	return s.repo.GetByID(ctx, id)
}

// =========================================
// GET PAYMENT BY ID FOR CUSTOMER
// =========================================

func (s *PaymentService) GetCustomerPaymentByID(
	ctx context.Context,
	id int64,
	customerID int64,
) (dbsqlc.Payment, error) {

	if id <= 0 || customerID <= 0 {
		return dbsqlc.Payment{}, ErrInvalidPaymentID
	}

	// Repository memastikan payment merupakan milik customer.
	return s.repo.GetCustomerPaymentByID(
		ctx,
		id,
		customerID,
	)
}

// =========================================
// GET PAYMENT BY ORDER ID
// =========================================

func (s *PaymentService) GetByOrderID(
	ctx context.Context,
	orderID int64,
) (dbsqlc.Payment, error) {

	if orderID <= 0 {
		return dbsqlc.Payment{}, ErrInvalidPaymentOrderID
	}

	return s.repo.GetByOrderID(ctx, orderID)
}
