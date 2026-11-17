
-- +goose Up

ALTER TABLE orders
ADD COLUMN courier_name VARCHAR(100),
ADD COLUMN tracking_number VARCHAR(100),
ADD COLUMN shipped_at TIMESTAMPTZ,
ADD COLUMN completed_at TIMESTAMPTZ;

CREATE INDEX idx_orders_tracking_number
ON orders(tracking_number)
WHERE tracking_number IS NOT NULL;

-- +goose Down

DROP INDEX IF EXISTS idx_orders_tracking_number;

ALTER TABLE orders
DROP COLUMN IF EXISTS completed_at,
DROP COLUMN IF EXISTS shipped_at,
DROP COLUMN IF EXISTS tracking_number,
DROP COLUMN IF EXISTS courier_name;
