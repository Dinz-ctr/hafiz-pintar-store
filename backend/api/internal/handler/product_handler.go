package handler

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"
	"strings"

	"hafiz-pintar-store/backend/api/internal/repository"
	"hafiz-pintar-store/backend/api/internal/service"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type ProductHandler struct {
	service *service.ProductService
}

func NewProductHandler(svc *service.ProductService) *ProductHandler {
	return &ProductHandler{
		service: svc,
	}
}

// =========================================
// REQUEST STRUCTS
// =========================================

type CreateProductRequest struct {
	CategoryID  *int64                        `json:"category_id"`
	Name        string                        `json:"name"`
	Slug        string                        `json:"slug"`
	Description *string                       `json:"description"`
	Variants    []CreateProductVariantRequest `json:"variants"`
}

type CreateProductVariantRequest struct {
	Sku        string            `json:"sku"`
	Name       string            `json:"name"`
	Price      int64             `json:"price"`
	StockQty   int32             `json:"stock_qty"`
	WeightGram int32             `json:"weight_gram"`
	IsDefault  bool              `json:"is_default"`
	Attributes map[string]string `json:"attributes"`
}

// =========================================
// JSON RESPONSE
// =========================================

func writeProductJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	_ = json.NewEncoder(w).Encode(data)
}

// =========================================
// GET /api/v1/products
// LIST ACTIVE PRODUCTS
// =========================================

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

	products, err := h.service.List(
		r.Context(),
		int32(limit),
		int32(offset),
	)

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

// =========================================
// GET /api/v1/products/{slug}
// GET PRODUCT DETAIL
// =========================================

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

// =========================================
// GET /api/v1/products/{slug}/variants
// GET PRODUCT VARIANTS WITH ATTRIBUTES
// =========================================

func (h *ProductHandler) GetVariantsBySlug(
	w http.ResponseWriter,
	r *http.Request,
) {
	slug := r.PathValue("slug")

	variants, err := h.service.GetVariantsBySlug(
		r.Context(),
		slug,
	)

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
			"error": "failed to retrieve product variants",
		})
		return
	}

	writeProductJSON(w, http.StatusOK, map[string]any{
		"data": variants,
		"meta": map[string]int{
			"total": len(variants),
		},
	})
}

// =========================================
// POST /api/v1/products
// CREATE PRODUCT WITH VARIANTS
// =========================================

func (h *ProductHandler) Create(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

	var request CreateProductRequest

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&request); err != nil {
		writeProductJSON(w, http.StatusBadRequest, map[string]string{
			"error": "invalid JSON request: " + err.Error(),
		})
		return
	}

	// Pastikan hanya ada satu JSON object.
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		writeProductJSON(w, http.StatusBadRequest, map[string]string{
			"error": "request must contain exactly one JSON object",
		})
		return
	}

	variants := make(
		[]repository.CreateProductVariantInput,
		0,
		len(request.Variants),
	)

	for _, variant := range request.Variants {
		variants = append(variants, repository.CreateProductVariantInput{
			Sku:        strings.TrimSpace(variant.Sku),
			Name:       strings.TrimSpace(variant.Name),
			Price:      variant.Price,
			StockQty:   variant.StockQty,
			WeightGram: variant.WeightGram,
			IsDefault:  variant.IsDefault,
			Attributes: variant.Attributes,
		})
	}

	input := repository.CreateProductInput{
		CategoryID:  request.CategoryID,
		Name:        request.Name,
		Slug:        request.Slug,
		Description: request.Description,
		Variants:    variants,
	}

	product, err := h.service.Create(r.Context(), input)

	if err != nil {
		// Validasi dari service dan repository.
		switch {
		case errors.Is(err, service.ErrInvalidProductName),
			errors.Is(err, service.ErrInvalidProductSlug),
			errors.Is(err, service.ErrInvalidVariants),
			errors.Is(err, service.ErrInvalidVariant),
			errors.Is(err, service.ErrInvalidAttribute),
			errors.Is(err, service.ErrDuplicateVariantSKU),
			errors.Is(err, repository.ErrInvalidDefaultVariant):

			writeProductJSON(w, http.StatusBadRequest, map[string]string{
				"error": err.Error(),
			})
			return
		}

		// Constraint PostgreSQL.
		var pgErr *pgconn.PgError

		if errors.As(err, &pgErr) {
			switch pgErr.Code {
			case "23505":
				writeProductJSON(w, http.StatusConflict, map[string]string{
					"error": "SKU or product slug already exists",
				})
				return

			case "23503":
				writeProductJSON(w, http.StatusBadRequest, map[string]string{
					"error": "referenced category does not exist",
				})
				return

			case "23514", "23502":
				writeProductJSON(w, http.StatusBadRequest, map[string]string{
					"error": "product data violates database constraints",
				})
				return
			}
		}

		writeProductJSON(w, http.StatusInternalServerError, map[string]string{
			"error": "failed to create product",
		})
		return
	}

	writeProductJSON(w, http.StatusCreated, map[string]any{
		"message": "product created successfully",
		"data":    product,
	})
}
