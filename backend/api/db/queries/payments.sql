
-- =========================================
-- PAYMENT QUERIES
-- Hafiz Pintar Store
-- =========================================


-- =========================================
-- CUSTOMER: CREATE PAYMENT
-- =========================================

-- name: CreatePayment :one
-- Membuat catatan pembayaran untuk pesanan.
INSERT INTO payments (
    order_id,
    payment_method,
    payment_provider,
    transaction_reference,
    amount,
    status
)
VALUES (
    sqlc.arg(order_id),
    sqlc.arg(payment_method),
    sqlc.narg(payment_provider),
    sqlc.narg(transaction_reference),
    sqlc.arg(amount),
    'pending'
)
RETURNING
    id,
    order_id,
    payment_method,
    payment_provider,
    transaction_reference,
    amount,
    status,
    paid_at,
    created_at,
    updated_at;


-- =========================================
-- CUSTOMER: GET PAYMENT
-- =========================================

-- name: GetPaymentByID :one
-- Mengambil detail pembayaran berdasarkan ID.
SELECT
    id,
    order_id,
    payment_method,
    payment_provider,
    transaction_reference,
    amount,
    status,
    paid_at,
    created_at,
    updated_at
FROM payments
WHERE id = $1
LIMIT 1;


-- name: GetPaymentByOrderID :one
-- Mengambil pembayaran terbaru berdasarkan ID pesanan.
SELECT
    id,
    order_id,
    payment_method,
    payment_provider,
    transaction_reference,
    amount,
    status,
    paid_at,
    created_at,
    updated_at
FROM payments
WHERE order_id = $1
ORDER BY created_at DESC, id DESC
LIMIT 1;


-- name: GetPaymentByTransactionReference :one
-- Mencari pembayaran berdasarkan referensi transaksi.
SELECT
    id,
    order_id,
    payment_method,
    payment_provider,
    transaction_reference,
    amount,
    status,
    paid_at,
    created_at,
    updated_at
FROM payments
WHERE transaction_reference = $1
LIMIT 1;


-- =========================================
-- PAYMENT STATUS
-- =========================================

-- name: UpdatePaymentStatus :one
-- Memperbarui status pembayaran.
UPDATE payments
SET
    status = $2,
    paid_at = CASE
        WHEN $2 = 'paid' THEN COALESCE(paid_at, NOW())
        ELSE paid_at
    END,
    updated_at = NOW()
WHERE id = $1
RETURNING
    id,
    order_id,
    payment_method,
    payment_provider,
    transaction_reference,
    amount,
    status,
    paid_at,
    created_at,
    updated_at;


-- =========================================
-- PAYMENT TRANSACTION
-- =========================================

-- name: UpdatePaymentTransactionReference :one
-- Memperbarui referensi transaksi dari payment gateway.
UPDATE payments
SET
    payment_provider = $2,
    transaction_reference = $3,
    updated_at = NOW()
WHERE id = $1
RETURNING
    id,
    order_id,
    payment_method,
    payment_provider,
    transaction_reference,
    amount,
    status,
    paid_at,
    created_at,
    updated_at;


-- =========================================
-- ADMIN: PAYMENT MANAGEMENT
-- =========================================

-- name: ListAllPayments :many
-- Mengambil seluruh data pembayaran.
SELECT
    id,
    order_id,
    payment_method,
    payment_provider,
    transaction_reference,
    amount,
    status,
    paid_at,
    created_at,
    updated_at
FROM payments
ORDER BY created_at DESC, id DESC
LIMIT $1
OFFSET $2;


-- name: ListPaymentsByStatus :many
-- Mengambil pembayaran berdasarkan status.
SELECT
    id,
    order_id,
    payment_method,
    payment_provider,
    transaction_reference,
    amount,
    status,
    paid_at,
    created_at,
    updated_at
FROM payments
WHERE status = $1
ORDER BY created_at DESC, id DESC
LIMIT $2
OFFSET $3;


-- name: CountAllPayments :one
-- Menghitung seluruh transaksi pembayaran.
SELECT COUNT(*)
FROM payments;


-- name: CountPaymentsByStatus :one
-- Menghitung transaksi berdasarkan status.
SELECT COUNT(*)
FROM payments
WHERE status = $1;


-- =========================================
-- ADMIN: PAYMENT STATISTICS
-- =========================================

-- name: CalculateTotalSuccessfulPayments :one
-- Menghitung total nominal pembayaran berhasil.
SELECT COALESCE(SUM(amount), 0)::bigint
FROM payments
WHERE status = 'paid';
