package repository

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrAdminCategoryNotFound       = errors.New("category not found")
	ErrAdminCategoryParentNotFound = errors.New("parent category not found")
	ErrAdminCategoryCycle          = errors.New("category cannot be its own descendant")
	ErrAdminCategoryHasChildren    = errors.New("category has active child categories")
)

type AdminCategoryRepository struct {
	pool *pgxpool.Pool
}

type AdminCategory struct {
	ID           int64   `json:"id"`
	ParentID     *int64  `json:"parent_id"`
	ParentName   *string `json:"parent_name"`
	Name         string  `json:"name"`
	Slug         string  `json:"slug"`
	Description  *string `json:"description"`
	IsActive     bool    `json:"is_active"`
	ProductCount int64   `json:"product_count"`
	CreatedAt    any     `json:"created_at"`
	UpdatedAt    any     `json:"updated_at"`
}

func NewAdminCategoryRepository(pool *pgxpool.Pool) *AdminCategoryRepository {
	return &AdminCategoryRepository{pool: pool}
}

func scanAdminCategory(row pgx.Row) (AdminCategory, error) {
	var category AdminCategory
	var parentID pgtype.Int8
	var parentName pgtype.Text
	var description pgtype.Text

	err := row.Scan(
		&category.ID,
		&parentID,
		&parentName,
		&category.Name,
		&category.Slug,
		&description,
		&category.IsActive,
		&category.ProductCount,
		&category.CreatedAt,
		&category.UpdatedAt,
	)
	if err != nil {
		return AdminCategory{}, err
	}

	if parentID.Valid {
		value := parentID.Int64
		category.ParentID = &value
	}

	if parentName.Valid {
		value := parentName.String
		category.ParentName = &value
	}

	if description.Valid {
		value := description.String
		category.Description = &value
	}

	return category, nil
}

const adminCategorySelect = `
	SELECT
		c.id,
		c.parent_id,
		parent.name,
		c.name,
		c.slug,
		c.description,
		c.is_active,
		(
			SELECT COUNT(*)
			FROM products p
			WHERE p.category_id = c.id
		) AS product_count,
		c.created_at,
		c.updated_at
	FROM categories c
	LEFT JOIN categories parent ON parent.id = c.parent_id
`

func (r *AdminCategoryRepository) List(
	ctx context.Context,
	limit, offset int32,
	search, status string,
) ([]AdminCategory, error) {
	rows, err := r.pool.Query(ctx, adminCategorySelect+`
		WHERE
			($1 = '' OR c.name ILIKE '%' || $1 || '%' OR c.slug ILIKE '%' || $1 || '%')
			AND ($2 = '' OR c.is_active = ($2 = 'active'))
		ORDER BY c.created_at DESC, c.id DESC
		LIMIT $3 OFFSET $4
	`, search, status, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	categories := make([]AdminCategory, 0)

	for rows.Next() {
		category, err := scanAdminCategory(rows)
		if err != nil {
			return nil, err
		}
		categories = append(categories, category)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return categories, nil
}

func (r *AdminCategoryRepository) GetByID(
	ctx context.Context,
	id int64,
) (AdminCategory, error) {
	category, err := scanAdminCategory(
		r.pool.QueryRow(ctx, adminCategorySelect+` WHERE c.id = $1`, id),
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return AdminCategory{}, ErrAdminCategoryNotFound
	}

	return category, err
}

func (r *AdminCategoryRepository) ValidateParent(
	ctx context.Context,
	parentID int64,
	categoryID int64,
) error {
	var exists bool

	err := r.pool.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1 FROM categories WHERE id = $1
		)
	`, parentID).Scan(&exists)
	if err != nil {
		return err
	}

	if !exists {
		return ErrAdminCategoryParentNotFound
	}

	if categoryID == 0 {
		return nil
	}

	var isDescendant bool

	err = r.pool.QueryRow(ctx, `
		WITH RECURSIVE descendants AS (
			SELECT id
			FROM categories
			WHERE id = $1

			UNION ALL

			SELECT c.id
			FROM categories c
			INNER JOIN descendants d ON c.parent_id = d.id
		)
		SELECT EXISTS (
			SELECT 1 FROM descendants WHERE id = $2
		)
	`, categoryID, parentID).Scan(&isDescendant)

	if err != nil {
		return err
	}

	if isDescendant {
		return ErrAdminCategoryCycle
	}

	return nil
}

func (r *AdminCategoryRepository) Create(
	ctx context.Context,
	parentID *int64,
	name, slug string,
	description *string,
	isActive bool,
) (AdminCategory, error) {
	return scanAdminCategory(r.pool.QueryRow(ctx, `
		INSERT INTO categories (
			parent_id, name, slug, description, is_active
		)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING
			id,
			parent_id,
			NULL::VARCHAR AS parent_name,
			name,
			slug,
			description,
			is_active,
			0::BIGINT AS product_count,
			created_at,
			updated_at
	`, parentID, name, slug, description, isActive))
}

func (r *AdminCategoryRepository) UpdateDetails(
	ctx context.Context,
	id int64,
	parentID *int64,
	name, slug string,
	description *string,
) error {
	tag, err := r.pool.Exec(ctx, `
		UPDATE categories
		SET
			parent_id = $2,
			name = $3,
			slug = $4,
			description = $5,
			updated_at = NOW()
		WHERE id = $1
	`, id, parentID, name, slug, description)

	if err != nil {
		return err
	}

	if tag.RowsAffected() == 0 {
		return ErrAdminCategoryNotFound
	}

	return nil
}

func (r *AdminCategoryRepository) UpdateStatus(
	ctx context.Context,
	id int64,
	isActive bool,
) error {
	tag, err := r.pool.Exec(ctx, `
		UPDATE categories
		SET is_active = $2, updated_at = NOW()
		WHERE id = $1
	`, id, isActive)

	if err != nil {
		return err
	}

	if tag.RowsAffected() == 0 {
		return ErrAdminCategoryNotFound
	}

	return nil
}

func (r *AdminCategoryRepository) Archive(
	ctx context.Context,
	id int64,
) error {
	var hasActiveChildren bool

	err := r.pool.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1
			FROM categories
			WHERE parent_id = $1 AND is_active = TRUE
		)
	`, id).Scan(&hasActiveChildren)

	if err != nil {
		return err
	}

	if hasActiveChildren {
		return ErrAdminCategoryHasChildren
	}

	return r.UpdateStatus(ctx, id, false)
}
