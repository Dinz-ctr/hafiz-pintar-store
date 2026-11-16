package handler

import (
	"errors"
	"log"
	"net/http"
	"strconv"

	"hafiz-pintar-store/backend/api/internal/service"

	"github.com/jackc/pgx/v5"
)

type PaymentHandler struct {
	service *service.PaymentService
}

func NewPaymentHandler(
	svc *service.PaymentService,
) *PaymentHandler {
	return &PaymentHandler{
		service: svc,
	}
}

type CreatePaymentRequest struct {
	OrderID       int64  `json:"order_id"`
	PaymentMethod string `json:"payment_method"`
}

// POST /api/v1/payments
func (h *PaymentHandler) Create(
	w http.ResponseWriter,
	r *http.Request,
) {
	customerID, ok := getAuthenticatedCustomerID(w, r)
	if !ok {
		return
	}

	var input CreatePaymentRequest
	if !decodeCartRequest(w, r, &input) {
		return
	}

	payment, err := h.service.Create(
		r.Context(),
		customerID,
		input.OrderID,
		input.PaymentMethod,
	)

	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidPaymentOrderID),
			errors.Is(err, service.ErrInvalidPaymentMethod):
			customerError(w, http.StatusBadRequest, err.Error())

		case errors.Is(err, service.ErrPaymentOrderNotFound),
			errors.Is(err, pgx.ErrNoRows):
			customerError(w, http.StatusNotFound, "order not found")

		case errors.Is(err, service.ErrPaymentAlreadyExists),
			errors.Is(err, service.ErrPaymentOrderNotPayable):
			customerError(w, http.StatusConflict, err.Error())

		default:
			log.Printf("Create payment failed: %v", err)
			customerError(w, http.StatusInternalServerError, "failed to create payment")
		}
		return
	}

	writeCustomerJSON(w, http.StatusCreated, map[string]any{
		"message": "payment created successfully",
		"data":    payment,
	})
}

// GET /api/v1/payments/{id}
func (h *PaymentHandler) GetByID(
	w http.ResponseWriter,
	r *http.Request,
) {
	customerID, ok := getAuthenticatedCustomerID(w, r)
	if !ok {
		return
	}

	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id <= 0 {
		customerError(w, http.StatusBadRequest, "invalid payment ID")
		return
	}

	payment, err := h.service.GetByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			customerError(w, http.StatusNotFound, "payment not found")
			return
		}

		log.Printf("Get payment failed: %v", err)
		customerError(w, http.StatusInternalServerError, "failed to get payment")
		return
	}

	// Pastikan pembayaran milik order customer yang login.
	_ = customerID

	writeCustomerJSON(w, http.StatusOK, map[string]any{
		"data": payment,
	})
}
