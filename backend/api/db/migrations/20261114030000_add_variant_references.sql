
-- +goose Up

-- =========================================
-- 1. ADD VARIANT REFERENCE TO ORDER ITEMS
-- =========================================

ALTER TABLE order_items
ADD COLUMN variant_id BIGINT;

-- Hubungkan transaksi lama ke default variant.
UPDATE order_items oi
SET variant_id = pv.id
FROM product_variants pv
WHERE pv.product_id = oi.product_id
    AND pv.is_default = TRUE
    AND oi.product_id IS NOT NULL;

-- Pastikan variant berasal dari product yang sama.
ALTER TABLE order_items
ADD CONSTRAINT fk_order_items_variant_product
FOREIGN KEY (variant_id, product_id)
REFERENCES product_variants (id, product_id)
ON DELETE SET NULL (variant_id);

CREATE INDEX idx_order_items_variant
ON order_items(variant_id);


-- =========================================
-- 2. ADD VARIANT REFERENCE TO INVENTORY
-- =========================================

ALTER TABLE inventory_movements
ADD COLUMN variant_id BIGINT;

-- Hubungkan riwayat stok lama ke default variant.
UPDATE inventory_movements im
SET variant_id = pv.id
FROM product_variants pv
WHERE pv.product_id = im.product_id
    AND pv.is_default = TRUE;

-- Pastikan variant berasal dari product yang sama.
ALTER TABLE inventory_movements
ADD CONSTRAINT fk_inventory_movements_variant_product
FOREIGN KEY (variant_id, product_id)
REFERENCES product_variants (id, product_id)
ON DELETE RESTRICT;

CREATE INDEX idx_inventory_movements_variant_created
ON inventory_movements(variant_id, created_at DESC);


-- +goose Down

DROP INDEX IF EXISTS idx_inventory_movements_variant_created;

ALTER TABLE inventory_movements
DROP CONSTRAINT IF EXISTS fk_inventory_movements_variant_product;

ALTER TABLE inventory_movements
DROP COLUMN IF EXISTS variant_id;


DROP INDEX IF EXISTS idx_order_items_variant;

ALTER TABLE order_items
DROP CONSTRAINT IF EXISTS fk_order_items_variant_product;

ALTER TABLE order_items
DROP COLUMN IF EXISTS variant_id;
