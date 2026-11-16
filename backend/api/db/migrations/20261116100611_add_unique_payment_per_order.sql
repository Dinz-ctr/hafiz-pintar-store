-- +goose Up

ALTER TABLE payments
ADD CONSTRAINT payments_order_id_unique UNIQUE (order_id);

-- +goose Down

ALTER TABLE payments
DROP CONSTRAINT IF EXISTS payments_order_id_unique;