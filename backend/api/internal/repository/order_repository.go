package repository

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// =========================================
// CHECKOUT ERRORS
// =========================================

var (
	ErrCheckoutCartNotFound = errors.New("cart not found")
	ErrCheckoutCartEmpty    = errors.New("cart is empty")
	ErrCheckoutAddress      = errors.New("shipping address not found")
	ErrCheckoutStock        = errors.New("insufficient product stock")
	ErrCheckoutUnavailable  = errors.New("product or variant unavailable")
)

// =========================================
// CHECKOUT RESPONSE
// =========================================

type CheckoutResult struct {
	ID             int64     `json:"id"`
	OrderNumber    string    `json:"order_number"`
	Status         string    `json:"status"`
	Subtotal       int64     `json:"subtotal"`
	ShippingCost   int64     `json:"shipping_cost"`
	DiscountAmount int64     `json:"discount_amount"`
	GrandTotal     int64     `json:"grand_total"`
	CreatedAt      time.Time `json:"created_at"`
}

// =========================================
// ORDER HISTORY RESPONSE
// =========================================

type OrderSummary struct {
	ID             int64     `json:"id"`
	OrderNumber    string    `json:"order_number"`
	Status         string    `json:"status"`
	Subtotal       int64     `json:"subtotal"`
	ShippingCost   int64     `json:"shipping_cost"`
	DiscountAmount int64     `json:"discount_amount"`
	GrandTotal     int64     `json:"grand_total"`
	CreatedAt      time.Time `json:"created_at"`
}

// =========================================
// ORDER ITEM DETAIL RESPONSE
// =========================================

type OrderItemDetail struct {
	ID          int64  `json:"id"`
	ProductSKU  string `json:"product_sku"`
	ProductName string `json:"product_name"`
	Quantity    int32  `json:"quantity"`
	UnitPrice   int64  `json:"unit_price"`
	Subtotal    int64  `json:"subtotal"`
}

// =========================================
// ORDER DETAIL RESPONSE
// =========================================

type OrderDetail struct {
	OrderSummary

	RecipientName      string            `json:"recipient_name"`
	RecipientPhone     string            `json:"recipient_phone"`
	ShippingAddress    string            `json:"shipping_address"`
	ShippingCity       string            `json:"shipping_city"`
	ShippingProvince   string            `json:"shipping_province"`
	ShippingPostalCode string            `json:"shipping_postal_code"`
	CustomerNote       string            `json:"customer_note"`
	Items              []OrderItemDetail `json:"items"`
}

// =========================================
// ORDER REPOSITORY
// =========================================

type OrderRepository struct {
	pool *pgxpool.Pool
}

func NewOrderRepository(pool *pgxpool.Pool) *OrderRepository {
	return &OrderRepository{
		pool: pool,
	}
}

// =========================================
// CHECKOUT ITEM
// =========================================

type checkoutItem struct {
	ProductID   int64
	VariantID   int64
	Quantity    int32
	ProductName string
	ProductSKU  string
	VariantName string
	VariantSKU  string
	Price       int64
	Stock       int32
}

// =========================================
// GENERATE ORDER NUMBER
// =========================================

func generateOrderNumber() (string, error) {
	randomBytes := make([]byte, 4)

	if _, err := rand.Read(randomBytes); err != nil {
		return "", err
	}

	return fmt.Sprintf(
		"HP-%s-%s",
		time.Now().Format("20060102"),
		strings.ToUpper(hex.EncodeToString(randomBytes)),
	), nil
}

// =========================================
// CHECKOUT
// =========================================

