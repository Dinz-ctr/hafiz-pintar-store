package service

import (
	"context"
	"errors"
	"strings"

	"hafiz-pintar-store/backend/api/internal/database/sqlc"
	"hafiz-pintar-store/backend/api/internal/repository"
)

var ErrInvalidCategorySlug = errors.New("category slug is required")

type CategoryService struct {
	repository *repository.CategoryRepository
}

func NewCategoryService(
	repo *repository.CategoryRepository,
) *CategoryService {
	return &CategoryService{
		repository: repo,
	}
}

// ListActive mengambil seluruh kategori aktif.
func (s *CategoryService) ListActive(
	ctx context.Context,
) ([]sqlc.Category, error) {
	return s.repository.ListActive(ctx)
}

// GetActiveBySlug mengambil kategori aktif berdasarkan slug.
func (s *CategoryService) GetActiveBySlug(
	ctx context.Context,
	slug string,
) (sqlc.Category, error) {
	slug = strings.TrimSpace(strings.ToLower(slug))

	if slug == "" {
		return sqlc.Category{}, ErrInvalidCategorySlug
	}

	return s.repository.GetActiveBySlug(ctx, slug)
}
