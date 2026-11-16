
-- =========================================
-- CUSTOMER ADDRESS QUERIES
-- Hafiz Pintar Store
-- =========================================


-- =========================================
-- CUSTOMER: LIST ADDRESSES
-- =========================================

-- name: ListCustomerAddresses :many
-- Mengambil semua alamat milik pelanggan.
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
ORDER BY
    is_default DESC,
    created_at DESC,
    id DESC;


-- =========================================
-- CUSTOMER: GET ADDRESS
-- =========================================

-- name: GetCustomerAddressByID :one
-- Mengambil alamat berdasarkan ID dan pemiliknya.
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
WHERE id = $1
    AND customer_id = $2
LIMIT 1;


-- name: GetDefaultCustomerAddress :one
-- Mengambil alamat utama pelanggan.
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
    AND is_default = TRUE
LIMIT 1;


-- =========================================
-- CUSTOMER: CREATE ADDRESS
-- =========================================

-- name: CreateCustomerAddress :one
-- Menambahkan alamat baru.
INSERT INTO customer_addresses (
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
    is_default
)
VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11
)
RETURNING
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
    updated_at;


-- =========================================
-- CUSTOMER: UPDATE ADDRESS
-- =========================================

-- name: UpdateCustomerAddress :one
-- Memperbarui alamat milik pelanggan.
UPDATE customer_addresses
SET
    label = $3,
    recipient_name = $4,
    recipient_phone = $5,
    address_line = $6,
    subdistrict = $7,
    district = $8,
    city = $9,
    province = $10,
    postal_code = $11,
    updated_at = NOW()
WHERE id = $1
    AND customer_id = $2
RETURNING
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
    updated_at;


-- =========================================
-- CUSTOMER: SET DEFAULT ADDRESS
-- =========================================

-- name: ClearDefaultCustomerAddress :exec
-- Menghapus status default dari semua alamat pelanggan.
UPDATE customer_addresses
SET
    is_default = FALSE,
    updated_at = NOW()
WHERE customer_id = $1;


-- name: SetDefaultCustomerAddress :execrows
-- Menjadikan alamat tertentu sebagai alamat utama.
UPDATE customer_addresses
SET
    is_default = TRUE,
    updated_at = NOW()
WHERE id = $1
    AND customer_id = $2;


-- =========================================
-- CUSTOMER: DELETE ADDRESS
-- =========================================

-- name: DeleteCustomerAddress :execrows
-- Menghapus alamat milik pelanggan.
DELETE FROM customer_addresses
WHERE id = $1
    AND customer_id = $2;


-- =========================================
-- CUSTOMER: ADDRESS STATISTICS
-- =========================================

-- name: CountCustomerAddresses :one
-- Menghitung jumlah alamat pelanggan.
SELECT COUNT(*)
FROM customer_addresses
WHERE customer_id = $1;
