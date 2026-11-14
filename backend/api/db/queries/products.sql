
-- name: ListActiveProducts :many
SELECT
    p.id,
    p.sku,
    p.name,
    p.slug,
    p.description,
    p.price,
    p.stock_qty,
    p.weight_gram,
    p.category_id,
    c.name AS category_name,
    p.created_at
FROM products p
LEFT JOIN categories c ON c.id = p.category_id
WHERE p.status = 'active'
ORDER BY p.created_at DESC
LIMIT $1 OFFSET $2;

-- name: GetProductBySlug :one
SELECT
    p.id,
    p.sku,
    p.name,
    p.slug,
    p.description,
    p.price,
    p.stock_qty,
    p.weight_gram,
    p.category_id,
    c.name AS category_name,
    p.created_at
FROM products p
LEFT JOIN categories c ON c.id = p.category_id
WHERE p.slug = $1
  AND p.status = 'active'
LIMIT 1;

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
ORDER BY is_primary DESC, sort_order ASC, id ASC;
