
-- +goose Up

-- 1. Product variants
CREATE TABLE product_variants (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,

    product_id BIGINT NOT NULL
        REFERENCES products(id) ON DELETE CASCADE,

    sku VARCHAR(100) NOT NULL UNIQUE,
    name VARCHAR(150) NOT NULL DEFAULT 'Default',

    price BIGINT NOT NULL CHECK (price >= 0),
    stock_qty INTEGER NOT NULL DEFAULT 0
        CHECK (stock_qty >= 0),

    weight_gram INTEGER NOT NULL DEFAULT 0
        CHECK (weight_gram >= 0),

    is_default BOOLEAN NOT NULL DEFAULT FALSE,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    UNIQUE (id, product_id)
);

CREATE UNIQUE INDEX uq_product_default_variant
ON product_variants(product_id)
WHERE is_default = TRUE;

CREATE INDEX idx_product_variants_product_id
ON product_variants(product_id);

-- 2. Variant attributes
CREATE TABLE variant_attributes (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,

    product_id BIGINT NOT NULL
        REFERENCES products(id) ON DELETE CASCADE,

    name VARCHAR(100) NOT NULL,

    UNIQUE (id, product_id),
    UNIQUE (product_id, name)
);

-- 3. Variant attribute values
CREATE TABLE variant_attribute_values (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,

    variant_id BIGINT NOT NULL,
    product_id BIGINT NOT NULL,
    attribute_id BIGINT NOT NULL,

    value VARCHAR(150) NOT NULL,

    FOREIGN KEY (variant_id, product_id)
        REFERENCES product_variants(id, product_id)
        ON DELETE CASCADE,

    FOREIGN KEY (attribute_id, product_id)
        REFERENCES variant_attributes(id, product_id)
        ON DELETE CASCADE,

    UNIQUE (variant_id, attribute_id)
);

CREATE INDEX idx_variant_attribute_values_variant
ON variant_attribute_values(variant_id);

-- 4. Migrate existing products into default variants
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
SELECT
    id,
    sku,
    'Default',
    price,
    stock_qty,
    weight_gram,
    TRUE,
    (status = 'active')
FROM products;

-- +goose Down

DROP TABLE IF EXISTS variant_attribute_values;
DROP TABLE IF EXISTS variant_attributes;
DROP TABLE IF EXISTS product_variants;
