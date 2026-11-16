package repository

import (
	"context"
	"errors"

	dbsqlc "hafiz-pintar-store/backend/api/internal/database/sqlc"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrPaymentAlreadyExists = errors.New("payment already exists")

type PaymentRepository struct {
	queries *dbsqlc.Queries
	pool    *pgxpool.Pool
}

func NewPaymentRepository(
	queries *dbsqlc.Queries,
	pool *pgxpool.Pool,
) *PaymentRepository {
	return &PaymentRepository{
		queries: queries,
		pool:    pool,
	}
}

// CREATE PAYMENT
func (r *PaymentRepository) Create(
	ctx context.Context,
	params dbsqlc.CreatePaymentParams,
) (dbsqlc.Payment, error) {
	payment, err := r.queries.CreatePayment(ctx, params)
	if err != nil {
		var pgErr *pgconn.PgError

		if errors.As(err, &pgErr) &&
			pgErr.Code == "23505" &&
			pgErr.ConstraintName == "payments_order_id_unique" {
			return dbsqlc.Payment{}, ErrPaymentAlreadyExists
		}

		return dbsqlc.Payment{}, err
	}

	return payment, nil
}

// GET PAYMENT BY ID
func (r *PaymentRepository) GetByID(
	ctx context.Context,
	id int64,
) (dbsqlc.Payment, error) {
	return r.queries.GetPaymentByID(ctx, id)
}

// GET PAYMENT BY ORDER ID
func (r *PaymentRepository) GetByOrderID(
	ctx context.Context,
	orderID int64,
) (dbsqlc.Payment, error) {
	return r.queries.GetPaymentByOrderID(ctx, orderID)
}

// GET ORDER OWNED BY CUSTOMER
func (r *PaymentRepository) GetCustomerOrder(
	ctx context.Context,
	orderID int64,
	customerID int64,
) (dbsqlc.Order, error) {
	return r.queries.GetCustomerOrderByID(
		ctx,
		dbsqlc.GetCustomerOrderByIDParams{
			ID:         orderID,
			CustomerID: customerID,
		},
	)
}

// UPDATE PAYMENT STATUS
func (r *PaymentRepository) UpdateStatus(
	ctx context.Context,
	id int64,
	status string,
) (dbsqlc.Payment, error) {
	return r.queries.UpdatePaymentStatus(
		ctx,
		dbsqlc.UpdatePaymentStatusParams{
			ID:     id,
			Status: status,
		},
	)
}

// UPDATE TRANSACTION REFERENCE
func (r *PaymentRepository) UpdateTransactionReference(
	ctx context.Context,
	id int64,
	provider *string,
	reference *string,
) (dbsqlc.Payment, error) {
	return r.queries.UpdatePaymentTransactionReference(
		ctx,
		dbsqlc.UpdatePaymentTransactionReferenceParams{
			ID:                   id,
			PaymentProvider:      provider,
			TransactionReference: reference,
		},
	)
}
