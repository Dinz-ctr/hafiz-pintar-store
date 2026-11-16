package repository

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

// =========================================
// ERRORS
// =========================================

var ErrAdminProductNotFound = errors.New("product not found")

// =========================================
// REPOSITORY
// =========================================

type AdminProductRepository struct {
	pool *pgxpool.Pool
}

// =========================================
// RESPONSE STRUCT
// =========================================

type AdminProduct struct {
	ID           int64   `json:"id"`
	CategoryID   *int64  `json:"category_id"`
	CategoryName *string `json:"category_name"`
	SKU          string  `json:"sku"`
	Name         string  `json:"name"`
	Slug         string  `json:"slug"`
	Description  *string `json:"description"`
	Price        int64   `json:"price"`
	StockQty     int32   `json:"stock_qty"`
	WeightGram   int32   `json:"weight_gram"`
	Status       string  `json:"status"`
}

// =========================================
// CONSTRUCTOR
// =========================================

func NewAdminProductRepository(
	pool *pgxpool.Pool,
) *AdminProductRepository {
	return &AdminProductRepository{
		pool: pool,
	}
}

// =========================================
// LIST PRODUCTS
// GET /api/v1/admin/products
// =========================================

func (r *AdminProductRepository) List(
	ctx context.Context,
	limit int32,
	offset int32,
	search string,
	status string,
) ([]AdminProduct, error) {

	rows, err := r.pool.Query(ctx, `
		SELECT
			p.id,
			p.category_id,
			c.name,
			p.sku,
			p.name,
			p.slug,
			p.description,
			p.price,
			p.stock_qty,
			p.weight_gram,
			p.status::text
		FROM products p
		LEFT JOIN categories c
			ON c.id = p.category_id
		WHERE
			(
				$1 = ''
				OR p.name ILIKE '%' || $1 || '%'
				OR p.sku ILIKE '%' || $1 || '%'
			)
			AND (
				$2 = ''
				OR p.status::text = $2
			)
		ORDER BY
			p.created_at DESC,
			p.id DESC
		LIMIT $3
		OFFSET $4
	`,
		search,
		status,
		limit,
		offset,
	)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	products := make([]AdminProduct, 0)

	for rows.Next() {
		var product AdminProduct

		var categoryID pgtype.Int8
		var categoryName pgtype.Text
		var description pgtype.Text

		err := rows.Scan(
			&product.ID,
			&categoryID,
			&categoryName,
			&product.SKU,
			&product.Name,
			&product.Slug,
			&description,
			&product.Price,
			&product.StockQty,
			&product.WeightGram,
			&product.Status,
		)

		if err != nil {
			return nil, err
		}

		if categoryID.Valid {
			value := categoryID.Int64
			product.CategoryID = &value
		}

		if categoryName.Valid {
			value := categoryName.String
			product.CategoryName = &value
		}

		if description.Valid {
			value := description.String
			product.Description = &value
		}

		products = append(products, product)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return products, nil
}

// =========================================
// GET PRODUCT BY ID
// GET /api/v1/admin/products/{id}
// =========================================

func (r *AdminProductRepository) GetByID(
	ctx context.Context,
	id int64,
) (AdminProduct, error) {

	var product AdminProduct

	var categoryID pgtype.Int8
	var categoryName pgtype.Text
	var description pgtype.Text

	err := r.pool.QueryRow(ctx, `
		SELECT
			p.id,
			p.category_id,
			c.name,
			p.sku,
			p.name,
			p.slug,
			p.description,
			p.price,
			p.stock_qty,
			p.weight_gram,
			p.status::text
		FROM products p
		LEFT JOIN categories c
			ON c.id = p.category_id
		WHERE p.id = $1
	`, id).Scan(
		&product.ID,
		&categoryID,
		&categoryName,
		&product.SKU,
		&product.Name,
		&product.Slug,
		&description,
		&product.Price,
		&product.StockQty,
		&product.WeightGram,
		&product.Status,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return AdminProduct{}, ErrAdminProductNotFound
	}

	if err != nil {
		return AdminProduct{}, err
	}

	if categoryID.Valid {
		value := categoryID.Int64
		product.CategoryID = &value
	}

	if categoryName.Valid {
		value := categoryName.String
		product.CategoryName = &value
	}

	if description.Valid {
		value := description.String
		product.Description = &value
	}

	return product, nil
}

// =========================================
// UPDATE PRODUCT DETAILS
// PUT /api/v1/admin/products/{id}
// =========================================

func (r *AdminProductRepository) UpdateDetails(
	ctx context.Context,
	id int64,
	categoryID *int64,
	name string,
	slug string,
	description *string,
) error {

	tag, err := r.pool.Exec(ctx, `
		UPDATE products
		SET
			category_id = $2,
			name = $3,
			slug = $4,
			description = $5,
			updated_at = NOW()
		WHERE id = $1
	`,
		id,
		categoryID,
		name,
		slug,
		description,
	)

	if err != nil {
		return err
	}

	if tag.RowsAffected() == 0 {
		return ErrAdminProductNotFound
	}

	return nil
}

// =========================================
// UPDATE PRODUCT STATUS
// PATCH /api/v1/admin/products/{id}/status
// =========================================

func (r *AdminProductRepository) UpdateStatus(
	ctx context.Context,
	id int64,
	status string,
) error {

	tag, err := r.pool.Exec(ctx, `
		UPDATE products
		SET
			status = $2,
			updated_at = NOW()
		WHERE id = $1
	`,
		id,
		status,
	)

	if err != nil {
		return err
	}

	if tag.RowsAffected() == 0 {
		return ErrAdminProductNotFound
	}

	return nil
}

// =========================================
// ARCHIVE PRODUCT
// DELETE /api/v1/admin/products/{id}
// =========================================

func (r *AdminProductRepository) Archive(
	ctx context.Context,
	id int64,
) error {

	return r.UpdateStatus(
		ctx,
		id,
		"archived",
	)
}
