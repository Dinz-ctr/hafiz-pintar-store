package handler

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"

	"hafiz-pintar-store/backend/api/internal/repository"
	"hafiz-pintar-store/backend/api/internal/service"
)

type AdminOrderHandler struct {
	service *service.AdminOrderService
}

func NewAdminOrderHandler(
	svc *service.AdminOrderService,
) *AdminOrderHandler {
	return &AdminOrderHandler{service: svc}
}

type adminOrderStatusRequest struct {
	Status string `json:"status"`
}

type adminOrderShippingRequest struct {
	CourierName    string `json:"courier_name"`
	TrackingNumber string `json:"tracking_number"`
}

func writeAdminOrderJSON(
	w http.ResponseWriter,
	statusCode int,
	payload any,
) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(payload); err != nil {
		log.Printf("Admin order JSON response error: %v", err)
	}
}

func adminOrderError(
	w http.ResponseWriter,
	statusCode int,
	message string,
) {
	writeAdminOrderJSON(w, statusCode, map[string]any{
		"message": message,
	})
}

func (h *AdminOrderHandler) handleError(
	w http.ResponseWriter,
	err error,
) {
	switch {
	case errors.Is(err, service.ErrInvalidAdminOrderID):
		adminOrderError(w, http.StatusBadRequest, "invalid order ID")

	case errors.Is(err, service.ErrInvalidAdminOrderStatus):
		adminOrderError(w, http.StatusBadRequest, "invalid order status")

	case errors.Is(err, service.ErrInvalidAdminOrderFilter):
		adminOrderError(w, http.StatusBadRequest, "invalid order status filter")

	case errors.Is(err, service.ErrInvalidOrderTransition):
		adminOrderError(w, http.StatusConflict, "order status transition is not allowed")

	case errors.Is(err, service.ErrInvalidShippingDetails):
		adminOrderError(w, http.StatusBadRequest, "courier name and tracking number are required and must not exceed 100 characters")

	case errors.Is(err, repository.ErrAdminOrderNotFound):
		adminOrderError(w, http.StatusNotFound, "order not found")

	case errors.Is(err, repository.ErrAdminOrderConflict):
		adminOrderError(w, http.StatusConflict, "order status changed; refresh and try again")

	case errors.Is(err, repository.ErrAdminOrderStockRestore):
		adminOrderError(w, http.StatusConflict, "order cancellation failed because its stock could not be safely restored")

	default:
		log.Printf("Admin order operation failed: %v", err)
		adminOrderError(w, http.StatusInternalServerError, "failed to process admin order request")
	}
}

// GET /api/v1/admin/orders
// Query: page, limit, keyword, status
func (h *AdminOrderHandler) List(
	w http.ResponseWriter,
	r *http.Request,
) {
	query := r.URL.Query()

	page := int64(1)
	limit := int64(20)

	if value := query.Get("page"); value != "" {
		parsed, err := strconv.ParseInt(value, 10, 32)
		if err != nil || parsed <= 0 {
			adminOrderError(w, http.StatusBadRequest, "page must be a positive integer")
			return
		}
		page = parsed
	}

	if value := query.Get("limit"); value != "" {
		parsed, err := strconv.ParseInt(value, 10, 32)
		if err != nil || parsed <= 0 {
			adminOrderError(w, http.StatusBadRequest, "limit must be a positive integer")
			return
		}
		limit = parsed
	}

	if limit > 100 {
		limit = 100
	}

	result, err := h.service.List(
		r.Context(),
		int32(page),
		int32(limit),
		query.Get("keyword"),
		query.Get("status"),
	)
	if err != nil {
		h.handleError(w, err)
		return
	}

	writeAdminOrderJSON(w, http.StatusOK, map[string]any{
		"data": result.Data,
		"meta": map[string]any{
			"page":        result.Page,
			"limit":       result.Limit,
			"total":       result.Total,
			"total_pages": result.TotalPages,
			"keyword":     result.Keyword,
			"status":      result.Status,
		},
	})
}

// GET /api/v1/admin/orders/{id}
func (h *AdminOrderHandler) GetByID(
	w http.ResponseWriter,
	r *http.Request,
) {
	orderID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || orderID <= 0 {
		adminOrderError(w, http.StatusBadRequest, "valid order ID is required")
		return
	}

	order, err := h.service.GetByID(r.Context(), orderID)
	if err != nil {
		h.handleError(w, err)
		return
	}

	writeAdminOrderJSON(w, http.StatusOK, map[string]any{
		"data": order,
	})
}

// PATCH /api/v1/admin/orders/{id}/status
func (h *AdminOrderHandler) UpdateStatus(
	w http.ResponseWriter,
	r *http.Request,
) {
	orderID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || orderID <= 0 {
		adminOrderError(w, http.StatusBadRequest, "valid order ID is required")
		return
	}

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	var input adminOrderStatusRequest
	if err := decoder.Decode(&input); err != nil {
		adminOrderError(w, http.StatusBadRequest, "invalid JSON request body")
		return
	}

	var extra any
	if err := decoder.Decode(&extra); !errors.Is(err, io.EOF) {
		adminOrderError(w, http.StatusBadRequest, "request body must contain one JSON object")
		return
	}

	input.Status = strings.ToLower(strings.TrimSpace(input.Status))
	if input.Status == "" {
		adminOrderError(w, http.StatusBadRequest, "status is required")
		return
	}

	order, err := h.service.UpdateStatus(
		r.Context(),
		orderID,
		input.Status,
	)
	if err != nil {
		h.handleError(w, err)
		return
	}

	message := "order status updated successfully"
	if input.Status == "cancelled" {
		message = "order cancelled and stock restored successfully"
	}

	writeAdminOrderJSON(w, http.StatusOK, map[string]any{
		"message": message,
		"data":    order,
	})
}

// PATCH /api/v1/admin/orders/{id}/shipping
func (h *AdminOrderHandler) UpdateShipping(
	w http.ResponseWriter,
	r *http.Request,
) {
	orderID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || orderID <= 0 {
		adminOrderError(w, http.StatusBadRequest, "valid order ID is required")
		return
	}

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	var input adminOrderShippingRequest
	if err := decoder.Decode(&input); err != nil {
		adminOrderError(w, http.StatusBadRequest, "invalid JSON request body")
		return
	}

	var extra any
	if err := decoder.Decode(&extra); !errors.Is(err, io.EOF) {
		adminOrderError(w, http.StatusBadRequest, "request body must contain one JSON object")
		return
	}

	input.CourierName = strings.TrimSpace(input.CourierName)
	input.TrackingNumber = strings.TrimSpace(input.TrackingNumber)

	if input.CourierName == "" || input.TrackingNumber == "" {
		adminOrderError(w, http.StatusBadRequest, "courier_name and tracking_number are required")
		return
	}

	order, err := h.service.UpdateShipping(
		r.Context(),
		orderID,
		input.CourierName,
		input.TrackingNumber,
	)
	if err != nil {
		h.handleError(w, err)
		return
	}

	writeAdminOrderJSON(w, http.StatusOK, map[string]any{
		"message": "order shipping information saved successfully",
		"data":    order,
	})
}
