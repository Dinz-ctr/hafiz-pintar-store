
-- =========================================
-- ORDER ITEM QUERIES
-- Hafiz Pintar Store
-- =========================================


-- =========================================
-- CUSTOMER: CREATE ORDER ITEMS
-- =========================================

-- name: CreateOrderItem :one
INSERT INTO order_items (
    order_id,
    product_id,
    variant_id,
    product_sku,
    product_name,
    quantity,
    unit_price,
    subtotal
)
VALUES (
    sqlc.arg(order_id),
    sqlc.narg(product_id),
    sqlc.narg(variant_id),
    sqlc.arg(product_sku),
    sqlc.arg(product_name),
    sqlc.arg(quantity),
    sqlc.arg(unit_price),
    sqlc.arg(subtotal)
)
RETURNING
    id,
    order_id,
    product_id,
    variant_id,
    product_sku,
    product_name,
    quantity,
    unit_price,
    subtotal,
    created_at;


-- =========================================
-- CUSTOMER: LIST ORDER ITEMS
-- =========================================

-- name: ListOrderItems :many
SELECT
    id,
    order_id,
    product_id,
    variant_id,
    product_sku,
    product_name,
    quantity,
    unit_price,
    subtotal,
    created_at
FROM order_items
WHERE order_id = $1
ORDER BY id ASC;


-- =========================================
-- CUSTOMER: GET ORDER ITEM
-- =========================================

-- name: GetOrderItemByID :one
SELECT
    id,
    order_id,
    product_id,
    variant_id,
    product_sku,
    product_name,
    quantity,
    unit_price,
    subtotal,
    created_at
FROM order_items
WHERE id = $1
    AND order_id = $2
LIMIT 1;


-- =========================================
-- ADMIN: UPDATE ORDER ITEM
-- =========================================

-- name: UpdateOrderItemQuantity :one
UPDATE order_items
SET
    quantity = $3,
    subtotal = unit_price * $3
WHERE id = $1
    AND order_id = $2
RETURNING
    id,
    order_id,
    product_id,
    variant_id,
    product_sku,
    product_name,
    quantity,
    unit_price,
    subtotal,
    created_at;


-- =========================================
-- ADMIN: DELETE ORDER ITEM
-- =========================================

-- name: DeleteOrderItem :execrows
DELETE FROM order_items
WHERE id = $1
    AND order_id = $2;


-- =========================================
-- ORDER ITEM STATISTICS
-- =========================================

-- name: CountOrderItems :one
SELECT COUNT(*)
FROM order_items
WHERE order_id = $1;


-- name: CalculateOrderItemsTotal :one
SELECT COALESCE(SUM(subtotal), 0)::bigint
FROM order_items
WHERE order_id = $1;


-- name: CalculateOrderItemsQuantity :one
SELECT COALESCE(SUM(quantity), 0)::bigint
FROM order_items
WHERE order_id = $1;


-- =========================================
-- ORDER ITEM: VARIANT HISTORY
-- =========================================

-- name: ListOrderItemsByVariant :many
SELECT
    id,
    order_id,
    product_id,
    variant_id,
    product_sku,
    product_name,
    quantity,
    unit_price,
    subtotal,
    created_at
FROM order_items
WHERE variant_id = $1
ORDER BY created_at DESC, id DESC
LIMIT $2
OFFSET $3;


-- name: CountOrderItemsByVariant :one
SELECT COUNT(*)
FROM order_items
WHERE variant_id = $1;
