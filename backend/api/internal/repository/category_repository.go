package repository

import (
	"context"

	"hafiz-pintar-store/backend/api/internal/database/sqlc"
)

type CategoryRepository struct {
	queries *sqlc.Queries
}

// Constructor category repository.
func NewCategoryRepository(
	queries *sqlc.Queries,
) *CategoryRepository {
	return &CategoryRepository{
		queries: queries,
	}
}

// ListActive mengambil seluruh kategori aktif.
func (r *CategoryRepository) ListActive(
	ctx context.Context,
) ([]sqlc.Category, error) {
	return r.queries.ListActiveCategories(ctx)
}

// GetActiveBySlug mengambil kategori aktif berdasarkan slug.
func (r *CategoryRepository) GetActiveBySlug(
	ctx context.Context,
	slug string,
) (sqlc.Category, error) {
	return r.queries.GetActiveCategoryBySlug(ctx, slug)
}