func (r *OrderRepository) Checkout(
	ctx context.Context,
	customerID int64,
	addressID int64,
	customerNote string,
) (CheckoutResult, error) {

	var result CheckoutResult

	if customerID <= 0 || addressID <= 0 {
		return result, ErrCheckoutAddress
	}

	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{
		IsoLevel: pgx.ReadCommitted,
	})
	if err != nil {
		return result, err
	}

	defer tx.Rollback(ctx)

	// 1. Lock cart customer.
	var cartID int64

	err = tx.QueryRow(ctx, `
		SELECT id
		FROM carts
		WHERE customer_id = $1
		FOR UPDATE
	`, customerID).Scan(&cartID)

	if errors.Is(err, pgx.ErrNoRows) {
		return result, ErrCheckoutCartNotFound
	}

	if err != nil {
		return result, err
	}

	// 2. Ambil alamat milik customer.
	var (
		recipientName  string
		recipientPhone string
		addressLine    string
		subdistrict    string
		district       string
		city           string
		province       string
		postalCode     string
	)

	err = tx.QueryRow(ctx, `
		SELECT
			recipient_name,
			recipient_phone,
			address_line,
			COALESCE(subdistrict, ''),
			COALESCE(district, ''),
			city,
			province,
			COALESCE(postal_code, '')
		FROM customer_addresses
		WHERE id = $1
		  AND customer_id = $2
		FOR SHARE
	`,
		addressID,
		customerID,
	).Scan(
		&recipientName,
		&recipientPhone,
		&addressLine,
		&subdistrict,
		&district,
		&city,
		&province,
		&postalCode,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return result, ErrCheckoutAddress
	}

	if err != nil {
		return result, err
	}

	// 3. Ambil item cart dan lock stok variant.
	rows, err := tx.Query(ctx, `
		SELECT
			ci.product_id,
			ci.variant_id,
			ci.quantity,
			p.name,
			p.sku,
			p.status,
			pv.name,
			pv.sku,
			pv.price,
			pv.stock_qty,
			pv.is_active
		FROM cart_items ci
		JOIN products p
			ON p.id = ci.product_id
		JOIN product_variants pv
			ON pv.id = ci.variant_id
		   AND pv.product_id = ci.product_id
		WHERE ci.cart_id = $1
		ORDER BY ci.id
		FOR UPDATE OF ci, p, pv
	`, cartID)

	if err != nil {
		return result, err
	}

	items := make([]checkoutItem, 0)

	for rows.Next() {
		var item checkoutItem
		var productStatus string
		var variantActive bool

		err = rows.Scan(
			&item.ProductID,
			&item.VariantID,
			&item.Quantity,
			&item.ProductName,
			&item.ProductSKU,
			&productStatus,
			&item.VariantName,
			&item.VariantSKU,
			&item.Price,
			&item.Stock,
			&variantActive,
		)

		if err != nil {
			rows.Close()
			return result, err
		}

		if productStatus != "active" || !variantActive {
			rows.Close()
			return result, ErrCheckoutUnavailable
		}

		if item.Quantity <= 0 || item.Stock < item.Quantity {
			rows.Close()
			return result, ErrCheckoutStock
		}

		items = append(items, item)
	}

	if err = rows.Err(); err != nil {
		rows.Close()
		return result, err
	}

	rows.Close()

	if len(items) == 0 {
		return result, ErrCheckoutCartEmpty
	}

	// 4. Hitung subtotal berdasarkan harga database.
	var subtotal int64

	const maxInt64 = int64(^uint64(0) >> 1)

	for _, item := range items {
		if item.Price < 0 ||
			int64(item.Quantity) > 0 &&
				item.Price > maxInt64/int64(item.Quantity) {
			return result, errors.New("invalid checkout total")
		}

		lineTotal := item.Price * int64(item.Quantity)

		if subtotal > maxInt64-lineTotal {
			return result, errors.New("invalid checkout total")
		}

		subtotal += lineTotal
	}

	// MVP: ongkir dan diskon belum diintegrasikan.
	const shippingCost int64 = 0
	const discountAmount int64 = 0

	grandTotal := subtotal + shippingCost - discountAmount

	if grandTotal < 0 {
		return result, errors.New("invalid grand total")
	}

	// 5. Generate nomor order.
	orderNumber, err := generateOrderNumber()
	if err != nil {
		return result, err
	}

	shippingAddress := strings.Join(
		filterEmpty([]string{
			addressLine,
			subdistrict,
			district,
			city,
			province,
			postalCode,
		}),
		", ",
	)

	var (
		postalCodePtr *string
		notePtr       *string
	)

	if postalCode != "" {
		postalCodePtr = &postalCode
	}

	if strings.TrimSpace(customerNote) != "" {
		note := strings.TrimSpace(customerNote)
		notePtr = &note
	}

	// 6. Buat order.
	err = tx.QueryRow(ctx, `
		INSERT INTO orders (
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
			customer_note
		)
		VALUES (
			$1, $2, 'pending',
			$3, $4, $5, $6,
			$7, $8, $9, $10, $11, $12, $13
		)
		RETURNING
			id,
			order_number,
			status,
			subtotal,
			shipping_cost,
			discount_amount,
			grand_total,
			created_at
	`,
		orderNumber,
		customerID,
		subtotal,
		shippingCost,
		discountAmount,
		grandTotal,
		recipientName,
		recipientPhone,
		shippingAddress,
		city,
		province,
		postalCodePtr,
		notePtr,
	).Scan(
		&result.ID,
		&result.OrderNumber,
		&result.Status,
		&result.Subtotal,
		&result.ShippingCost,
		&result.DiscountAmount,
		&result.GrandTotal,
		&result.CreatedAt,
	)

	if err != nil {
		return result, err
	}

	// 7. Simpan item, kurangi stok, catat inventory.
	for _, item := range items {
		lineTotal := item.Price * int64(item.Quantity)

		_, err = tx.Exec(ctx, `
			INSERT INTO order_items (
				order_id,
				product_id,
				product_sku,
				product_name,
				quantity,
				unit_price,
				subtotal
			)
			VALUES ($1, $2, $3, $4, $5, $6, $7)
		`,
			result.ID,
			item.ProductID,
			item.VariantSKU,
			item.ProductName+" - "+item.VariantName,
			item.Quantity,
			item.Price,
			lineTotal,
		)

		if err != nil {
			return result, err
		}

		var stockAfter int32

		err = tx.QueryRow(ctx, `
			UPDATE product_variants
			SET
				stock_qty = stock_qty - $1,
				updated_at = NOW()
			WHERE id = $2
			  AND product_id = $3
			  AND is_active = TRUE
			  AND stock_qty >= $1
			RETURNING stock_qty
		`,
			item.Quantity,
			item.VariantID,
			item.ProductID,
		).Scan(&stockAfter)

		if errors.Is(err, pgx.ErrNoRows) {
			return result, ErrCheckoutStock
		}

		if err != nil {
			return result, err
		}

		referenceType := "order"
		note := "Stock deducted from checkout"

		_, err = tx.Exec(ctx, `
			INSERT INTO inventory_movements (
				product_id,
				movement_type,
				quantity,
				stock_before,
				stock_after,
				reference_type,
				reference_id,
				note
			)
			VALUES ($1, 'out', $2, $3, $4, $5, $6, $7)
		`,
			item.ProductID,
			item.Quantity,
			item.Stock,
			stockAfter,
			referenceType,
			result.ID,
			note,
		)

		if err != nil {
			return result, err
		}
	}

	// 8. Kosongkan cart.
	_, err = tx.Exec(ctx, `
		DELETE FROM cart_items
		WHERE cart_id = $1
	`, cartID)

	if err != nil {
		return result, err
	}

	// 9. Commit transaksi.
	if err = tx.Commit(ctx); err != nil {
		return result, err
	}

	return result, nil
}

