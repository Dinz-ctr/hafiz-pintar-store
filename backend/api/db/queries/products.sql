
-- =========================================
-- PRODUCT QUERIES
-- Hafiz Pintar Store
-- =========================================


-- name: ListActiveProducts :many
-- Mengambil daftar produk aktif beserta varian default.
SELECT
    p.id,
    p.name,
    p.slug,
    p.description,
    p.category_id,
    c.name AS category_name,
    v.id AS variant_id,
    v.sku,
    v.name AS variant_name,
    v.price,
    v.stock_qty,
    v.weight_gram,
    p.created_at
FROM products p
JOIN product_variants v
    ON v.product_id = p.id
    AND v.is_default = TRUE
    AND v.is_active = TRUE
LEFT JOIN categories c
    ON c.id = p.category_id
WHERE p.status = 'active'
ORDER BY p.created_at DESC
LIMIT $1 OFFSET $2;


-- name: GetProductBySlug :one
-- Mengambil detail produk aktif berdasarkan slug.
SELECT
    p.id,
    p.name,
    p.slug,
    p.description,
    p.category_id,
    c.name AS category_name,
    v.id AS variant_id,
    v.sku,
    v.name AS variant_name,
    v.price,
    v.stock_qty,
    v.weight_gram,
    p.created_at
FROM products p
JOIN product_variants v
    ON v.product_id = p.id
    AND v.is_default = TRUE
    AND v.is_active = TRUE
LEFT JOIN categories c
    ON c.id = p.category_id
WHERE p.slug = $1
    AND p.status = 'active'
LIMIT 1;


-- =========================================
-- LEGACY QUERIES
-- Jangan digunakan untuk alur CRUD baru.
-- Akan digantikan dengan transaksi produk
-- dan varian pada tahap berikutnya.
-- =========================================


-- name: CreateProduct :one
INSERT INTO products (
    category_id,
    sku,
    name,
    slug,
    description,
    price,
    stock_qty,
    weight_gram
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING *;


-- name: UpdateProductStock :one
UPDATE products
SET
    stock_qty = $2,
    updated_at = NOW()
WHERE id = $1
RETURNING id, sku, name, stock_qty, updated_at;


-- =========================================
-- PRODUCT IMAGES
-- =========================================


-- name: ListProductImages :many
SELECT
    id,
    product_id,
    image_url,
    alt_text,
    is_primary,
    sort_order
FROM product_images
WHERE product_id = $1
ORDER BY
    is_primary DESC,
    sort_order ASC,
    id ASC;


-- name: CreateProductBase :one
INSERT INTO products (
    category_id,
    sku,
    name,
    slug,
    description,
    price,
    stock_qty,
    weight_gram,
    status
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, 'active')
RETURNING id, name, slug;


-- name: CreateProductVariant :one
INSERT INTO product_variants (
    product_id,
    sku,
    name,
    price,
    stock_qty,
    weight_gram,
    is_default,
    is_active
)
VALUES ($1, $2, $3, $4, $5, $6, $7, TRUE)
RETURNING id, product_id, sku, name;


-- name: CreateVariantAttribute :one
INSERT INTO variant_attributes (
    product_id,
    name
)
VALUES ($1, $2)
RETURNING id, product_id, name;


-- name: CreateVariantAttributeValue :one
INSERT INTO variant_attribute_values (
    variant_id,
    product_id,
    attribute_id,
    value
)
VALUES ($1, $2, $3, $4)
RETURNING id, variant_id, attribute_id, value;

-- name: ListProductVariants :many
SELECT
    id,
    product_id,
    sku,
    name,
    price,
    stock_qty,
    weight_gram,
    is_default,
    is_active
FROM product_variants
WHERE product_id = $1
  AND is_active = TRUE
ORDER BY is_default DESC, id ASC;


-- name: ListVariantAttributes :many
SELECT
    va.id AS attribute_id,
    va.name AS attribute_name,
    vav.value AS attribute_value
FROM variant_attribute_values vav
JOIN variant_attributes va
    ON va.id = vav.attribute_id
   AND va.product_id = vav.product_id
WHERE vav.variant_id = $1
ORDER BY va.name ASC;

-- =========================================
-- PRODUCT MANAGEMENT: UPDATE & DELETE
-- =========================================

-- name: UpdateProductDetails :one
UPDATE products
SET
    name = $2,
    slug = $3,
    description = $4,
    updated_at = NOW()
WHERE id = $1
RETURNING
    id,
    name,
    slug,
    description,
    category_id,
    updated_at;


-- name: UpdateProductVariant :one
UPDATE product_variants
SET
    name = $3,
    price = $4,
    stock_qty = $5,
    weight_gram = $6,
    updated_at = NOW()
WHERE id = $1
  AND product_id = $2
RETURNING
    id,
    product_id,
    sku,
    name,
    price,
    stock_qty,
    weight_gram,
    is_default,
    is_active;


-- name: DeleteProductVariant :execrows
DELETE FROM product_variants
WHERE id = $1
  AND product_id = $2
  AND is_default = FALSE;


-- name: DeleteProduct :execrows
DELETE FROM products
WHERE id = $1;


-- =========================================
-- PRODUCT MANAGEMENT: SEARCH & FILTER
-- =========================================

-- name: SearchActiveProducts :many
SELECT
    p.id,
    p.name,
    p.slug,
    p.description,
    p.category_id,
    c.name AS category_name,
    v.id AS variant_id,
    v.sku,
    v.name AS variant_name,
    v.price,
    v.stock_qty,
    v.weight_gram,
    p.created_at
FROM products p
JOIN product_variants v
    ON v.product_id = p.id
    AND v.is_default = TRUE
    AND v.is_active = TRUE
LEFT JOIN categories c
    ON c.id = p.category_id
WHERE p.status = 'active'
  AND (
      p.name ILIKE '%' || sqlc.arg(keyword)::text || '%'
      OR p.description ILIKE '%' || sqlc.arg(keyword)::text || '%'
  )
  AND (
      sqlc.narg(category_id)::bigint IS NULL
      OR p.category_id = sqlc.narg(category_id)::bigint
  )
ORDER BY p.created_at DESC
LIMIT sqlc.arg(limit_count)::integer
OFFSET sqlc.arg(offset_count)::integer;


