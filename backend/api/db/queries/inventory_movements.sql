
-- =========================================
-- INVENTORY MOVEMENT QUERIES
-- Hafiz Pintar Store
-- =========================================


-- =========================================
-- INVENTORY: CREATE MOVEMENT
-- =========================================

-- name: CreateInventoryMovement :one
-- Mencatat perubahan stok produk atau varian.
INSERT INTO inventory_movements (
    product_id,
    variant_id,
    movement_type,
    quantity,
    stock_before,
    stock_after,
    reference_type,
    reference_id,
    note
)
VALUES (
    sqlc.arg(product_id),
    sqlc.narg(variant_id),
    sqlc.arg(movement_type),
    sqlc.arg(quantity),
    sqlc.arg(stock_before),
    sqlc.arg(stock_after),
    sqlc.narg(reference_type),
    sqlc.narg(reference_id),
    sqlc.narg(note)
)
RETURNING
    id,
    product_id,
    variant_id,
    movement_type,
    quantity,
    stock_before,
    stock_after,
    reference_type,
    reference_id,
    note,
    created_at;


-- =========================================
-- INVENTORY: GET MOVEMENT
-- =========================================

-- name: GetInventoryMovementByID :one
SELECT
    id,
    product_id,
    variant_id,
    movement_type,
    quantity,
    stock_before,
    stock_after,
    reference_type,
    reference_id,
    note,
    created_at
FROM inventory_movements
WHERE id = $1
LIMIT 1;


-- =========================================
-- INVENTORY: PRODUCT HISTORY
-- =========================================

-- name: ListProductInventoryMovements :many
SELECT
    id,
    product_id,
    variant_id,
    movement_type,
    quantity,
    stock_before,
    stock_after,
    reference_type,
    reference_id,
    note,
    created_at
FROM inventory_movements
WHERE product_id = $1
ORDER BY created_at DESC, id DESC
LIMIT $2
OFFSET $3;


-- name: CountProductInventoryMovements :one
SELECT COUNT(*)
FROM inventory_movements
WHERE product_id = $1;


-- =========================================
-- INVENTORY: VARIANT HISTORY
-- =========================================

-- name: ListVariantInventoryMovements :many
SELECT
    id,
    product_id,
    variant_id,
    movement_type,
    quantity,
    stock_before,
    stock_after,
    reference_type,
    reference_id,
    note,
    created_at
FROM inventory_movements
WHERE variant_id = $1
ORDER BY created_at DESC, id DESC
LIMIT $2
OFFSET $3;


-- name: CountVariantInventoryMovements :one
SELECT COUNT(*)
FROM inventory_movements
WHERE variant_id = $1;


-- =========================================
-- INVENTORY: GLOBAL HISTORY
-- =========================================

-- name: ListAllInventoryMovements :many
SELECT
    id,
    product_id,
    variant_id,
    movement_type,
    quantity,
    stock_before,
    stock_after,
    reference_type,
    reference_id,
    note,
    created_at
FROM inventory_movements
ORDER BY created_at DESC, id DESC
LIMIT $1
OFFSET $2;


-- name: CountAllInventoryMovements :one
SELECT COUNT(*)
FROM inventory_movements;


-- =========================================
-- INVENTORY: FILTER BY MOVEMENT TYPE
-- =========================================

-- name: ListInventoryMovementsByType :many
SELECT
    id,
    product_id,
    variant_id,
    movement_type,
    quantity,
    stock_before,
    stock_after,
    reference_type,
    reference_id,
    note,
    created_at
FROM inventory_movements
WHERE movement_type = $1
ORDER BY created_at DESC, id DESC
LIMIT $2
OFFSET $3;


-- =========================================
-- INVENTORY: PRODUCT SUMMARY
-- =========================================

-- name: GetProductInventorySummary :one
SELECT
    product_id,

    COUNT(*)::bigint AS total_movements,

    COALESCE(
        SUM(
            CASE
                WHEN movement_type = 'in'
                THEN quantity
                ELSE 0
            END
        ),
        0
    )::bigint AS total_stock_in,

    COALESCE(
        SUM(
            CASE
                WHEN movement_type = 'out'
                THEN quantity
                ELSE 0
            END
        ),
        0
    )::bigint AS total_stock_out,

    COALESCE(
        SUM(
            CASE
                WHEN movement_type = 'adjustment'
                     AND stock_after > stock_before
                THEN stock_after - stock_before
                ELSE 0
            END
        ),
        0
    )::bigint AS adjustment_in,

    COALESCE(
        SUM(
            CASE
                WHEN movement_type = 'adjustment'
                     AND stock_after < stock_before
                THEN stock_before - stock_after
                ELSE 0
            END
        ),
        0
    )::bigint AS adjustment_out

FROM inventory_movements
WHERE product_id = $1
GROUP BY product_id;


-- =========================================
-- INVENTORY: VARIANT SUMMARY
-- =========================================

-- name: GetVariantInventorySummary :one
SELECT
    variant_id,

    COUNT(*)::bigint AS total_movements,

    COALESCE(
        SUM(
            CASE
                WHEN movement_type = 'in'
                THEN quantity
                ELSE 0
            END
        ),
        0
    )::bigint AS total_stock_in,

    COALESCE(
        SUM(
            CASE
                WHEN movement_type = 'out'
                THEN quantity
                ELSE 0
            END
        ),
        0
    )::bigint AS total_stock_out,

    COALESCE(
        SUM(
            CASE
                WHEN movement_type = 'adjustment'
                     AND stock_after > stock_before
                THEN stock_after - stock_before
                ELSE 0
            END
        ),
        0
    )::bigint AS adjustment_in,

    COALESCE(
        SUM(
            CASE
                WHEN movement_type = 'adjustment'
                     AND stock_after < stock_before
                THEN stock_before - stock_after
                ELSE 0
            END
        ),
        0
    )::bigint AS adjustment_out

FROM inventory_movements
WHERE variant_id = $1
GROUP BY variant_id;
