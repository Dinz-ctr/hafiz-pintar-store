package repository

import (
	"context"

	"hafiz-pintar-store/backend/api/internal/database/sqlc"
)

type ProductRepository struct {
	queries *sqlc.Queries
}

func NewProductRepository(queries *sqlc.Queries) *ProductRepository {
	return &ProductRepository{
		queries: queries,
	}
}

func (r *ProductRepository) ListActive(
	ctx context.Context,
	limit int32,
	offset int32,
) ([]sqlc.ListActiveProductsRow, error) {
	return r.queries.ListActiveProducts(ctx, sqlc.ListActiveProductsParams{
		Limit:  limit,
		Offset: offset,
	})
}

func (r *ProductRepository) GetBySlug(
	ctx context.Context,
	slug string,
) (sqlc.GetProductBySlugRow, error) {
	return r.queries.GetProductBySlug(ctx, slug)
}
