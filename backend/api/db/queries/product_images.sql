
-- =========================================
-- PRODUCT IMAGE QUERIES
-- Hafiz Pintar Store
-- =========================================


-- =========================================
-- PUBLIC: PRODUCT IMAGES
-- =========================================

-- name: ListProductImages :many
-- Mengambil seluruh gambar produk.
SELECT
    id,
    product_id,
    image_url,
    alt_text,
    is_primary,
    sort_order,
    created_at
FROM product_images
WHERE product_id = $1
ORDER BY
    is_primary DESC,
    sort_order ASC,
    id ASC;


-- name: GetPrimaryProductImage :one
-- Mengambil gambar utama produk.
SELECT
    id,
    product_id,
    image_url,
    alt_text,
    is_primary,
    sort_order,
    created_at
FROM product_images
WHERE product_id = $1
    AND is_primary = TRUE
ORDER BY id ASC
LIMIT 1;


-- =========================================
-- ADMIN: CREATE IMAGE
-- =========================================

-- name: CreateProductImage :one
-- Menambahkan gambar produk.
INSERT INTO product_images (
    product_id,
    image_url,
    alt_text,
    is_primary,
    sort_order
)
VALUES ($1, $2, $3, $4, $5)
RETURNING
    id,
    product_id,
    image_url,
    alt_text,
    is_primary,
    sort_order,
    created_at;


-- =========================================
-- ADMIN: UPDATE IMAGE
-- =========================================

-- name: UpdateProductImage :one
-- Memperbarui informasi gambar.
UPDATE product_images
SET
    image_url = $3,
    alt_text = $4,
    sort_order = $5
WHERE id = $1
    AND product_id = $2
RETURNING
    id,
    product_id,
    image_url,
    alt_text,
    is_primary,
    sort_order,
    created_at;


-- =========================================
-- ADMIN: SET PRIMARY IMAGE
-- =========================================

-- name: ClearPrimaryProductImages :exec
-- Menghapus status utama dari seluruh gambar produk.
UPDATE product_images
SET is_primary = FALSE
WHERE product_id = $1;


-- name: SetPrimaryProductImage :execrows
-- Menentukan gambar utama produk.
UPDATE product_images
SET is_primary = TRUE
WHERE id = $1
    AND product_id = $2;


-- =========================================
-- ADMIN: DELETE IMAGE
-- =========================================

-- name: DeleteProductImage :execrows
-- Menghapus gambar berdasarkan ID dan produk.
DELETE FROM product_images
WHERE id = $1
    AND product_id = $2;
