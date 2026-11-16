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

type AdminProductHandler struct {
	service        *service.AdminProductService
	productHandler *ProductHandler
}

func NewAdminProductHandler(
	svc *service.AdminProductService,
	productHandler *ProductHandler,
) *AdminProductHandler {
	return &AdminProductHandler{
		service:        svc,
		productHandler: productHandler,
	}
}

func writeAdminProductJSON(
	w http.ResponseWriter,
	status int,
	data any,
) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

func parseAdminProductID(
	r *http.Request,
	key string,
) (int64, error) {
	id, err := strconv.ParseInt(r.PathValue(key), 10, 64)

	if err != nil || id < 1 {
		return 0, service.ErrInvalidAdminProductID
	}

	return id, nil
}

func writeAdminProductError(
	w http.ResponseWriter,
	err error,
	operation string,
) {
	switch {
	case errors.Is(err, service.ErrInvalidAdminProductID),
		errors.Is(err, service.ErrInvalidAdminProductData),
		errors.Is(err, service.ErrInvalidAdminProductStatus),
		errors.Is(err, service.ErrInvalidAdminVariant):

		writeAdminProductJSON(
			w,
			http.StatusBadRequest,
			map[string]string{"error": err.Error()},
		)

	case errors.Is(err, repository.ErrAdminProductNotFound):

		writeAdminProductJSON(
			w,
			http.StatusNotFound,
			map[string]string{"error": "product not found"},
		)

	case errors.Is(err, repository.ErrProductVariantNotFound):

		writeAdminProductJSON(
			w,
			http.StatusConflict,
			map[string]string{"error": err.Error()},
		)

	default:
		var pgErr *pgconn.PgError

		if errors.As(err, &pgErr) {
			switch pgErr.Code {
			case "23505":
				writeAdminProductJSON(
					w,
					http.StatusConflict,
					map[string]string{
						"error": "product slug or SKU already exists",
					},
				)
				return

			case "23503":
				writeAdminProductJSON(
					w,
					http.StatusBadRequest,
					map[string]string{
						"error": "referenced category does not exist",
					},
				)
				return

			case "23514", "23502":
				writeAdminProductJSON(
					w,
					http.StatusBadRequest,
					map[string]string{
						"error": "product data violates database constraints",
					},
				)
				return
			}
		}

		writeAdminProductJSON(
			w,
			http.StatusInternalServerError,
			map[string]string{
				"error": "failed to " + operation,
			},
		)
	}
}

// GET /api/v1/admin/products
func (h *AdminProductHandler) List(
	w http.ResponseWriter,
	r *http.Request,
) {
	q := r.URL.Query()

	limit := int64(20)
	offset := int64(0)

	if raw := q.Get("limit"); raw != "" {
		value, err := strconv.ParseInt(raw, 10, 32)

		if err != nil || value < 1 || value > 100 {
			writeAdminProductJSON(
				w,
				http.StatusBadRequest,
				map[string]string{
					"error": "limit must be between 1 and 100",
				},
			)
			return
		}

		limit = value
	}

	if raw := q.Get("offset"); raw != "" {
		value, err := strconv.ParseInt(raw, 10, 32)

		if err != nil || value < 0 {
			writeAdminProductJSON(
				w,
				http.StatusBadRequest,
				map[string]string{
					"error": "offset must be zero or greater",
				},
			)
			return
		}

		offset = value
	}

	products, err := h.service.List(
		r.Context(),
		int32(limit),
		int32(offset),
		q.Get("search"),
		q.Get("status"),
	)

	if err != nil {
		writeAdminProductError(w, err, "retrieve products")
		return
	}

	writeAdminProductJSON(w, http.StatusOK, map[string]any{
		"data": products,
		"meta": map[string]int64{
			"limit":    limit,
			"offset":   offset,
			"returned": int64(len(products)),
		},
	})
}

// GET /api/v1/admin/products/{id}
func (h *AdminProductHandler) GetByID(
	w http.ResponseWriter,
	r *http.Request,
) {
	id, err := parseAdminProductID(r, "id")
	if err != nil {
		writeAdminProductError(w, err, "retrieve product")
		return
	}

	product, err := h.service.GetByID(r.Context(), id)
	if err != nil {
		writeAdminProductError(w, err, "retrieve product")
		return
	}

	writeAdminProductJSON(w, http.StatusOK, map[string]any{
		"data": product,
	})
}

// POST /api/v1/admin/products
func (h *AdminProductHandler) Create(
	w http.ResponseWriter,
	r *http.Request,
) {
	h.productHandler.Create(w, r)
}

type UpdateAdminProductRequest struct {
	CategoryID  *int64  `json:"category_id"`
	Name        string  `json:"name"`
	Slug        string  `json:"slug"`
	Description *string `json:"description"`
}

