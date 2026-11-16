package handler

import (
	"errors"
	"log"
	"net/http"
	"strconv"

	"hafiz-pintar-store/backend/api/internal/repository"
	"hafiz-pintar-store/backend/api/internal/service"
)

type OrderHandler struct {
	service *service.OrderService
}

func NewOrderHandler(svc *service.OrderService) *OrderHandler {
	return &OrderHandler{
		service: svc,
	}
}

// =========================================
// REQUEST DTO
// =========================================

type CheckoutRequest struct {
	AddressID    int64  `json:"address_id"`
	CustomerNote string `json:"customer_note"`
}

// =========================================
// ERROR HANDLER
// =========================================

func orderError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, service.ErrInvalidCustomer):
		customerError(
			w,
			http.StatusUnauthorized,
			"invalid customer",
		)

	case errors.Is(err, service.ErrInvalidOrderID):
		customerError(
			w,
			http.StatusBadRequest,
			"invalid order ID",
		)

	case errors.Is(err, service.ErrOrderNotFound):
		customerError(
			w,
			http.StatusNotFound,
			"order not found",
		)

	case errors.Is(err, service.ErrInvalidPagination):
		customerError(
			w,
			http.StatusBadRequest,
			"invalid pagination",
		)

	case errors.Is(err, repository.ErrCheckoutCartNotFound):
		customerError(
			w,
			http.StatusNotFound,
			"shopping cart not found",
		)

	case errors.Is(err, repository.ErrCheckoutCartEmpty):
		customerError(
			w,
			http.StatusBadRequest,
			"shopping cart is empty",
		)

	case errors.Is(err, repository.ErrCheckoutAddress):
		customerError(
			w,
			http.StatusNotFound,
			"shipping address not found",
		)

	case errors.Is(err, repository.ErrCheckoutStock):
		customerError(
			w,
			http.StatusConflict,
			"insufficient product stock",
		)

	case errors.Is(err, repository.ErrCheckoutUnavailable):
		customerError(
			w,
			http.StatusConflict,
			"product or variant is unavailable",
		)

	default:
		log.Printf("Order operation failed: %v", err)

		customerError(
			w,
			http.StatusInternalServerError,
			"failed to process order request",
		)
	}
}

// =========================================
// CHECKOUT
// =========================================

// POST /api/v1/orders/checkout
func (h *OrderHandler) Checkout(
	w http.ResponseWriter,
	r *http.Request,
) {
	if r.Method != http.MethodPost {
		customerError(
			w,
			http.StatusMethodNotAllowed,
			"method not allowed",
		)
		return
	}

	customerID, ok := getAuthenticatedCustomerID(w, r)
	if !ok {
		return
	}

	var input CheckoutRequest

	if !decodeCartRequest(w, r, &input) {
		return
	}

	if input.AddressID <= 0 {
		customerError(
			w,
			http.StatusBadRequest,
			"valid address_id is required",
		)
		return
	}

	order, err := h.service.Checkout(
		r.Context(),
		customerID,
		input.AddressID,
		input.CustomerNote,
	)

	if err != nil {
		orderError(w, err)
		return
	}

	writeCustomerJSON(w, http.StatusCreated, map[string]any{
		"message": "order created successfully",
		"data":    order,
	})
}

// =========================================
// GET CUSTOMER ORDER HISTORY
// =========================================

// GET /api/v1/orders
//
// Query parameters:
// ?page=1&limit=10
func (h *OrderHandler) GetCustomerOrders(
	w http.ResponseWriter,
	r *http.Request,
) {
	if r.Method != http.MethodGet {
		customerError(
			w,
			http.StatusMethodNotAllowed,
			"method not allowed",
		)
		return
	}

	customerID, ok := getAuthenticatedCustomerID(w, r)
	if !ok {
		return
	}

	// Default pagination.
	page := int64(1)
	limit := int64(10)

	query := r.URL.Query()

	// Ambil page dari query parameter.
	if value := query.Get("page"); value != "" {
		parsed, err := strconv.ParseInt(value, 10, 32)

		if err != nil || parsed <= 0 {
			customerError(
				w,
				http.StatusBadRequest,
				"page must be a positive integer",
			)
			return
		}

		page = parsed
	}

	// Ambil limit dari query parameter.
	if value := query.Get("limit"); value != "" {
		parsed, err := strconv.ParseInt(value, 10, 32)

		if err != nil || parsed <= 0 {
			customerError(
				w,
				http.StatusBadRequest,
				"limit must be a positive integer",
			)
			return
		}

		limit = parsed
	}

	if limit > 100 {
		limit = 100
	}

	orders, err := h.service.GetCustomerOrders(
		r.Context(),
		customerID,
		int32(page),
		int32(limit),
	)

	if err != nil {
		orderError(w, err)
		return
	}

	writeCustomerJSON(w, http.StatusOK, map[string]any{
		"message": "order history retrieved successfully",
		"data":    orders,
		"pagination": map[string]any{
			"page":  page,
			"limit": limit,
		},
	})
}

// =========================================
// GET CUSTOMER ORDER DETAIL
// =========================================

// GET /api/v1/orders/{id}
func (h *OrderHandler) GetCustomerOrderDetail(
	w http.ResponseWriter,
	r *http.Request,
) {
	if r.Method != http.MethodGet {
		customerError(
			w,
			http.StatusMethodNotAllowed,
			"method not allowed",
		)
		return
	}

	customerID, ok := getAuthenticatedCustomerID(w, r)
	if !ok {
		return
	}

	// Ambil ID order dari URL.
	orderIDValue := r.PathValue("id")

	orderID, err := strconv.ParseInt(orderIDValue, 10, 64)

	if err != nil || orderID <= 0 {
		customerError(
			w,
			http.StatusBadRequest,
			"valid order ID is required",
		)
		return
	}

	order, err := h.service.GetCustomerOrderDetail(
		r.Context(),
		customerID,
		orderID,
	)

	if err != nil {
		orderError(w, err)
		return
	}

	writeCustomerJSON(w, http.StatusOK, map[string]any{
		"message": "order detail retrieved successfully",
		"data":    order,
	})
}
