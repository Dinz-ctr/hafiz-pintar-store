package repository

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type AdminDashboardRepository struct {
	pool *pgxpool.Pool
}

func NewAdminDashboardRepository(
	pool *pgxpool.Pool,
) *AdminDashboardRepository {
	return &AdminDashboardRepository{
		pool: pool,
	}
}

type AdminDashboardSummary struct {
	TotalProducts   int64 `json:"total_products"`
	TotalCategories int64 `json:"total_categories"`
	TotalCustomers  int64 `json:"total_customers"`
	TotalOrders     int64 `json:"total_orders"`
	PendingOrders   int64 `json:"pending_orders"`
	TotalRevenue    int64 `json:"total_revenue"`
}

type AdminRecentOrder struct {
	ID           int64     `json:"id"`
	OrderNumber  string    `json:"order_number"`
	CustomerName string    `json:"customer_name"`
	Status       string    `json:"status"`
	GrandTotal   int64     `json:"grand_total"`
	CreatedAt    time.Time `json:"created_at"`
}

// GET DASHBOARD SUMMARY
func (r *AdminDashboardRepository) GetSummary(
	ctx context.Context,
) (AdminDashboardSummary, error) {
	var summary AdminDashboardSummary

	err := r.pool.QueryRow(ctx, `
		SELECT
			(SELECT COUNT(*) FROM products),
			(SELECT COUNT(*) FROM categories WHERE is_active = TRUE),
			(SELECT COUNT(*) FROM customers WHERE is_active = TRUE),
			(SELECT COUNT(*) FROM orders),
			(SELECT COUNT(*) FROM orders WHERE status = 'pending'),
			(
				SELECT COALESCE(SUM(amount), 0)
				FROM payments
				WHERE status = 'paid'
			)
	`).Scan(
		&summary.TotalProducts,
		&summary.TotalCategories,
		&summary.TotalCustomers,
		&summary.TotalOrders,
		&summary.PendingOrders,
		&summary.TotalRevenue,
	)

	if err != nil {
		return AdminDashboardSummary{}, err
	}

	return summary, nil
}

// GET FIVE MOST RECENT ORDERS
func (r *AdminDashboardRepository) GetRecentOrders(
	ctx context.Context,
	limit int32,
) ([]AdminRecentOrder, error) {
	if limit <= 0 || limit > 20 {
		limit = 5
	}

	rows, err := r.pool.Query(ctx, `
		SELECT
			o.id,
			o.order_number,
			c.full_name,
			o.status,
			o.grand_total,
			o.created_at
		FROM orders o
		JOIN customers c ON c.id = o.customer_id
		ORDER BY o.created_at DESC, o.id DESC
		LIMIT $1
	`, limit)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	orders := make([]AdminRecentOrder, 0)

	for rows.Next() {
		var order AdminRecentOrder

		if err := rows.Scan(
			&order.ID,
			&order.OrderNumber,
			&order.CustomerName,
			&order.Status,
			&order.GrandTotal,
			&order.CreatedAt,
		); err != nil {
			return nil, err
		}

		orders = append(orders, order)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return orders, nil
}