// PUT /api/v1/admin/products/{id}
func (h *AdminProductHandler) UpdateDetails(
	w http.ResponseWriter,
	r *http.Request,
) {
	id, err := parseAdminProductID(r, "id")
	if err != nil {
		writeAdminProductError(w, err, "update product")
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	defer r.Body.Close()

	var request UpdateAdminProductRequest

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&request); err != nil {
		writeAdminProductJSON(
			w,
			http.StatusBadRequest,
			map[string]string{"error": "invalid JSON request"},
		)
		return
	}

	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		writeAdminProductJSON(
			w,
			http.StatusBadRequest,
			map[string]string{
				"error": "request must contain exactly one JSON object",
			},
		)
		return
	}

	err = h.service.UpdateDetails(
		r.Context(),
		id,
		request.CategoryID,
		request.Name,
		request.Slug,
		request.Description,
	)

	if err != nil {
		writeAdminProductError(w, err, "update product")
		return
	}

	writeAdminProductJSON(
		w,
		http.StatusOK,
		map[string]string{
			"message": "product updated successfully",
		},
	)
}

type UpdateAdminProductVariantRequest struct {
	Name       string `json:"name"`
	Price      int64  `json:"price"`
	StockQty   int32  `json:"stock_qty"`
	WeightGram int32  `json:"weight_gram"`
}

// PUT /api/v1/admin/products/{id}/variants/{variantId}
func (h *AdminProductHandler) UpdateVariant(
	w http.ResponseWriter,
	r *http.Request,
) {
	productID, err := parseAdminProductID(r, "id")
	if err != nil {
		writeAdminProductError(w, err, "update product variant")
		return
	}

	variantID, err := parseAdminProductID(r, "variantId")
	if err != nil {
		writeAdminProductError(w, err, "update product variant")
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	defer r.Body.Close()

	var request UpdateAdminProductVariantRequest

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&request); err != nil {
		writeAdminProductJSON(
			w,
			http.StatusBadRequest,
			map[string]string{"error": "invalid JSON request"},
		)
		return
	}

	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		writeAdminProductJSON(
			w,
			http.StatusBadRequest,
			map[string]string{
				"error": "request must contain exactly one JSON object",
			},
		)
		return
	}

	err = h.service.UpdateVariant(
		r.Context(),
		productID,
		variantID,
		request.Name,
		request.Price,
		request.StockQty,
		request.WeightGram,
	)

	if err != nil {
		writeAdminProductError(w, err, "update product variant")
		return
	}

	writeAdminProductJSON(
		w,
		http.StatusOK,
		map[string]string{
			"message": "product variant updated successfully",
		},
	)
}

// DELETE /api/v1/admin/products/{id}/variants/{variantId}
func (h *AdminProductHandler) DeleteVariant(
	w http.ResponseWriter,
	r *http.Request,
) {
	productID, err := parseAdminProductID(r, "id")
	if err != nil {
		writeAdminProductError(w, err, "delete product variant")
		return
	}

	variantID, err := parseAdminProductID(r, "variantId")
	if err != nil {
		writeAdminProductError(w, err, "delete product variant")
		return
	}

	if err := h.service.DeleteVariant(
		r.Context(),
		productID,
		variantID,
	); err != nil {
		writeAdminProductError(w, err, "delete product variant")
		return
	}

	writeAdminProductJSON(
		w,
		http.StatusOK,
		map[string]string{
			"message": "product variant deleted successfully",
		},
	)
}

type UpdateAdminProductStatusRequest struct {
	Status string `json:"status"`
}

// PATCH /api/v1/admin/products/{id}/status
func (h *AdminProductHandler) UpdateStatus(
	w http.ResponseWriter,
	r *http.Request,
) {
	id, err := parseAdminProductID(r, "id")
	if err != nil {
		writeAdminProductError(w, err, "update product status")
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	defer r.Body.Close()

	var request UpdateAdminProductStatusRequest

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&request); err != nil {
		writeAdminProductJSON(
			w,
			http.StatusBadRequest,
			map[string]string{"error": "invalid JSON request"},
		)
		return
	}

	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		writeAdminProductJSON(
			w,
			http.StatusBadRequest,
			map[string]string{
				"error": "request must contain exactly one JSON object",
			},
		)
		return
	}

	if err := h.service.UpdateStatus(
		r.Context(),
		id,
		request.Status,
	); err != nil {
		writeAdminProductError(w, err, "update product status")
		return
	}

	writeAdminProductJSON(
		w,
		http.StatusOK,
		map[string]string{
			"message": "product status updated successfully",
		},
	)
}

// DELETE /api/v1/admin/products/{id}
func (h *AdminProductHandler) Archive(
	w http.ResponseWriter,
	r *http.Request,
) {
	id, err := parseAdminProductID(r, "id")
	if err != nil {
		writeAdminProductError(w, err, "archive product")
		return
	}

	if err := h.service.Archive(r.Context(), id); err != nil {
		writeAdminProductError(w, err, "archive product")
		return
	}

	writeAdminProductJSON(
		w,
		http.StatusOK,
		map[string]string{
			"message": "product archived successfully",
		},
	)
}
