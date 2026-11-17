package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrAdminPaymentNotFound = errors.New("payment not found")
	ErrAdminPaymentConflict = errors.New("payment status transition is not allowed")
)

type AdminPayment struct {
	ID                   int64      `json:"id"`
	OrderID              int64      `json:"order_id"`
	OrderNumber          string     `json:"order_number"`
	CustomerID           int64      `json:"customer_id"`
	CustomerName         string     `json:"customer_name"`
	CustomerEmail        *string    `json:"customer_email,omitempty"`
	PaymentMethod        string     `json:"payment_method"`
	PaymentProvider      *string    `json:"payment_provider,omitempty"`
	TransactionReference *string    `json:"transaction_reference,omitempty"`
	Amount               int64      `json:"amount"`
	Status               string     `json:"status"`
	PaidAt               *time.Time `json:"paid_at,omitempty"`
	CreatedAt            time.Time  `json:"created_at"`
	UpdatedAt            time.Time  `json:"updated_at"`
}

type AdminPaymentRepository struct {
	pool *pgxpool.Pool
}

func NewAdminPaymentRepository(pool *pgxpool.Pool) *AdminPaymentRepository {
	return &AdminPaymentRepository{pool: pool}
}

const adminPaymentSelect = `
SELECT
	p.id,
	p.order_id,
	o.order_number,
	o.customer_id,
	c.full_name,
	c.email,
	p.payment_method,
	p.payment_provider,
	p.transaction_reference,
	p.amount,
	p.status,
	p.paid_at,
	p.created_at,
	p.updated_at
FROM payments p
JOIN orders o ON o.id = p.order_id
JOIN customers c ON c.id = o.customer_id
`

func scanAdminPayment(row pgx.Row) (AdminPayment, error) {
	var item AdminPayment
	var provider, reference, email pgtype.Text
	var paidAt pgtype.Timestamptz

	err := row.Scan(
		&item.ID,
		&item.OrderID,
		&item.OrderNumber,
		&item.CustomerID,
		&item.CustomerName,
		&email,
		&item.PaymentMethod,
		&provider,
		&reference,
		&item.Amount,
		&item.Status,
		&paidAt,
		&item.CreatedAt,
		&item.UpdatedAt,
	)
	if err != nil {
		return AdminPayment{}, err
	}

	if email.Valid {
		item.CustomerEmail = &email.String
	}

	if provider.Valid {
		item.PaymentProvider = &provider.String
	}

	if reference.Valid {
		item.TransactionReference = &reference.String
	}

	if paidAt.Valid {
		item.PaidAt = &paidAt.Time
	}

	return item, nil
}

func (r *AdminPaymentRepository) List(
	ctx context.Context,
	status string,
	limit, offset int32,
) ([]AdminPayment, int64, error) {

	var total int64
	var err error

	if status == "" {
		err = r.pool.QueryRow(
			ctx,
			`SELECT COUNT(*) FROM payments`,
		).Scan(&total)
	} else {
		err = r.pool.QueryRow(
			ctx,
			`SELECT COUNT(*) FROM payments WHERE status = $1`,
			status,
		).Scan(&total)
	}

	if err != nil {
		return nil, 0, err
	}

	query := adminPaymentSelect
	args := []any{}

	if status != "" {
		query += ` WHERE p.status = $1`
		args = append(args, status)

		query += fmt.Sprintf(
			` ORDER BY p.created_at DESC, p.id DESC LIMIT $%d OFFSET $%d`,
			len(args)+1,
			len(args)+2,
		)
	} else {
		query += `
			ORDER BY p.created_at DESC, p.id DESC
			LIMIT $1 OFFSET $2
		`
	}

	args = append(args, limit, offset)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	items := make([]AdminPayment, 0)

	for rows.Next() {
		item, err := scanAdminPayment(rows)
		if err != nil {
			return nil, 0, err
		}

		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	return items, total, nil
}

func (r *AdminPaymentRepository) GetByID(
	ctx context.Context,
	id int64,
) (AdminPayment, error) {

	item, err := scanAdminPayment(
		r.pool.QueryRow(
			ctx,
			adminPaymentSelect+`
				WHERE p.id = $1
				LIMIT 1
			`,
			id,
		),
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return AdminPayment{}, ErrAdminPaymentNotFound
	}

	return item, err
}

func (r *AdminPaymentRepository) UpdateStatus(
	ctx context.Context,
	id int64,
	status string,
) (AdminPayment, error) {

	// Transisi yang diizinkan:
	// pending -> paid
	// pending -> failed
	// paid -> refunded

	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return AdminPayment{}, err
	}

	defer tx.Rollback(context.Background())

	// 1. Update payment di dalam transaction.
	tag, err := tx.Exec(
		ctx,
		`
		UPDATE payments
		SET
			status = $2,
			paid_at = CASE
				WHEN $2 = 'paid'
				THEN COALESCE(paid_at, NOW())
				ELSE paid_at
			END,
			updated_at = NOW()
		WHERE id = $1
		AND (
			(status = 'pending' AND $2 IN ('paid', 'failed'))
			OR
			(status = 'paid' AND $2 = 'refunded')
		)
		`,
		id,
		status,
	)
	if err != nil {
		return AdminPayment{}, err
	}

	// 2. Pastikan payment ditemukan dan transisinya valid.
	if tag.RowsAffected() == 0 {
		var exists bool

		err := tx.QueryRow(
			ctx,
			`SELECT EXISTS(SELECT 1 FROM payments WHERE id = $1)`,
			id,
		).Scan(&exists)

		if err != nil {
			return AdminPayment{}, err
		}

		if !exists {
			return AdminPayment{}, ErrAdminPaymentNotFound
		}

		return AdminPayment{}, ErrAdminPaymentConflict
	}

	// 3. Jika payment menjadi paid, konfirmasi order.
	if status == "paid" {
		orderTag, err := tx.Exec(
			ctx,
			`
			UPDATE orders
			SET
				status = 'confirmed',
				updated_at = NOW()
			WHERE id = (
				SELECT order_id
				FROM payments
				WHERE id = $1
			)
			AND status = 'pending'
			`,
			id,
		)
		if err != nil {
			return AdminPayment{}, err
		}

		// Jika order bukan pending, batalkan seluruh transaction.
		if orderTag.RowsAffected() == 0 {
			return AdminPayment{}, ErrAdminPaymentConflict
		}
	}

	// 4. Commit payment dan order secara bersamaan.
	if err := tx.Commit(ctx); err != nil {
		return AdminPayment{}, err
	}

	// 5. Ambil data payment terbaru setelah commit.
	return r.GetByID(ctx, id)
}
