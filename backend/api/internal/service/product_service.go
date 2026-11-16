package service

import (
	"context"
	"errors"
	"strings"

	"hafiz-pintar-store/backend/api/internal/database/sqlc"
	"hafiz-pintar-store/backend/api/internal/repository"
)

var (
	ErrInvalidProductSlug  = errors.New("product slug is required")
	ErrInvalidProductName  = errors.New("product name is required")
	ErrInvalidVariants     = errors.New("product must have at least one variant")
	ErrInvalidVariant      = errors.New("variant data is invalid")
	ErrInvalidAttribute    = errors.New("variant attribute is invalid")
	ErrDuplicateVariantSKU = errors.New("duplicate variant SKU")
)

type ProductService struct {
	repository *repository.ProductRepository
}

func NewProductService(repo *repository.ProductRepository) *ProductService {
	return &ProductService{
		repository: repo,
	}
}

// List mengambil daftar produk aktif beserta varian default.
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

// GetBySlug mengambil detail produk berdasarkan slug.
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

// Create membuat produk beserta seluruh variannya.
func (s *ProductService) Create(
	ctx context.Context,
	input repository.CreateProductInput,
) (sqlc.CreateProductBaseRow, error) {
	input.Name = strings.TrimSpace(input.Name)

	if input.Name == "" {
		return sqlc.CreateProductBaseRow{}, ErrInvalidProductName
	}

	input.Slug = strings.TrimSpace(strings.ToLower(input.Slug))

	if input.Slug == "" {
		return sqlc.CreateProductBaseRow{}, ErrInvalidProductSlug
	}

	if len(input.Variants) == 0 {
		return sqlc.CreateProductBaseRow{}, ErrInvalidVariants
	}

	usedSKU := make(map[string]bool)

	for i := range input.Variants {
		variant := &input.Variants[i]

		variant.Sku = strings.TrimSpace(variant.Sku)
		variant.Name = strings.TrimSpace(variant.Name)

		if variant.Sku == "" ||
			variant.Name == "" ||
			variant.Price < 0 ||
			variant.StockQty < 0 ||
			variant.WeightGram < 0 {
			return sqlc.CreateProductBaseRow{}, ErrInvalidVariant
		}

		skuKey := strings.ToUpper(variant.Sku)

		if usedSKU[skuKey] {
			return sqlc.CreateProductBaseRow{}, ErrDuplicateVariantSKU
		}

		usedSKU[skuKey] = true

		normalizedAttributes := make(map[string]string)

		for name, value := range variant.Attributes {
			name = strings.TrimSpace(name)
			value = strings.TrimSpace(value)

			if name == "" || value == "" {
				return sqlc.CreateProductBaseRow{}, ErrInvalidAttribute
			}

			normalizedAttributes[name] = value
		}

		variant.Attributes = normalizedAttributes
	}

	// Repository menangani transaksi PostgreSQL.
	return s.repository.Create(ctx, input)
}

// GetVariantsBySlug mengambil seluruh varian aktif
// beserta atributnya berdasarkan slug produk.
func (s *ProductService) GetVariantsBySlug(
	ctx context.Context,
	slug string,
) ([]repository.ProductVariantDetail, error) {
	slug = strings.TrimSpace(slug)

	if slug == "" {
		return nil, ErrInvalidProductSlug
	}

	product, err := s.repository.GetBySlug(ctx, slug)
	if err != nil {
		return nil, err
	}

	return s.repository.ListVariantsWithAttributes(
		ctx,
		product.ID,
	)
}

// GetImagesBySlug mengambil seluruh gambar produk berdasarkan slug.
func (s *ProductService) GetImagesBySlug(
	ctx context.Context,
	slug string,
) ([]sqlc.ProductImage, error) {
	slug = strings.TrimSpace(slug)

	if slug == "" {
		return nil, ErrInvalidProductSlug
	}

	product, err := s.repository.GetBySlug(ctx, slug)
	if err != nil {
		return nil, err
	}

	return s.repository.ListImagesByProductID(
		ctx,
		product.ID,
	)
}
