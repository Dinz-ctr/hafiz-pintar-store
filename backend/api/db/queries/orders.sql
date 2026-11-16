
-- =========================================
-- ORDER QUERIES
-- Hafiz Pintar Store
-- =========================================


-- =========================================
-- CUSTOMER: CREATE ORDER
-- =========================================

-- name: CreateOrder :one
-- Membuat pesanan baru.
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
    sqlc.arg(order_number),
    sqlc.arg(customer_id),
    'pending',
    sqlc.arg(subtotal),
    sqlc.arg(shipping_cost),
    sqlc.arg(discount_amount),
    sqlc.arg(grand_total),
    sqlc.arg(recipient_name),
    sqlc.arg(recipient_phone),
    sqlc.arg(shipping_address),
    sqlc.arg(shipping_city),
    sqlc.arg(shipping_province),
    sqlc.arg(shipping_postal_code),
    sqlc.narg(customer_note)
)
RETURNING
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
    updated_at;


-- =========================================
-- CUSTOMER: ORDER DETAIL
-- =========================================

-- name: GetOrderByID :one
-- Mengambil detail pesanan berdasarkan ID.
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
WHERE id = $1
LIMIT 1;


-- name: GetOrderByOrderNumber :one
-- Mengambil pesanan berdasarkan nomor pesanan.
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
WHERE order_number = $1
LIMIT 1;


-- name: GetCustomerOrderByID :one
-- Memastikan pesanan dimiliki pelanggan yang sedang login.
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
WHERE id = $1
    AND customer_id = $2
LIMIT 1;


-- =========================================
-- CUSTOMER: ORDER HISTORY
-- =========================================

-- name: ListCustomerOrders :many
-- Mengambil riwayat pesanan pelanggan.
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
LIMIT $2
OFFSET $3;


-- name: CountCustomerOrders :one
-- Menghitung jumlah pesanan pelanggan.
SELECT COUNT(*)
FROM orders
WHERE customer_id = $1;


-- =========================================
-- CUSTOMER: CANCEL ORDER
-- =========================================

-- name: CancelCustomerOrder :execrows
-- Pelanggan hanya dapat membatalkan pesanan yang masih pending.
UPDATE orders
SET
    status = 'cancelled',
    updated_at = NOW()
WHERE id = $1
    AND customer_id = $2
    AND status = 'pending';


-- =========================================
-- ADMIN: ORDER MANAGEMENT
-- =========================================

-- name: ListAllOrders :many
-- Mengambil seluruh pesanan untuk admin.
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
ORDER BY created_at DESC, id DESC
LIMIT $1
OFFSET $2;


-- name: ListOrdersByStatus :many
-- Mengambil pesanan berdasarkan status.
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
WHERE status = $1
ORDER BY created_at DESC, id DESC
LIMIT $2
OFFSET $3;


-- name: CountAllOrders :one
-- Menghitung seluruh pesanan.
SELECT COUNT(*)
FROM orders;


-- name: CountOrdersByStatus :one
-- Menghitung pesanan berdasarkan status.
SELECT COUNT(*)
FROM orders
WHERE status = $1;


-- =========================================
-- ADMIN: UPDATE ORDER STATUS
-- =========================================

-- name: UpdateOrderStatus :one
-- Memperbarui status pesanan.
UPDATE orders
SET
    status = $2,
    updated_at = NOW()
WHERE id = $1
RETURNING
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
    updated_at;


-- =========================================
-- ADMIN: SEARCH ORDERS
-- =========================================

-- name: SearchOrders :many
-- Mencari pesanan berdasarkan nomor pesanan,
-- nama penerima, atau nomor telepon.
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
WHERE
    COALESCE(order_number, '') ILIKE
        '%' || sqlc.arg(keyword)::text || '%'
    OR COALESCE(recipient_name, '') ILIKE
        '%' || sqlc.arg(keyword)::text || '%'
    OR COALESCE(recipient_phone, '') ILIKE
        '%' || sqlc.arg(keyword)::text || '%'
ORDER BY created_at DESC, id DESC
LIMIT sqlc.arg(limit_count)::integer
OFFSET sqlc.arg(offset_count)::integer;
