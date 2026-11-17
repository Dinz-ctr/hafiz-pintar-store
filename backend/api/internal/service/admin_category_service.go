package service

import (
	"context"
	"errors"
	"strings"

	"hafiz-pintar-store/backend/api/internal/repository"
)

var (
	ErrInvalidAdminCategoryID     = errors.New("invalid category ID")
	ErrInvalidAdminCategoryData   = errors.New("category name and slug are required")
	ErrInvalidAdminCategoryStatus = errors.New("invalid category status")
)

type AdminCategoryService struct {
	repository *repository.AdminCategoryRepository
}

func NewAdminCategoryService(
	repo *repository.AdminCategoryRepository,
) *AdminCategoryService {
	return &AdminCategoryService{repository: repo}
}

func (s *AdminCategoryService) List(
	ctx context.Context,
	limit, offset int32,
	search, status string,
) ([]repository.AdminCategory, error) {
	if limit < 1 || limit > 100 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}

	status = strings.ToLower(strings.TrimSpace(status))
	if status != "" && status != "active" && status != "inactive" {
		return nil, ErrInvalidAdminCategoryStatus
	}

	return s.repository.List(
		ctx,
		limit,
		offset,
		strings.TrimSpace(search),
		status,
	)
}

func (s *AdminCategoryService) GetByID(
	ctx context.Context,
	id int64,
) (repository.AdminCategory, error) {
	if id < 1 {
		return repository.AdminCategory{}, ErrInvalidAdminCategoryID
	}
	return s.repository.GetByID(ctx, id)
}

func normalizeCategoryInput(
	parentID *int64,
	name, slug string,
	description *string,
) (*int64, string, string, *string, error) {
	name = strings.TrimSpace(name)
	slug = strings.ToLower(strings.TrimSpace(slug))

	if name == "" || slug == "" {
		return nil, "", "", nil, ErrInvalidAdminCategoryData
	}

	if parentID != nil && *parentID < 1 {
		return nil, "", "", nil, ErrInvalidAdminCategoryData
	}

	if description != nil {
		value := strings.TrimSpace(*description)
		if value == "" {
			description = nil
		} else {
			description = &value
		}
	}

	return parentID, name, slug, description, nil
}

func (s *AdminCategoryService) Create(
	ctx context.Context,
	parentID *int64,
	name, slug string,
	description *string,
	isActive bool,
) (repository.AdminCategory, error) {
	parentID, name, slug, description, err := normalizeCategoryInput(
		parentID, name, slug, description,
	)
	if err != nil {
		return repository.AdminCategory{}, err
	}

	if parentID != nil {
		if err := s.repository.ValidateParent(ctx, *parentID, 0); err != nil {
			return repository.AdminCategory{}, err
		}
	}

	return s.repository.Create(
		ctx, parentID, name, slug, description, isActive,
	)
}

func (s *AdminCategoryService) UpdateDetails(
	ctx context.Context,
	id int64,
	parentID *int64,
	name, slug string,
	description *string,
) error {
	if id < 1 {
		return ErrInvalidAdminCategoryID
	}

	parentID, name, slug, description, err := normalizeCategoryInput(
		parentID, name, slug, description,
	)
	if err != nil {
		return err
	}

	if parentID != nil {
		if *parentID == id {
			return repository.ErrAdminCategoryCycle
		}
		if err := s.repository.ValidateParent(ctx, *parentID, id); err != nil {
			return err
		}
	}

	return s.repository.UpdateDetails(
		ctx, id, parentID, name, slug, description,
	)
}

func (s *AdminCategoryService) UpdateStatus(
	ctx context.Context,
	id int64,
	isActive bool,
) error {
	if id < 1 {
		return ErrInvalidAdminCategoryID
	}
	return s.repository.UpdateStatus(ctx, id, isActive)
}

func (s *AdminCategoryService) Archive(
	ctx context.Context,
	id int64,
) error {
	if id < 1 {
		return ErrInvalidAdminCategoryID
	}
	return s.repository.Archive(ctx, id)
}
