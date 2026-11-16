
-- =========================================
-- SHOPPING CART QUERIES
-- =========================================


-- name: GetOrCreateCart :one
INSERT INTO carts (customer_id)
VALUES ($1)
ON CONFLICT (customer_id) DO UPDATE
SET updated_at = carts.updated_at
RETURNING *;


-- name: GetCartByCustomerID :one
SELECT *
FROM carts
WHERE customer_id = $1
LIMIT 1;


-- =========================================
-- CART ITEMS: READ
-- =========================================

-- name: GetCartItems :many
SELECT
    ci.id,
    ci.cart_id,
    ci.product_id,
    ci.variant_id,
    ci.quantity,
    ci.created_at,
    ci.updated_at,

    p.name AS product_name,
    p.slug AS product_slug,

    pv.sku AS variant_sku,
    pv.name AS variant_name,
    pv.price AS unit_price,
    pv.stock_qty,
    pv.weight_gram,
    pv.is_active AS variant_is_active,

    (pv.price * ci.quantity)::BIGINT AS subtotal

FROM cart_items ci
JOIN products p
    ON p.id = ci.product_id
JOIN product_variants pv
    ON pv.id = ci.variant_id
    AND pv.product_id = ci.product_id

WHERE ci.cart_id = $1

ORDER BY ci.created_at DESC;


-- =========================================
-- CART ITEMS: CREATE
-- =========================================

-- name: AddCartItem :one
INSERT INTO cart_items (
    cart_id,
    product_id,
    variant_id,
    quantity
)
SELECT
    $1,
    $2,
    pv.id,
    $4
FROM product_variants pv
JOIN products p
    ON p.id = pv.product_id

WHERE pv.id = $3
    AND pv.product_id = $2
    AND pv.is_active = TRUE
    AND pv.stock_qty >= $4
    AND p.status = 'active'

ON CONFLICT (cart_id, variant_id)

DO UPDATE SET
    quantity = cart_items.quantity + EXCLUDED.quantity,
    updated_at = NOW()

WHERE cart_items.quantity + EXCLUDED.quantity <= (
    SELECT stock_qty
    FROM product_variants
    WHERE id = EXCLUDED.variant_id
        AND product_id = EXCLUDED.product_id
        AND is_active = TRUE
)

RETURNING *;


-- =========================================
-- CART ITEMS: UPDATE
-- =========================================

-- name: UpdateCartItemQuantity :one
UPDATE cart_items ci

SET
    quantity = $3,
    updated_at = NOW()

FROM product_variants pv
JOIN products p
    ON p.id = pv.product_id

WHERE ci.id = $1
    AND ci.cart_id = $2
    AND pv.id = ci.variant_id
    AND pv.product_id = ci.product_id
    AND pv.is_active = TRUE
    AND pv.stock_qty >= $3
    AND p.status = 'active'

RETURNING
    ci.id,
    ci.cart_id,
    ci.product_id,
    ci.variant_id,
    ci.quantity,
    ci.created_at,
    ci.updated_at;


-- =========================================
-- CART ITEMS: DELETE
-- =========================================

-- name: RemoveCartItem :execrows
DELETE FROM cart_items
WHERE id = $1
    AND cart_id = $2;


-- =========================================
-- CART: CLEAR
-- =========================================

-- name: ClearCart :execrows
DELETE FROM cart_items
WHERE cart_id = $1;
