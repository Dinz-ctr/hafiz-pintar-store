package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"hafiz-pintar-store/backend/api/internal/service"

	"github.com/jackc/pgx/v5"
)

type CategoryHandler struct {
	service *service.CategoryService
}

func NewCategoryHandler(svc *service.CategoryService) *CategoryHandler {
	return &CategoryHandler{
		service: svc,
	}
}

func writeCategoryJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(data); err != nil {
		return
	}
}

// List mengambil seluruh kategori aktif.
func (h *CategoryHandler) List(w http.ResponseWriter, r *http.Request) {
	categories, err := h.service.ListActive(r.Context())
	if err != nil {
		writeCategoryJSON(w, http.StatusInternalServerError, map[string]string{
			"error": "failed to retrieve categories",
		})
		return
	}

	writeCategoryJSON(w, http.StatusOK, map[string]any{
		"data": categories,
		"meta": map[string]int{
			"total": len(categories),
		},
	})
}

// GetBySlug mengambil kategori aktif berdasarkan slug.
func (h *CategoryHandler) GetBySlug(w http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("slug")

	category, err := h.service.GetActiveBySlug(r.Context(), slug)
	if err != nil {
		if errors.Is(err, service.ErrInvalidCategorySlug) {
			writeCategoryJSON(w, http.StatusBadRequest, map[string]string{
				"error": err.Error(),
			})
			return
		}

		if errors.Is(err, pgx.ErrNoRows) {
			writeCategoryJSON(w, http.StatusNotFound, map[string]string{
				"error": "category not found",
			})
			return
		}

		writeCategoryJSON(w, http.StatusInternalServerError, map[string]string{
			"error": "failed to retrieve category",
		})
		return
	}

	writeCategoryJSON(w, http.StatusOK, map[string]any{
		"data": category,
	})
}
