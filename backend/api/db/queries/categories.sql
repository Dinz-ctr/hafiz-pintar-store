
-- =========================================
-- CATEGORY QUERIES
-- Hafiz Pintar Store
-- =========================================


-- =========================================
-- CATEGORY: PUBLIC CATALOG
-- =========================================

-- name: ListActiveCategories :many
SELECT
    c.id,
    c.parent_id,
    c.name,
    c.slug,
    c.description,
    c.is_active,
    c.created_at,
    c.updated_at
FROM categories c
WHERE c.is_active = TRUE
ORDER BY c.name ASC;


-- name: GetActiveCategoryBySlug :one
SELECT
    c.id,
    c.parent_id,
    c.name,
    c.slug,
    c.description,
    c.is_active,
    c.created_at,
    c.updated_at
FROM categories c
WHERE c.slug = $1
    AND c.is_active = TRUE
LIMIT 1;


-- =========================================
-- CATEGORY: CHILD CATEGORIES
-- =========================================

-- name: ListActiveChildCategories :many
SELECT
    id,
    parent_id,
    name,
    slug,
    description,
    is_active,
    created_at,
    updated_at
FROM categories
WHERE parent_id = $1
    AND is_active = TRUE
ORDER BY name ASC;


-- =========================================
-- CATEGORY: ADMIN MANAGEMENT
-- =========================================

-- name: ListAllCategories :many
SELECT
    id,
    parent_id,
    name,
    slug,
    description,
    is_active,
    created_at,
    updated_at
FROM categories
ORDER BY created_at DESC, id DESC;


-- name: GetCategoryByID :one
SELECT
    id,
    parent_id,
    name,
    slug,
    description,
    is_active,
    created_at,
    updated_at
FROM categories
WHERE id = $1
LIMIT 1;


-- name: CreateCategory :one
INSERT INTO categories (
    parent_id,
    name,
    slug,
    description,
    is_active
)
VALUES ($1, $2, $3, $4, $5)
RETURNING
    id,
    parent_id,
    name,
    slug,
    description,
    is_active,
    created_at,
    updated_at;


-- name: UpdateCategory :one
UPDATE categories
SET
    parent_id = $2,
    name = $3,
    slug = $4,
    description = $5,
    is_active = $6,
    updated_at = NOW()
WHERE id = $1
RETURNING
    id,
    parent_id,
    name,
    slug,
    description,
    is_active,
    created_at,
    updated_at;


-- name: DeleteCategory :execrows
DELETE FROM categories
WHERE id = $1;


-- =========================================
-- CATEGORY: VALIDATION
-- =========================================

-- name: CountProductsByCategory :one
SELECT COUNT(*)
FROM products
WHERE category_id = $1;


-- name: CheckCategorySlugExists :one
-- Memeriksa apakah slug sudah digunakan.
SELECT EXISTS (
    SELECT 1
    FROM categories
    WHERE slug = $1
) AS slug_exists;


-- name: CheckCategorySlugExistsForUpdate :one
-- Memeriksa slug tanpa menghitung kategori yang sedang diedit.
SELECT EXISTS (
    SELECT 1
    FROM categories
    WHERE slug = $1
        AND id != $2
) AS slug_exists;


-- =========================================
-- CATEGORY: PARENT VALIDATION
-- =========================================

-- name: CheckCategoryParentExists :one
-- Memeriksa apakah kategori induk tersedia.
SELECT EXISTS (
    SELECT 1
    FROM categories
    WHERE id = $1
) AS parent_exists;


-- name: CheckCategoryParentExistsForUpdate :one
-- Memastikan kategori induk bukan dirinya sendiri.
SELECT EXISTS (
    SELECT 1
    FROM categories
    WHERE id = $1
        AND id != $2
) AS parent_exists;