// =========================================
// GET CUSTOMER ORDER HISTORY
// =========================================

func (r *OrderRepository) GetCustomerOrders(
	ctx context.Context,
	customerID int64,
	limit int32,
	offset int32,
) ([]OrderSummary, error) {

	orders := make([]OrderSummary, 0)

	rows, err := r.pool.Query(ctx, `
		SELECT
			id,
			order_number,
			status,
			subtotal,
			shipping_cost,
			discount_amount,
			grand_total,
			created_at
		FROM orders
		WHERE customer_id = $1
		ORDER BY created_at DESC, id DESC
		LIMIT $2 OFFSET $3
	`,
		customerID,
		limit,
		offset,
	)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	for rows.Next() {
		var order OrderSummary

		err := rows.Scan(
			&order.ID,
			&order.OrderNumber,
			&order.Status,
			&order.Subtotal,
			&order.ShippingCost,
			&order.DiscountAmount,
			&order.GrandTotal,
			&order.CreatedAt,
		)

		if err != nil {
			return nil, err
		}

		orders = append(orders, order)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return orders, nil
}

// =========================================
// GET CUSTOMER ORDER DETAIL
// =========================================

func (r *OrderRepository) GetCustomerOrderDetail(
	ctx context.Context,
	customerID int64,
	orderID int64,
) (OrderDetail, error) {

	var order OrderDetail

	err := r.pool.QueryRow(ctx, `
		SELECT
			id,
			order_number,
			status,
			subtotal,
			shipping_cost,
			discount_amount,
			grand_total,
			created_at,
			recipient_name,
			recipient_phone,
			shipping_address,
			shipping_city,
			shipping_province,
			COALESCE(shipping_postal_code, ''),
			COALESCE(customer_note, '')
		FROM orders
		WHERE id = $1
		  AND customer_id = $2
	`,
		orderID,
		customerID,
	).Scan(
		&order.ID,
		&order.OrderNumber,
		&order.Status,
		&order.Subtotal,
		&order.ShippingCost,
		&order.DiscountAmount,
		&order.GrandTotal,
		&order.CreatedAt,
		&order.RecipientName,
		&order.RecipientPhone,
		&order.ShippingAddress,
		&order.ShippingCity,
		&order.ShippingProvince,
		&order.ShippingPostalCode,
		&order.CustomerNote,
	)

	if err != nil {
		return OrderDetail{}, err
	}

	order.Items = make([]OrderItemDetail, 0)

	rows, err := r.pool.Query(ctx, `
		SELECT
			id,
			product_sku,
			product_name,
			quantity,
			unit_price,
			subtotal
		FROM order_items
		WHERE order_id = $1
		ORDER BY id ASC
	`, order.ID)

	if err != nil {
		return OrderDetail{}, err
	}

	defer rows.Close()

	for rows.Next() {
		var item OrderItemDetail

		err := rows.Scan(
			&item.ID,
			&item.ProductSKU,
			&item.ProductName,
			&item.Quantity,
			&item.UnitPrice,
			&item.Subtotal,
		)

		if err != nil {
			return OrderDetail{}, err
		}

		order.Items = append(order.Items, item)
	}

	if err := rows.Err(); err != nil {
		return OrderDetail{}, err
	}

	return order, nil
}

// =========================================
// FILTER EMPTY ADDRESS VALUES
// =========================================

func filterEmpty(values []string) []string {
	result := make([]string, 0, len(values))

	for _, value := range values {
		value = strings.TrimSpace(value)

		if value != "" {
			result = append(result, value)
		}
	}

	return result
}
