package service

import (
	"context"
	"errors"
	"strings"

	"hafiz-pintar-store/backend/api/internal/repository"
)

var (
	ErrInvalidAdminProductID     = errors.New("invalid product ID")
	ErrInvalidAdminProductStatus = errors.New("invalid product status")
	ErrInvalidAdminProductData   = errors.New("product name and slug are required")
	ErrInvalidAdminVariant       = errors.New("variant data is invalid")
)

type AdminProductService struct {
	repository        *repository.AdminProductRepository
	productRepository *repository.ProductRepository
}

type AdminProductDetail struct {
	Product  repository.AdminProduct           `json:"product"`
	Variants []repository.ProductVariantDetail `json:"variants"`
}

func NewAdminProductService(
	repo *repository.AdminProductRepository,
	productRepo *repository.ProductRepository,
) *AdminProductService {
	return &AdminProductService{
		repository:        repo,
		productRepository: productRepo,
	}
}

func (s *AdminProductService) List(
	ctx context.Context,
	limit, offset int32,
	search, status string,
) ([]repository.AdminProduct, error) {
	if limit < 1 || limit > 100 {
		limit = 20
	}

	if offset < 0 {
		offset = 0
	}

	status = strings.ToLower(strings.TrimSpace(status))

	if status != "" &&
		status != "active" &&
		status != "inactive" &&
		status != "archived" {
		return nil, ErrInvalidAdminProductStatus
	}

	return s.repository.List(
		ctx,
		limit,
		offset,
		strings.TrimSpace(search),
		status,
	)
}

func (s *AdminProductService) GetByID(
	ctx context.Context,
	id int64,
) (AdminProductDetail, error) {
	if id < 1 {
		return AdminProductDetail{}, ErrInvalidAdminProductID
	}

	product, err := s.repository.GetByID(ctx, id)
	if err != nil {
		return AdminProductDetail{}, err
	}

	variants, err := s.productRepository.ListVariantsWithAttributes(
		ctx,
		id,
	)
	if err != nil {
		return AdminProductDetail{}, err
	}

	return AdminProductDetail{
		Product:  product,
		Variants: variants,
	}, nil
}

func (s *AdminProductService) UpdateDetails(
	ctx context.Context,
	id int64,
	categoryID *int64,
	name, slug string,
	description *string,
) error {
	if id < 1 {
		return ErrInvalidAdminProductID
	}

	name = strings.TrimSpace(name)
	slug = strings.ToLower(strings.TrimSpace(slug))

	if name == "" || slug == "" {
		return ErrInvalidAdminProductData
	}

	if categoryID != nil && *categoryID < 1 {
		return ErrInvalidAdminProductData
	}

	if description != nil {
		value := strings.TrimSpace(*description)
		description = &value
	}

	return s.repository.UpdateDetails(
		ctx,
		id,
		categoryID,
		name,
		slug,
		description,
	)
}

func (s *AdminProductService) UpdateVariant(
	ctx context.Context,
	productID, variantID int64,
	name string,
	price int64,
	stockQty, weightGram int32,
) error {
	if productID < 1 || variantID < 1 {
		return ErrInvalidAdminProductID
	}

	name = strings.TrimSpace(name)

	if name == "" || price < 1 || stockQty < 0 || weightGram < 0 {
		return ErrInvalidAdminVariant
	}

	_, err := s.productRepository.UpdateVariant(
		ctx,
		productID,
		variantID,
		name,
		price,
		stockQty,
		weightGram,
	)

	return err
}

func (s *AdminProductService) DeleteVariant(
	ctx context.Context,
	productID, variantID int64,
) error {
	if productID < 1 || variantID < 1 {
		return ErrInvalidAdminProductID
	}

	return s.productRepository.DeleteVariant(
		ctx,
		productID,
		variantID,
	)
}

func (s *AdminProductService) UpdateStatus(
	ctx context.Context,
	id int64,
	status string,
) error {
	if id < 1 {
		return ErrInvalidAdminProductID
	}

	status = strings.ToLower(strings.TrimSpace(status))

	switch status {
	case "active", "inactive", "archived":
	default:
		return ErrInvalidAdminProductStatus
	}

	return s.repository.UpdateStatus(ctx, id, status)
}

func (s *AdminProductService) Archive(
	ctx context.Context,
	id int64,
) error {
	if id < 1 {
		return ErrInvalidAdminProductID
	}

	return s.repository.Archive(ctx, id)
}
