package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrAdminCustomerNotFound = errors.New("customer not found")

type AdminCustomerRepository struct {
	pool *pgxpool.Pool
}

func NewAdminCustomerRepository(pool *pgxpool.Pool) *AdminCustomerRepository {
	return &AdminCustomerRepository{pool: pool}
}

type AdminCustomer struct {
	ID           int64     `json:"id"`
	FullName     string    `json:"full_name"`
	Email        *string   `json:"email"`
	Phone        *string   `json:"phone"`
	IsActive     bool      `json:"is_active"`
	AddressCount int64     `json:"address_count,omitempty"`
	OrderCount   int64     `json:"order_count,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type AdminCustomerAddress struct {
	ID             int64     `json:"id"`
	CustomerID     int64     `json:"customer_id"`
	Label          string    `json:"label"`
	RecipientName  string    `json:"recipient_name"`
	RecipientPhone string    `json:"recipient_phone"`
	AddressLine    string    `json:"address_line"`
	Subdistrict    *string   `json:"subdistrict"`
	District       *string   `json:"district"`
	City           string    `json:"city"`
	Province       string    `json:"province"`
	PostalCode     *string   `json:"postal_code"`
	IsDefault      bool      `json:"is_default"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type AdminCustomerOrder struct {
	ID                 int64     `json:"id"`
	OrderNumber        string    `json:"order_number"`
	CustomerID         int64     `json:"customer_id"`
	Status             string    `json:"status"`
	Subtotal           int64     `json:"subtotal"`
	ShippingCost       int64     `json:"shipping_cost"`
	DiscountAmount     int64     `json:"discount_amount"`
	GrandTotal         int64     `json:"grand_total"`
	RecipientName      string    `json:"recipient_name"`
	RecipientPhone     string    `json:"recipient_phone"`
	ShippingAddress    string    `json:"shipping_address"`
	ShippingCity       string    `json:"shipping_city"`
	ShippingProvince   string    `json:"shipping_province"`
	ShippingPostalCode *string   `json:"shipping_postal_code"`
	CustomerNote       *string   `json:"customer_note"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}

func (r *AdminCustomerRepository) List(
	ctx context.Context,
	keyword string,
	limit int32,
	offset int32,
) ([]AdminCustomer, int64, error) {
	keyword = strings.TrimSpace(keyword)
	pattern := "%" + keyword + "%"

	var total int64
	err := r.pool.QueryRow(ctx, `
		SELECT COUNT(*)
		FROM customers
		WHERE (
			$1 = ''
			OR full_name ILIKE $2
			OR COALESCE(email, '') ILIKE $2
			OR COALESCE(phone, '') ILIKE $2
		)
	`, keyword, pattern).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	rows, err := r.pool.Query(ctx, `
		SELECT
			id,
			full_name,
			email,
			phone,
			is_active,
			created_at,
			updated_at
		FROM customers
		WHERE (
			$1 = ''
			OR full_name ILIKE $2
			OR COALESCE(email, '') ILIKE $2
			OR COALESCE(phone, '') ILIKE $2
		)
		ORDER BY created_at DESC, id DESC
		LIMIT $3 OFFSET $4
	`, keyword, pattern, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	customers := make([]AdminCustomer, 0)

	for rows.Next() {
		var customer AdminCustomer

		if err := rows.Scan(
			&customer.ID,
			&customer.FullName,
			&customer.Email,
			&customer.Phone,
			&customer.IsActive,
			&customer.CreatedAt,
			&customer.UpdatedAt,
		); err != nil {
			return nil, 0, err
		}

		customers = append(customers, customer)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	return customers, total, nil
}

func (r *AdminCustomerRepository) GetByID(
	ctx context.Context,
	id int64,
) (AdminCustomer, error) {
	var customer AdminCustomer

	err := r.pool.QueryRow(ctx, `
		SELECT
			c.id,
			c.full_name,
			c.email,
			c.phone,
			c.is_active,
			c.created_at,
			c.updated_at,
			(
				SELECT COUNT(*)
				FROM customer_addresses ca
				WHERE ca.customer_id = c.id
			) AS address_count,
			(
				SELECT COUNT(*)
				FROM orders o
				WHERE o.customer_id = c.id
			) AS order_count
		FROM customers c
		WHERE c.id = $1
	`, id).Scan(
		&customer.ID,
		&customer.FullName,
		&customer.Email,
		&customer.Phone,
		&customer.IsActive,
		&customer.CreatedAt,
		&customer.UpdatedAt,
		&customer.AddressCount,
		&customer.OrderCount,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return AdminCustomer{}, ErrAdminCustomerNotFound
	}
	if err != nil {
		return AdminCustomer{}, err
	}

	return customer, nil
}

func (r *AdminCustomerRepository) UpdateStatus(
	ctx context.Context,
	id int64,
	isActive bool,
) (AdminCustomer, error) {
	var customer AdminCustomer

	err := r.pool.QueryRow(ctx, `
		UPDATE customers
		SET
			is_active = $2,
			updated_at = NOW()
		WHERE id = $1
		RETURNING
			id,
			full_name,
			email,
			phone,
			is_active,
			created_at,
			updated_at
	`, id, isActive).Scan(
		&customer.ID,
		&customer.FullName,
		&customer.Email,
		&customer.Phone,
		&customer.IsActive,
		&customer.CreatedAt,
		&customer.UpdatedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return AdminCustomer{}, ErrAdminCustomerNotFound
	}
	if err != nil {
		return AdminCustomer{}, err
	}

	return customer, nil
}

func (r *AdminCustomerRepository) ListAddresses(
	ctx context.Context,
	customerID int64,
) ([]AdminCustomerAddress, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT
			id,
			customer_id,
			label,
			recipient_name,
			recipient_phone,
			address_line,
			subdistrict,
			district,
			city,
			province,
			postal_code,
			is_default,
			created_at,
			updated_at
		FROM customer_addresses
		WHERE customer_id = $1
		ORDER BY is_default DESC, created_at DESC, id DESC
	`, customerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	addresses := make([]AdminCustomerAddress, 0)

	for rows.Next() {
		var address AdminCustomerAddress
		var subdistrict, district, postalCode pgtype.Text

		if err := rows.Scan(
			&address.ID,
			&address.CustomerID,
			&address.Label,
			&address.RecipientName,
			&address.RecipientPhone,
			&address.AddressLine,
			&subdistrict,
			&district,
			&address.City,
			&address.Province,
			&postalCode,
			&address.IsDefault,
			&address.CreatedAt,
			&address.UpdatedAt,
		); err != nil {
			return nil, err
		}

		if subdistrict.Valid {
			address.Subdistrict = &subdistrict.String
		}
		if district.Valid {
			address.District = &district.String
		}
		if postalCode.Valid {
			address.PostalCode = &postalCode.String
		}

		addresses = append(addresses, address)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return addresses, nil
}

func (r *AdminCustomerRepository) ListOrders(
	ctx context.Context,
	customerID int64,
	limit int32,
	offset int32,
) ([]AdminCustomerOrder, int64, error) {
	var total int64

	err := r.pool.QueryRow(ctx, `
		SELECT COUNT(*)
		FROM orders
		WHERE customer_id = $1
	`, customerID).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	rows, err := r.pool.Query(ctx, `
		SELECT
			id,
			order_number,
			customer_id,
			status,
			subtotal,
			shipping_cost,
			discount_amount,
			grand_total,
			recipient_name,
			recipient_phone,
			shipping_address,
			shipping_city,
			shipping_province,
			shipping_postal_code,
			customer_note,
			created_at,
			updated_at
		FROM orders
		WHERE customer_id = $1
		ORDER BY created_at DESC, id DESC
		LIMIT $2 OFFSET $3
	`, customerID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	orders := make([]AdminCustomerOrder, 0)

	for rows.Next() {
		var order AdminCustomerOrder
		var postalCode, customerNote pgtype.Text

		if err := rows.Scan(
			&order.ID,
			&order.OrderNumber,
			&order.CustomerID,
			&order.Status,
			&order.Subtotal,
			&order.ShippingCost,
			&order.DiscountAmount,
			&order.GrandTotal,
			&order.RecipientName,
			&order.RecipientPhone,
			&order.ShippingAddress,
			&order.ShippingCity,
			&order.ShippingProvince,
			&postalCode,
			&customerNote,
			&order.CreatedAt,
			&order.UpdatedAt,
		); err != nil {
			return nil, 0, err
		}

		if postalCode.Valid {
			order.ShippingPostalCode = &postalCode.String
		}
		if customerNote.Valid {
			order.CustomerNote = &customerNote.String
		}

		orders = append(orders, order)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	return orders, total, nil
}

func ValidateAdminCustomerPagination(page, limit int32) (int32, int32) {
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

	return limit, offset
}

func AdminCustomerPaginationMeta(page, limit int32, total int64) map[string]any {
	totalPages := int64(0)
	if limit > 0 {
		totalPages = (total + int64(limit) - 1) / int64(limit)
	}

	return map[string]any{
		"page":        page,
		"limit":       limit,
		"total":       total,
		"total_pages": totalPages,
	}
}

func adminCustomerOffset(page, limit int32) int32 {
	_, offset := ValidateAdminCustomerPagination(page, limit)
	return offset
}

var _ = fmt.Sprintf
