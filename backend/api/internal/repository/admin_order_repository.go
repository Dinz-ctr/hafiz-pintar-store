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

var (
	ErrAdminOrderNotFound     = errors.New("admin order not found")
	ErrAdminOrderConflict     = errors.New("admin order status changed concurrently")
	ErrAdminOrderStockRestore = errors.New("unable to restore order stock")
)

type AdminOrderRepository struct {
	pool *pgxpool.Pool
}

func NewAdminOrderRepository(pool *pgxpool.Pool) *AdminOrderRepository {
	return &AdminOrderRepository{pool: pool}
}

type AdminOrderSummary struct {
	ID             int64     `json:"id"`
	OrderNumber    string    `json:"order_number"`
	CustomerID     int64     `json:"customer_id"`
	CustomerName   string    `json:"customer_name"`
	CustomerEmail  string    `json:"customer_email"`
	CustomerPhone  string    `json:"customer_phone"`
	Status         string    `json:"status"`
	Subtotal       int64     `json:"subtotal"`
	ShippingCost   int64     `json:"shipping_cost"`
	DiscountAmount int64     `json:"discount_amount"`
	GrandTotal     int64     `json:"grand_total"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type AdminOrderItem struct {
	ID          int64  `json:"id"`
	ProductID   int64  `json:"product_id"`
	ProductSKU  string `json:"product_sku"`
	ProductName string `json:"product_name"`
	Quantity    int32  `json:"quantity"`
	UnitPrice   int64  `json:"unit_price"`
	Subtotal    int64  `json:"subtotal"`
}

type AdminOrderDetail struct {
	AdminOrderSummary
	RecipientName      string           `json:"recipient_name"`
	RecipientPhone     string           `json:"recipient_phone"`
	ShippingAddress    string           `json:"shipping_address"`
	ShippingCity       string           `json:"shipping_city"`
	ShippingProvince   string           `json:"shipping_province"`
	ShippingPostalCode string           `json:"shipping_postal_code"`
	CourierName        *string          `json:"courier_name"`
	TrackingNumber     *string          `json:"tracking_number"`
	ShippedAt          *time.Time       `json:"shipped_at"`
	CompletedAt        *time.Time       `json:"completed_at"`
	CustomerNote       string           `json:"customer_note"`
	Items              []AdminOrderItem `json:"items"`
}

type AdminOrderListParams struct {
	Keyword string
	Status  string
	Limit   int32
	Offset  int32
}

func (r *AdminOrderRepository) List(
	ctx context.Context,
	params AdminOrderListParams,
) ([]AdminOrderSummary, int64, error) {
	orders := make([]AdminOrderSummary, 0)

	conditions := []string{"1=1"}
	args := make([]any, 0, 4)
	argPosition := 1

	if params.Status != "" {
		conditions = append(
			conditions,
			fmt.Sprintf("o.status = $%d", argPosition),
		)
		args = append(args, params.Status)
		argPosition++
	}

	if params.Keyword != "" {
		conditions = append(
			conditions,
			fmt.Sprintf(
				`(
					o.order_number ILIKE $%d
					OR c.full_name ILIKE $%d
					OR COALESCE(c.email, '') ILIKE $%d
					OR COALESCE(c.phone, '') ILIKE $%d
				)`,
				argPosition,
				argPosition,
				argPosition,
				argPosition,
			),
		)
		args = append(args, "%"+params.Keyword+"%")
		argPosition++
	}

	whereClause := strings.Join(conditions, " AND ")

	countQuery := `
		SELECT COUNT(*)
		FROM orders o
		JOIN customers c ON c.id = o.customer_id
		WHERE ` + whereClause

	var total int64

	if err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	listQuery := `
		SELECT
			o.id,
			o.order_number,
			o.customer_id,
			c.full_name,
			COALESCE(c.email, ''),
			COALESCE(c.phone, ''),
			o.status,
			o.subtotal,
			o.shipping_cost,
			o.discount_amount,
			o.grand_total,
			o.created_at,
			o.updated_at
		FROM orders o
		JOIN customers c ON c.id = o.customer_id
		WHERE ` + whereClause +
		fmt.Sprintf(
			" ORDER BY o.created_at DESC, o.id DESC LIMIT $%d OFFSET $%d",
			argPosition,
			argPosition+1,
		)

	args = append(args, params.Limit, params.Offset)

	rows, err := r.pool.Query(ctx, listQuery, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	for rows.Next() {
		var order AdminOrderSummary

		if err := rows.Scan(
			&order.ID,
			&order.OrderNumber,
			&order.CustomerID,
			&order.CustomerName,
			&order.CustomerEmail,
			&order.CustomerPhone,
			&order.Status,
			&order.Subtotal,
			&order.ShippingCost,
			&order.DiscountAmount,
			&order.GrandTotal,
			&order.CreatedAt,
			&order.UpdatedAt,
		); err != nil {
			return nil, 0, err
		}

		orders = append(orders, order)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	return orders, total, nil
}

func (r *AdminOrderRepository) GetByID(
	ctx context.Context,
	orderID int64,
) (AdminOrderDetail, error) {
	var order AdminOrderDetail

	err := r.pool.QueryRow(ctx, `
		SELECT
			o.id,
			o.order_number,
			o.customer_id,
			c.full_name,
			COALESCE(c.email, ''),
			COALESCE(c.phone, ''),
			o.status,
			o.subtotal,
			o.shipping_cost,
			o.discount_amount,
			o.grand_total,
			o.created_at,
			o.updated_at,
			o.recipient_name,
			o.recipient_phone,
			o.shipping_address,
			o.shipping_city,
			o.shipping_province,
			COALESCE(o.shipping_postal_code, ''),
			o.courier_name,
			o.tracking_number,
			o.shipped_at,
			o.completed_at,
			COALESCE(o.customer_note, '')
		FROM orders o
		JOIN customers c ON c.id = o.customer_id
		WHERE o.id = $1
	`, orderID).Scan(
		&order.ID,
		&order.OrderNumber,
		&order.CustomerID,
		&order.CustomerName,
		&order.CustomerEmail,
		&order.CustomerPhone,
		&order.Status,
		&order.Subtotal,
		&order.ShippingCost,
		&order.DiscountAmount,
		&order.GrandTotal,
		&order.CreatedAt,
		&order.UpdatedAt,
		&order.RecipientName,
		&order.RecipientPhone,
		&order.ShippingAddress,
		&order.ShippingCity,
		&order.ShippingProvince,
		&order.ShippingPostalCode,
		&order.CourierName,
		&order.TrackingNumber,
		&order.ShippedAt,
		&order.CompletedAt,
		&order.CustomerNote,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return AdminOrderDetail{}, ErrAdminOrderNotFound
	}
	if err != nil {
		return AdminOrderDetail{}, err
	}

	order.Items = make([]AdminOrderItem, 0)

	rows, err := r.pool.Query(ctx, `
		SELECT
			id,
			product_id,
			product_sku,
			product_name,
			quantity,
			unit_price,
			subtotal
		FROM order_items
		WHERE order_id = $1
		ORDER BY id ASC
	`, orderID)

	if err != nil {
		return AdminOrderDetail{}, err
	}
	defer rows.Close()

	for rows.Next() {
		var item AdminOrderItem

		if err := rows.Scan(
			&item.ID,
			&item.ProductID,
			&item.ProductSKU,
			&item.ProductName,
			&item.Quantity,
			&item.UnitPrice,
			&item.Subtotal,
		); err != nil {
			return AdminOrderDetail{}, err
		}

		order.Items = append(order.Items, item)
	}

	if err := rows.Err(); err != nil {
		return AdminOrderDetail{}, err
	}

	return order, nil
}

func (r *AdminOrderRepository) GetStatus(
	ctx context.Context,
	orderID int64,
) (string, error) {
	var status string

	err := r.pool.QueryRow(ctx, `
		SELECT status
		FROM orders
		WHERE id = $1
	`, orderID).Scan(&status)

	if errors.Is(err, pgx.ErrNoRows) {
		return "", ErrAdminOrderNotFound
	}

	return status, err
}

func (r *AdminOrderRepository) UpdateStatus(
	ctx context.Context,
	orderID int64,
	expectedStatus string,
	newStatus string,
) (AdminOrderSummary, error) {
	var order AdminOrderSummary

	err := r.pool.QueryRow(ctx, `
		UPDATE orders o
		SET
			status = $1,
			completed_at = CASE
				WHEN $4 THEN NOW()
				ELSE o.completed_at
			END,
			updated_at = NOW()
		FROM customers c
		WHERE o.id = $2
		  AND o.status = $3
		  AND c.id = o.customer_id
		RETURNING
			o.id,
			o.order_number,
			o.customer_id,
			c.full_name,
			COALESCE(c.email, ''),
			COALESCE(c.phone, ''),
			o.status,
			o.subtotal,
			o.shipping_cost,
			o.discount_amount,
			o.grand_total,
			o.created_at,
			o.updated_at
	`,
		newStatus,
		orderID,
		expectedStatus,
		newStatus == "completed",
	).Scan(
		&order.ID,
		&order.OrderNumber,
		&order.CustomerID,
		&order.CustomerName,
		&order.CustomerEmail,
		&order.CustomerPhone,
		&order.Status,
		&order.Subtotal,
		&order.ShippingCost,
		&order.DiscountAmount,
		&order.GrandTotal,
		&order.CreatedAt,
		&order.UpdatedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		_, checkErr := r.GetStatus(ctx, orderID)

		if errors.Is(checkErr, ErrAdminOrderNotFound) {
			return AdminOrderSummary{}, ErrAdminOrderNotFound
		}
		if checkErr != nil {
			return AdminOrderSummary{}, checkErr
		}

		return AdminOrderSummary{}, ErrAdminOrderConflict
	}

	return order, err
}

// CancelOrder membatalkan order dan mengembalikan stok varian
// dalam satu transaksi database.
//
// Row order dikunci FOR UPDATE agar dua request pembatalan
// bersamaan tidak dapat mengembalikan stok dua kali.
func (r *AdminOrderRepository) CancelOrder(
	ctx context.Context,
	orderID int64,
	expectedStatus string,
) (AdminOrderSummary, error) {
	var order AdminOrderSummary

	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{
		IsoLevel: pgx.ReadCommitted,
	})
	if err != nil {
		return order, err
	}
	defer tx.Rollback(ctx)

	// 1. Lock order dan pastikan status masih sesuai.
	var currentStatus string

	err = tx.QueryRow(ctx, `
		SELECT status
		FROM orders
		WHERE id = $1
		FOR UPDATE
	`, orderID).Scan(&currentStatus)

	if errors.Is(err, pgx.ErrNoRows) {
		return order, ErrAdminOrderNotFound
	}
	if err != nil {
		return order, err
	}

	if currentStatus != expectedStatus {
		return order, ErrAdminOrderConflict
	}

	if currentStatus != "pending" && currentStatus != "confirmed" {
		return order, ErrAdminOrderConflict
	}

	// 2. Ambil semua item order.
	rows, err := tx.Query(ctx, `
		SELECT
			product_id,
			variant_id,
			product_sku,
			quantity
		FROM order_items
		WHERE order_id = $1
		ORDER BY id ASC
	`, orderID)

	if err != nil {
		return order, err
	}

	type cancelItem struct {
		ProductID pgtype.Int8
		VariantID pgtype.Int8
		SKU       string
		Quantity  int32
	}

	items := make([]cancelItem, 0)

	for rows.Next() {
		var item cancelItem

		if err := rows.Scan(
			&item.ProductID,
			&item.VariantID,
			&item.SKU,
			&item.Quantity,
		); err != nil {
			rows.Close()
			return order, err
		}

		if item.Quantity <= 0 {
			rows.Close()
			return order, ErrAdminOrderStockRestore
		}

		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		rows.Close()
		return order, err
	}
	rows.Close()

	if len(items) == 0 {
		return order, ErrAdminOrderStockRestore
	}

	// 3. Kembalikan stok ke varian yang tepat.
	for _, item := range items {
		var (
			variantID int64
			productID int64
		)

		if item.VariantID.Valid {
			variantID = item.VariantID.Int64

			if !item.ProductID.Valid {
				return order, ErrAdminOrderStockRestore
			}

			productID = item.ProductID.Int64
		} else {
			// Fallback untuk order lama yang variant_id-nya NULL.
			// Pencarian wajib cocok dengan product_id dan SKU.
			if !item.ProductID.Valid || strings.TrimSpace(item.SKU) == "" {
				return order, ErrAdminOrderStockRestore
			}

			productID = item.ProductID.Int64

			err = tx.QueryRow(ctx, `
				SELECT id
				FROM product_variants
				WHERE product_id = $1
				  AND sku = $2
				FOR UPDATE
			`,
				productID,
				item.SKU,
			).Scan(&variantID)

			if errors.Is(err, pgx.ErrNoRows) {
				return order, ErrAdminOrderStockRestore
			}
			if err != nil {
				return order, err
			}
		}

		// Lock stok varian dan baca stok sebelumnya.
		var stockBefore int32

		err = tx.QueryRow(ctx, `
			SELECT stock_qty
			FROM product_variants
			WHERE id = $1
			  AND product_id = $2
			FOR UPDATE
		`,
			variantID,
			productID,
		).Scan(&stockBefore)

		if errors.Is(err, pgx.ErrNoRows) {
			return order, ErrAdminOrderStockRestore
		}
		if err != nil {
			return order, err
		}

		// Cegah integer overflow pada kolom INTEGER.
		var stockAfter int32

		err = tx.QueryRow(ctx, `
			UPDATE product_variants
			SET
				stock_qty = (stock_qty::BIGINT + $1::BIGINT)::INTEGER,
				updated_at = NOW()
			WHERE id = $2
			  AND product_id = $3
			  AND stock_qty::BIGINT + $1::BIGINT <= 2147483647
			RETURNING stock_qty
		`,
			item.Quantity,
			variantID,
			productID,
		).Scan(&stockAfter)

		if errors.Is(err, pgx.ErrNoRows) {
			return order, ErrAdminOrderStockRestore
		}
		if err != nil {
			return order, err
		}

		// Catat pengembalian stok.
		_, err = tx.Exec(ctx, `
			INSERT INTO inventory_movements (
				product_id,
				variant_id,
				movement_type,
				quantity,
				stock_before,
				stock_after,
				reference_type,
				reference_id,
				note
			)
			VALUES (
				$1,
				$2,
				'in',
				$3,
				$4,
				$5,
				'order',
				$6,
				$7
			)
		`,
			productID,
			variantID,
			item.Quantity,
			stockBefore,
			stockAfter,
			orderID,
			"Stock restored because order was cancelled",
		)

		if err != nil {
			return order, err
		}
	}

	// 4. Ubah status order menjadi cancelled.
	// Jika update gagal, seluruh perubahan stok di atas rollback.
	err = tx.QueryRow(ctx, `
		UPDATE orders o
		SET
			status = 'cancelled',
			updated_at = NOW()
		FROM customers c
		WHERE o.id = $1
		  AND o.status = $2
		  AND c.id = o.customer_id
		RETURNING
			o.id,
			o.order_number,
			o.customer_id,
			c.full_name,
			COALESCE(c.email, ''),
			COALESCE(c.phone, ''),
			o.status,
			o.subtotal,
			o.shipping_cost,
			o.discount_amount,
			o.grand_total,
			o.created_at,
			o.updated_at
	`,
		orderID,
		expectedStatus,
	).Scan(
		&order.ID,
		&order.OrderNumber,
		&order.CustomerID,
		&order.CustomerName,
		&order.CustomerEmail,
		&order.CustomerPhone,
		&order.Status,
		&order.Subtotal,
		&order.ShippingCost,
		&order.DiscountAmount,
		&order.GrandTotal,
		&order.CreatedAt,
		&order.UpdatedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return AdminOrderSummary{}, ErrAdminOrderConflict
	}
	if err != nil {
		return AdminOrderSummary{}, err
	}

	// 5. Commit seluruh perubahan.
	if err := tx.Commit(ctx); err != nil {
		return AdminOrderSummary{}, err
	}

	return order, nil
}

// UpdateShipping mengisi kurir dan nomor resi
// sekaligus mengubah status menjadi shipped.
// Hanya order berstatus processing yang dapat dikirim.
func (r *AdminOrderRepository) UpdateShipping(
	ctx context.Context,
	orderID int64,
	courierName string,
	trackingNumber string,
) (AdminOrderDetail, error) {
	var updatedOrderID int64

	err := r.pool.QueryRow(ctx, `
		UPDATE orders
		SET
			status = 'shipped',
			courier_name = $1,
			tracking_number = $2,
			shipped_at = NOW(),
			updated_at = NOW()
		WHERE id = $3
		  AND status = 'processing'
		RETURNING id
	`,
		courierName,
		trackingNumber,
		orderID,
	).Scan(&updatedOrderID)

	if errors.Is(err, pgx.ErrNoRows) {
		_, checkErr := r.GetStatus(ctx, orderID)

		if errors.Is(checkErr, ErrAdminOrderNotFound) {
			return AdminOrderDetail{}, ErrAdminOrderNotFound
		}
		if checkErr != nil {
			return AdminOrderDetail{}, checkErr
		}

		return AdminOrderDetail{}, ErrAdminOrderConflict
	}

	if err != nil {
		return AdminOrderDetail{}, err
	}

	return r.GetByID(ctx, updatedOrderID)
}
