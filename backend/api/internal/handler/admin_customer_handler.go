package handler

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"
	"strings"

	"hafiz-pintar-store/backend/api/internal/repository"
	"hafiz-pintar-store/backend/api/internal/service"
)

type AdminCustomerHandler struct {
	service *service.AdminCustomerService
}

func NewAdminCustomerHandler(
	svc *service.AdminCustomerService,
) *AdminCustomerHandler {
	return &AdminCustomerHandler{service: svc}
}

func adminCustomerJSON(
	w http.ResponseWriter,
	status int,
	data any,
) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Printf("Admin customer JSON response error: %v", err)
	}
}

func adminCustomerError(
	w http.ResponseWriter,
	status int,
	message string,
) {
	adminCustomerJSON(w, status, map[string]string{
		"error": message,
	})
}

func adminCustomerID(r *http.Request) (int64, error) {
	return strconv.ParseInt(r.PathValue("id"), 10, 64)
}

func adminCustomerHandleError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, repository.ErrAdminCustomerNotFound):
		adminCustomerError(w, http.StatusNotFound, "customer not found")
	case errors.Is(err, service.ErrInvalidAdminCustomerID):
		adminCustomerError(w, http.StatusBadRequest, "invalid customer ID")
	default:
		log.Printf("Admin customer operation failed: %v", err)
		adminCustomerError(
			w,
			http.StatusInternalServerError,
			"failed to process customer request",
		)
	}
}

func parseAdminCustomerPagination(r *http.Request) (int32, int32) {
	page := int32(1)
	limit := int32(20)

	if value := r.URL.Query().Get("page"); value != "" {
		if parsed, err := strconv.ParseInt(value, 10, 32); err == nil {
			page = int32(parsed)
		}
	}

	if value := r.URL.Query().Get("limit"); value != "" {
		if parsed, err := strconv.ParseInt(value, 10, 32); err == nil {
			limit = int32(parsed)
		}
	}

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	return page, limit
}

func (h *AdminCustomerHandler) List(
	w http.ResponseWriter,
	r *http.Request,
) {
	page, limit := parseAdminCustomerPagination(r)
	keyword := strings.TrimSpace(r.URL.Query().Get("q"))

	customers, total, page, limit, err := h.service.List(
		r.Context(),
		keyword,
		page,
		limit,
	)
	if err != nil {
		adminCustomerHandleError(w, err)
		return
	}

	totalPages := int64(0)
	if limit > 0 {
		totalPages = (total + int64(limit) - 1) / int64(limit)
	}

	adminCustomerJSON(w, http.StatusOK, map[string]any{
		"data": customers,
		"meta": map[string]any{
			"page":        page,
			"limit":       limit,
			"total":       total,
			"total_pages": totalPages,
			"keyword":     keyword,
		},
	})
}

func (h *AdminCustomerHandler) GetByID(
	w http.ResponseWriter,
	r *http.Request,
) {
	id, err := adminCustomerID(r)
	if err != nil || id < 1 {
		adminCustomerError(w, http.StatusBadRequest, "invalid customer ID")
		return
	}

	customer, err := h.service.GetByID(r.Context(), id)
	if err != nil {
		adminCustomerHandleError(w, err)
		return
	}

	adminCustomerJSON(w, http.StatusOK, map[string]any{
		"data": customer,
	})
}

type UpdateAdminCustomerStatusRequest struct {
	IsActive *bool `json:"is_active"`
}

func (h *AdminCustomerHandler) UpdateStatus(
	w http.ResponseWriter,
	r *http.Request,
) {
	id, err := adminCustomerID(r)
	if err != nil || id < 1 {
		adminCustomerError(w, http.StatusBadRequest, "invalid customer ID")
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	defer r.Body.Close()

	var input UpdateAdminCustomerStatusRequest

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&input); err != nil {
		adminCustomerError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if input.IsActive == nil {
		adminCustomerError(w, http.StatusBadRequest, "is_active is required")
		return
	}

	var extra any
	if err := decoder.Decode(&extra); err == nil {
		adminCustomerError(
			w,
			http.StatusBadRequest,
			"request body must contain only one JSON object",
		)
		return
	}

	customer, err := h.service.UpdateStatus(
		r.Context(),
		id,
		*input.IsActive,
	)
	if err != nil {
		adminCustomerHandleError(w, err)
		return
	}

	adminCustomerJSON(w, http.StatusOK, map[string]any{
		"message": "customer status updated successfully",
		"data":    customer,
	})
}

func (h *AdminCustomerHandler) ListAddresses(
	w http.ResponseWriter,
	r *http.Request,
) {
	id, err := adminCustomerID(r)
	if err != nil || id < 1 {
		adminCustomerError(w, http.StatusBadRequest, "invalid customer ID")
		return
	}

	addresses, err := h.service.ListAddresses(r.Context(), id)
	if err != nil {
		adminCustomerHandleError(w, err)
		return
	}

	adminCustomerJSON(w, http.StatusOK, map[string]any{
		"data": addresses,
	})
}

func (h *AdminCustomerHandler) ListOrders(
	w http.ResponseWriter,
	r *http.Request,
) {
	id, err := adminCustomerID(r)
	if err != nil || id < 1 {
		adminCustomerError(w, http.StatusBadRequest, "invalid customer ID")
		return
	}

	page, limit := parseAdminCustomerPagination(r)

	orders, total, page, limit, err := h.service.ListOrders(
		r.Context(),
		id,
		page,
		limit,
	)
	if err != nil {
		adminCustomerHandleError(w, err)
		return
	}

	totalPages := int64(0)
	if limit > 0 {
		totalPages = (total + int64(limit) - 1) / int64(limit)
	}

	adminCustomerJSON(w, http.StatusOK, map[string]any{
		"data": orders,
		"meta": map[string]any{
			"page":        page,
			"limit":       limit,
			"total":       total,
			"total_pages": totalPages,
		},
	})
}
