package handler

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"

	"hafiz-pintar-store/backend/api/internal/repository"
	"hafiz-pintar-store/backend/api/internal/service"

	"github.com/jackc/pgx/v5/pgconn"
)

type AdminCategoryHandler struct {
	service *service.AdminCategoryService
}

func NewAdminCategoryHandler(
	svc *service.AdminCategoryService,
) *AdminCategoryHandler {
	return &AdminCategoryHandler{service: svc}
}

func writeAdminCategoryJSON(
	w http.ResponseWriter,
	status int,
	data any,
) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

func parseAdminCategoryID(r *http.Request) (int64, error) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id < 1 {
		return 0, service.ErrInvalidAdminCategoryID
	}
	return id, nil
}

func decodeAdminCategoryJSON(
	w http.ResponseWriter,
	r *http.Request,
	target any,
) error {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	defer r.Body.Close()

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(target); err != nil {
		return err
	}

	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		return errors.New("request must contain exactly one JSON object")
	}

	return nil
}

func writeAdminCategoryError(
	w http.ResponseWriter,
	err error,
	operation string,
) {
	switch {
	case errors.Is(err, service.ErrInvalidAdminCategoryID),
		errors.Is(err, service.ErrInvalidAdminCategoryData),
		errors.Is(err, service.ErrInvalidAdminCategoryStatus):

		writeAdminCategoryJSON(w, http.StatusBadRequest,
			map[string]string{"error": err.Error()})

	case errors.Is(err, repository.ErrAdminCategoryNotFound):

		writeAdminCategoryJSON(w, http.StatusNotFound,
			map[string]string{"error": "category not found"})

	case errors.Is(err, repository.ErrAdminCategoryParentNotFound):

		writeAdminCategoryJSON(w, http.StatusBadRequest,
			map[string]string{"error": "parent category does not exist"})

	case errors.Is(err, repository.ErrAdminCategoryCycle):

		writeAdminCategoryJSON(w, http.StatusBadRequest,
			map[string]string{"error": "category cannot be its own parent or descendant"})

	case errors.Is(err, repository.ErrAdminCategoryHasChildren):

		writeAdminCategoryJSON(w, http.StatusConflict,
			map[string]string{"error": "archive active child categories first"})

	default:
		var pgErr *pgconn.PgError

		if errors.As(err, &pgErr) {
			switch pgErr.Code {
			case "23505":
				writeAdminCategoryJSON(w, http.StatusConflict,
					map[string]string{"error": "category slug already exists"})
				return

			case "23503":
				writeAdminCategoryJSON(w, http.StatusBadRequest,
					map[string]string{"error": "referenced parent category does not exist"})
				return

			case "23502", "23514":
				writeAdminCategoryJSON(w, http.StatusBadRequest,
					map[string]string{"error": "category data violates database constraints"})
				return
			}
		}

		writeAdminCategoryJSON(w, http.StatusInternalServerError,
			map[string]string{"error": "failed to " + operation})
	}
}

// GET /api/v1/admin/categories
func (h *AdminCategoryHandler) List(
	w http.ResponseWriter,
	r *http.Request,
) {
	q := r.URL.Query()

	limit := int64(20)
	offset := int64(0)

	if raw := q.Get("limit"); raw != "" {
		value, err := strconv.ParseInt(raw, 10, 32)
		if err != nil || value < 1 || value > 100 {
			writeAdminCategoryJSON(w, http.StatusBadRequest,
				map[string]string{"error": "limit must be between 1 and 100"})
			return
		}
		limit = value
	}

	if raw := q.Get("offset"); raw != "" {
		value, err := strconv.ParseInt(raw, 10, 32)
		if err != nil || value < 0 {
			writeAdminCategoryJSON(w, http.StatusBadRequest,
				map[string]string{"error": "offset must be zero or greater"})
			return
		}
		offset = value
	}

	categories, err := h.service.List(
		r.Context(),
		int32(limit),
		int32(offset),
		q.Get("search"),
		q.Get("status"),
	)
	if err != nil {
		writeAdminCategoryError(w, err, "retrieve categories")
		return
	}

	writeAdminCategoryJSON(w, http.StatusOK, map[string]any{
		"data": categories,
		"meta": map[string]int64{
			"limit":    limit,
			"offset":   offset,
			"returned": int64(len(categories)),
		},
	})
}

