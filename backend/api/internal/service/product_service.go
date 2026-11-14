package service

import (
	"context"
	"errors"
	"strings"

	"hafiz-pintar-store/backend/api/internal/database/sqlc"
	"hafiz-pintar-store/backend/api/internal/repository"
)

var ErrInvalidProductSlug = errors.New("product slug is required")

type ProductService struct {
	repository *repository.ProductRepository
}

func NewProductService(repo *repository.ProductRepository) *ProductService {
	return &ProductService{
		repository: repo,
	}
}

func (s *ProductService) List(
	ctx context.Context,
	limit int32,
	offset int32,
) ([]sqlc.ListActiveProductsRow, error) {
	if limit < 1 || limit > 100 {
		limit = 20
	}

	if offset < 0 {
		offset = 0
	}

	return s.repository.ListActive(ctx, limit, offset)
}

func (s *ProductService) GetBySlug(
	ctx context.Context,
	slug string,
) (sqlc.GetProductBySlugRow, error) {
	slug = strings.TrimSpace(slug)

	if slug == "" {
		return sqlc.GetProductBySlugRow{}, ErrInvalidProductSlug
	}

	return s.repository.GetBySlug(ctx, slug)
}
