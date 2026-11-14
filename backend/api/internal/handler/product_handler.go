package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"hafiz-pintar-store/backend/api/internal/service"

	"github.com/jackc/pgx/v5"
)

type ProductHandler struct {
	service *service.ProductService
}

func NewProductHandler(svc *service.ProductService) *ProductHandler {
	return &ProductHandler{
		service: svc,
	}
}

func writeProductJSON(
	w http.ResponseWriter,
	status int,
	data any,
) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(data); err != nil {
		return
	}
}

func (h *ProductHandler) List(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()

	limit := int64(20)
	offset := int64(0)

	if value := query.Get("limit"); value != "" {
		parsed, err := strconv.ParseInt(value, 10, 32)
		if err != nil || parsed < 1 || parsed > 100 {
			writeProductJSON(w, http.StatusBadRequest, map[string]string{
				"error": "limit must be between 1 and 100",
			})
			return
		}
		limit = parsed
	}

	if value := query.Get("offset"); value != "" {
		parsed, err := strconv.ParseInt(value, 10, 32)
		if err != nil || parsed < 0 {
			writeProductJSON(w, http.StatusBadRequest, map[string]string{
				"error": "offset must be zero or greater",
			})
			return
		}
		offset = parsed
	}

	products, err := h.service.List(r.Context(), int32(limit), int32(offset))
	if err != nil {
		writeProductJSON(w, http.StatusInternalServerError, map[string]string{
			"error": "failed to retrieve products",
		})
		return
	}

	writeProductJSON(w, http.StatusOK, map[string]any{
		"data": products,
		"meta": map[string]int32{
			"limit":  int32(limit),
			"offset": int32(offset),
		},
	})
}

func (h *ProductHandler) GetBySlug(w http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("slug")

	product, err := h.service.GetBySlug(r.Context(), slug)
	if err != nil {
		if errors.Is(err, service.ErrInvalidProductSlug) {
			writeProductJSON(w, http.StatusBadRequest, map[string]string{
				"error": err.Error(),
			})
			return
		}

		if errors.Is(err, pgx.ErrNoRows) {
			writeProductJSON(w, http.StatusNotFound, map[string]string{
				"error": "product not found",
			})
			return
		}

		writeProductJSON(w, http.StatusInternalServerError, map[string]string{
			"error": "failed to retrieve product",
		})
		return
	}

	writeProductJSON(w, http.StatusOK, map[string]any{
		"data": product,
	})
}