// GET /api/v1/admin/categories/{id}
func (h *AdminCategoryHandler) GetByID(
	w http.ResponseWriter,
	r *http.Request,
) {
	id, err := parseAdminCategoryID(r)
	if err != nil {
		writeAdminCategoryError(w, err, "retrieve category")
		return
	}

	category, err := h.service.GetByID(r.Context(), id)
	if err != nil {
		writeAdminCategoryError(w, err, "retrieve category")
		return
	}

	writeAdminCategoryJSON(w, http.StatusOK, map[string]any{
		"data": category,
	})
}

type CreateAdminCategoryRequest struct {
	ParentID    *int64  `json:"parent_id"`
	Name        string  `json:"name"`
	Slug        string  `json:"slug"`
	Description *string `json:"description"`
	IsActive    *bool   `json:"is_active"`
}

// POST /api/v1/admin/categories
func (h *AdminCategoryHandler) Create(
	w http.ResponseWriter,
	r *http.Request,
) {
	var request CreateAdminCategoryRequest

	if err := decodeAdminCategoryJSON(w, r, &request); err != nil {
		writeAdminCategoryJSON(w, http.StatusBadRequest,
			map[string]string{"error": "invalid JSON request"})
		return
	}

	isActive := true
	if request.IsActive != nil {
		isActive = *request.IsActive
	}

	category, err := h.service.Create(
		r.Context(),
		request.ParentID,
		request.Name,
		request.Slug,
		request.Description,
		isActive,
	)
	if err != nil {
		writeAdminCategoryError(w, err, "create category")
		return
	}

	writeAdminCategoryJSON(w, http.StatusCreated, map[string]any{
		"message": "category created successfully",
		"data":    category,
	})
}

type UpdateAdminCategoryRequest struct {
	ParentID    *int64  `json:"parent_id"`
	Name        string  `json:"name"`
	Slug        string  `json:"slug"`
	Description *string `json:"description"`
}

// PUT /api/v1/admin/categories/{id}
func (h *AdminCategoryHandler) UpdateDetails(
	w http.ResponseWriter,
	r *http.Request,
) {
	id, err := parseAdminCategoryID(r)
	if err != nil {
		writeAdminCategoryError(w, err, "update category")
		return
	}

	var request UpdateAdminCategoryRequest

	if err := decodeAdminCategoryJSON(w, r, &request); err != nil {
		writeAdminCategoryJSON(w, http.StatusBadRequest,
			map[string]string{"error": "invalid JSON request"})
		return
	}

	err = h.service.UpdateDetails(
		r.Context(),
		id,
		request.ParentID,
		request.Name,
		request.Slug,
		request.Description,
	)
	if err != nil {
		writeAdminCategoryError(w, err, "update category")
		return
	}

	writeAdminCategoryJSON(w, http.StatusOK, map[string]string{
		"message": "category updated successfully",
	})
}

type UpdateAdminCategoryStatusRequest struct {
	IsActive *bool `json:"is_active"`
}

// PATCH /api/v1/admin/categories/{id}/status
func (h *AdminCategoryHandler) UpdateStatus(
	w http.ResponseWriter,
	r *http.Request,
) {
	id, err := parseAdminCategoryID(r)
	if err != nil {
		writeAdminCategoryError(w, err, "update category status")
		return
	}

	var request UpdateAdminCategoryStatusRequest

	if err := decodeAdminCategoryJSON(w, r, &request); err != nil {
		writeAdminCategoryJSON(w, http.StatusBadRequest,
			map[string]string{"error": "invalid JSON request"})
		return
	}

	if request.IsActive == nil {
		writeAdminCategoryJSON(w, http.StatusBadRequest,
			map[string]string{"error": "is_active is required"})
		return
	}

	if err := h.service.UpdateStatus(
		r.Context(),
		id,
		*request.IsActive,
	); err != nil {
		writeAdminCategoryError(w, err, "update category status")
		return
	}

	writeAdminCategoryJSON(w, http.StatusOK, map[string]string{
		"message": "category status updated successfully",
	})
}

// DELETE /api/v1/admin/categories/{id}
// Soft delete: mengubah is_active menjadi FALSE.
func (h *AdminCategoryHandler) Archive(
	w http.ResponseWriter,
	r *http.Request,
) {
	id, err := parseAdminCategoryID(r)
	if err != nil {
		writeAdminCategoryError(w, err, "archive category")
		return
	}

	if err := h.service.Archive(r.Context(), id); err != nil {
		writeAdminCategoryError(w, err, "archive category")
		return
	}

	writeAdminCategoryJSON(w, http.StatusOK, map[string]string{
		"message": "category archived successfully",
	})
}
