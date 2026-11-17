package handler

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"

	"hafiz-pintar-store/backend/api/internal/repository"
	"hafiz-pintar-store/backend/api/internal/service"
)

type AdminPaymentHandler struct {
	service *service.AdminPaymentService
}

func NewAdminPaymentHandler(
	svc *service.AdminPaymentService,
) *AdminPaymentHandler {
	return &AdminPaymentHandler{
		service: svc,
	}
}

func adminPaymentJSON(
	w http.ResponseWriter,
	code int,
	payload any,
) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)

	if err := json.NewEncoder(w).Encode(payload); err != nil {
		log.Printf("Admin payment JSON response error: %v", err)
	}
}

// GET /api/v1/admin/payments
func (h *AdminPaymentHandler) List(
	w http.ResponseWriter,
	r *http.Request,
) {
	page64, _ := strconv.ParseInt(
		r.URL.Query().Get("page"),
		10,
		32,
	)

	limit64, _ := strconv.ParseInt(
		r.URL.Query().Get("limit"),
		10,
		32,
	)

	items, total, err := h.service.List(
		r.Context(),
		r.URL.Query().Get("status"),
		int32(page64),
		int32(limit64),
	)

	if err != nil {
		if errors.Is(err, service.ErrInvalidAdminPaymentStatus) {
			adminPaymentJSON(
				w,
				http.StatusBadRequest,
				map[string]any{"message": err.Error()},
			)
			return
		}

		log.Printf("Admin list payments failed: %v", err)

		adminPaymentJSON(
			w,
			http.StatusInternalServerError,
			map[string]any{
				"message": "failed to get payments",
			},
		)
		return
	}

	page := int32(page64)
	limit := int32(limit64)

	if page < 1 {
		page = 1
	}

	if limit < 1 {
		limit = 10
	}

	if limit > 100 {
		limit = 100
	}

	totalPages := (total + int64(limit) - 1) / int64(limit)

	adminPaymentJSON(
		w,
		http.StatusOK,
		map[string]any{
			"message": "payments retrieved successfully",
			"data":    items,
			"meta": map[string]any{
				"page":        page,
				"limit":       limit,
				"total":       total,
				"total_pages": totalPages,
			},
		},
	)
}

// GET /api/v1/admin/payments/{id}
func (h *AdminPaymentHandler) GetByID(
	w http.ResponseWriter,
	r *http.Request,
) {
	id, err := strconv.ParseInt(
		r.PathValue("id"),
		10,
		64,
	)

	if err != nil || id <= 0 {
		adminPaymentJSON(
			w,
			http.StatusBadRequest,
			map[string]any{
				"message": "invalid payment ID",
			},
		)
		return
	}

	item, err := h.service.GetByID(
		r.Context(),
		id,
	)

	if err != nil {
		if errors.Is(err, repository.ErrAdminPaymentNotFound) {
			adminPaymentJSON(
				w,
				http.StatusNotFound,
				map[string]any{
					"message": "payment not found",
				},
			)
			return
		}

		log.Printf("Admin get payment failed: %v", err)

		adminPaymentJSON(
			w,
			http.StatusInternalServerError,
			map[string]any{
				"message": "failed to get payment",
			},
		)
		return
	}

	adminPaymentJSON(
		w,
		http.StatusOK,
		map[string]any{
			"data": item,
		},
	)
}

type UpdateAdminPaymentStatusRequest struct {
	Status string `json:"status"`
}

// PATCH /api/v1/admin/payments/{id}/status
func (h *AdminPaymentHandler) UpdateStatus(
	w http.ResponseWriter,
	r *http.Request,
) {
	id, err := strconv.ParseInt(
		r.PathValue("id"),
		10,
		64,
	)

	if err != nil || id <= 0 {
		adminPaymentJSON(
			w,
			http.StatusBadRequest,
			map[string]any{
				"message": "invalid payment ID",
			},
		)
		return
	}

	var input UpdateAdminPaymentStatusRequest

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		adminPaymentJSON(
			w,
			http.StatusBadRequest,
			map[string]any{
				"message": "invalid JSON body",
			},
		)
		return
	}

	item, err := h.service.UpdateStatus(
		r.Context(),
		id,
		input.Status,
	)

	switch {
	case errors.Is(err, service.ErrInvalidAdminPaymentStatus):
		adminPaymentJSON(
			w,
			http.StatusBadRequest,
			map[string]any{"message": err.Error()},
		)

	case errors.Is(err, service.ErrInvalidAdminPaymentTransition):
		adminPaymentJSON(
			w,
			http.StatusConflict,
			map[string]any{"message": err.Error()},
		)

	case errors.Is(err, repository.ErrAdminPaymentNotFound):
		adminPaymentJSON(
			w,
			http.StatusNotFound,
			map[string]any{
				"message": "payment not found",
			},
		)

	case err != nil:
		log.Printf("Admin update payment status failed: %v", err)

		adminPaymentJSON(
			w,
			http.StatusInternalServerError,
			map[string]any{
				"message": "failed to update payment status",
			},
		)

	default:
		adminPaymentJSON(
			w,
			http.StatusOK,
			map[string]any{
				"message": "payment status updated successfully",
				"data":    item,
			},
		)
	}
}
