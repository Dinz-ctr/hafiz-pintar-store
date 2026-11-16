
-- =========================================
-- CUSTOMER QUERIES
-- Hafiz Pintar Store
-- =========================================


-- =========================================
-- CUSTOMER AUTHENTICATION
-- =========================================

-- name: GetCustomerByEmail :one
-- Mencari pelanggan berdasarkan email.
SELECT
    id,
    full_name,
    email,
    phone,
    password_hash,
    is_active,
    created_at,
    updated_at
FROM customers
WHERE email = $1
LIMIT 1;


-- name: GetCustomerByPhone :one
-- Mencari pelanggan berdasarkan nomor HP.
SELECT
    id,
    full_name,
    email,
    phone,
    password_hash,
    is_active,
    created_at,
    updated_at
FROM customers
WHERE phone = $1
LIMIT 1;


-- name: GetCustomerByEmailOrPhone :one
-- Login menggunakan email atau nomor HP.
SELECT
    id,
    full_name,
    email,
    phone,
    password_hash,
    is_active,
    created_at,
    updated_at
FROM customers
WHERE email = $1
    OR phone = $1
LIMIT 1;


-- =========================================
-- CUSTOMER REGISTRATION
-- =========================================

-- name: CreateCustomer :one
-- Registrasi pelanggan baru.
INSERT INTO customers (
    full_name,
    email,
    phone,
    password_hash,
    is_active
)
VALUES ($1, $2, $3, $4, TRUE)
RETURNING
    id,
    full_name,
    email,
    phone,
    is_active,
    created_at,
    updated_at;


-- =========================================
-- CUSTOMER PROFILE
-- =========================================

-- name: GetCustomerByID :one
-- Mengambil profil pelanggan berdasarkan ID.
SELECT
    id,
    full_name,
    email,
    phone,
    is_active,
    created_at,
    updated_at
FROM customers
WHERE id = $1
LIMIT 1;


-- name: UpdateCustomerProfile :one
-- Memperbarui profil pelanggan.
UPDATE customers
SET
    full_name = $2,
    email = $3,
    phone = $4,
    updated_at = NOW()
WHERE id = $1
RETURNING
    id,
    full_name,
    email,
    phone,
    is_active,
    created_at,
    updated_at;


-- =========================================
-- CUSTOMER PASSWORD
-- =========================================

-- name: UpdateCustomerPassword :execrows
-- Memperbarui password hash pelanggan.
UPDATE customers
SET
    password_hash = $2,
    updated_at = NOW()
WHERE id = $1;


-- =========================================
-- CUSTOMER ACCOUNT STATUS
-- =========================================

-- name: DeactivateCustomer :execrows
-- Menonaktifkan akun pelanggan.
UPDATE customers
SET
    is_active = FALSE,
    updated_at = NOW()
WHERE id = $1;


-- name: ActivateCustomer :execrows
-- Mengaktifkan kembali akun pelanggan.
UPDATE customers
SET
    is_active = TRUE,
    updated_at = NOW()
WHERE id = $1;


-- =========================================
-- ADMIN: CUSTOMER MANAGEMENT
-- =========================================

-- name: ListCustomers :many
-- Mengambil daftar pelanggan untuk admin.
SELECT
    id,
    full_name,
    email,
    phone,
    is_active,
    created_at,
    updated_at
FROM customers
ORDER BY created_at DESC, id DESC
LIMIT $1
OFFSET $2;


-- name: CountCustomers :one
-- Menghitung jumlah seluruh pelanggan.
SELECT COUNT(*)
FROM customers;


-- name: SearchCustomers :many
-- Mencari pelanggan berdasarkan nama, email, atau nomor HP.
SELECT
    id,
    full_name,
    email,
    phone,
    is_active,
    created_at,
    updated_at
FROM customers
WHERE
    full_name ILIKE '%' || sqlc.arg(keyword)::text || '%'
    OR COALESCE(email, '') ILIKE
        '%' || sqlc.arg(keyword)::text || '%'
    OR COALESCE(phone, '') ILIKE
        '%' || sqlc.arg(keyword)::text || '%'
ORDER BY created_at DESC, id DESC
LIMIT sqlc.arg(limit_count)::integer
OFFSET sqlc.arg(offset_count)::integer;
