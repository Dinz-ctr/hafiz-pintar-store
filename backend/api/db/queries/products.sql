
-- =========================================
-- PRODUCT QUERIES
-- Hafiz Pintar Store
-- =========================================


-- =========================================
-- PRODUCT CATALOG: LIST, SEARCH & FILTER
-- =========================================

-- name: ListActiveProducts :many
-- Mengambil produk aktif dengan filter kategori dan pencarian.
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
        sqlc.arg(category_slug)::text = ''
        OR c.slug = sqlc.arg(category_slug)::text
    )
    AND (
        sqlc.arg(search)::text = ''
        OR p.name ILIKE '%' || sqlc.arg(search)::text || '%'
        OR COALESCE(p.description, '') ILIKE
            '%' || sqlc.arg(search)::text || '%'
    )
ORDER BY p.created_at DESC, p.id DESC
LIMIT sqlc.arg(limit_count)::integer
OFFSET sqlc.arg(offset_count)::integer;


-- =========================================
-- PRODUCT: DETAIL
-- =========================================

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
-- LEGACY PRODUCT QUERIES
-- Jangan digunakan untuk alur CRUD baru.
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
-- LEGACY: stok baru dikelola melalui product_variants.
UPDATE products
SET
    stock_qty = $2,
    updated_at = NOW()
WHERE id = $1
RETURNING id, sku, name, stock_qty, updated_at;


-- =========================================
-- PRODUCT CREATION
-- =========================================

-- name: CreateProductBase :one
-- Membuat data dasar produk.
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
-- Membuat varian produk.
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


-- name: CreateVariantAttribute :one
-- Membuat atribut varian.
INSERT INTO variant_attributes (
    product_id,
    name
)
VALUES ($1, $2)
RETURNING id, product_id, name;


-- name: CreateVariantAttributeValue :one
-- Membuat nilai atribut varian.
INSERT INTO variant_attribute_values (
    variant_id,
    product_id,
    attribute_id,
    value
)
VALUES ($1, $2, $3, $4)
RETURNING
    id,
    variant_id,
    product_id,
    attribute_id,
    value;


-- =========================================
-- PRODUCT VARIANTS
-- =========================================

-- name: ListProductVariants :many
-- Mengambil seluruh varian aktif dari produk.
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
-- Mengambil atribut dari varian tertentu.
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
-- PRODUCT MANAGEMENT: UPDATE
-- =========================================

-- name: UpdateProductDetails :one
-- Memperbarui informasi dasar produk.
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
-- Memperbarui informasi dan stok varian.
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


-- =========================================
-- PRODUCT MANAGEMENT: DELETE
-- =========================================

-- name: DeleteProductVariant :execrows
-- Menghapus varian non-default.
DELETE FROM product_variants
WHERE id = $1
    AND product_id = $2
    AND is_default = FALSE;


-- name: DeleteProduct :execrows
-- Menghapus produk berdasarkan ID.
DELETE FROM products
WHERE id = $1;


-- =========================================
-- PRODUCT SEARCH: CATEGORY ID
-- =========================================

-- name: SearchActiveProducts :many
-- Pencarian produk berdasarkan keyword dan ID kategori.
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
        OR COALESCE(p.description, '') ILIKE
            '%' || sqlc.arg(keyword)::text || '%'
    )
    AND (
        sqlc.narg(category_id)::bigint IS NULL
        OR p.category_id = sqlc.narg(category_id)::bigint
    )
ORDER BY p.created_at DESC, p.id DESC
LIMIT sqlc.arg(limit_count)::integer
OFFSET sqlc.arg(offset_count)::integer;
