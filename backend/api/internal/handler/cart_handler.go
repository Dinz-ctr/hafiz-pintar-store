package handler

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"strconv"

	"hafiz-pintar-store/backend/api/internal/service"

	"github.com/jackc/pgx/v5"
)

type CartHandler struct {
	service *service.CartService
}

func NewCartHandler(svc *service.CartService) *CartHandler {
	return &CartHandler{
		service: svc,
	}
}

// =========================================
// REQUEST DTO
// =========================================

type AddCartItemRequest struct {
	ProductID int64 `json:"product_id"`
	VariantID int64 `json:"variant_id"`
	Quantity  int32 `json:"quantity"`
}

type UpdateCartItemRequest struct {
	Quantity int32 `json:"quantity"`
}

// =========================================
// HELPERS
// =========================================

func parseCartItemID(
	w http.ResponseWriter,
	r *http.Request,
) (int64, bool) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)

	if err != nil || id <= 0 {
		customerError(
			w,
			http.StatusBadRequest,
			"invalid cart item ID",
		)
		return 0, false
	}

	return id, true
}

func decodeCartRequest(
	w http.ResponseWriter,
	r *http.Request,
	target any,
) bool {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	defer r.Body.Close()

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(target); err != nil {
		customerError(
			w,
			http.StatusBadRequest,
			"invalid request body",
		)
		return false
	}

	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		customerError(
			w,
			http.StatusBadRequest,
			"request body must contain only one JSON object",
		)
		return false
	}

	return true
}

func cartError(
	w http.ResponseWriter,
	err error,
) {
	switch {
	case errors.Is(err, pgx.ErrNoRows),
		errors.Is(err, service.ErrCartItemNotFound):

		customerError(
			w,
			http.StatusNotFound,
			"cart item not found",
		)

	case errors.Is(err, service.ErrInvalidCartCustomerID),
		errors.Is(err, service.ErrInvalidCartProductID),
		errors.Is(err, service.ErrInvalidCartVariantID),
		errors.Is(err, service.ErrInvalidCartItemID),
		errors.Is(err, service.ErrInvalidCartQuantity):

		customerError(
			w,
			http.StatusBadRequest,
			err.Error(),
		)

	case errors.Is(err, service.ErrCartInsufficientStock):

		customerError(
			w,
			http.StatusConflict,
			err.Error(),
		)

	case errors.Is(err, service.ErrCartVariantUnavailable):

		customerError(
			w,
			http.StatusConflict,
			err.Error(),
		)

	default:
		log.Printf("Cart operation failed: %v", err)

		customerError(
			w,
			http.StatusInternalServerError,
			"failed to process shopping cart",
		)
	}
}

// =========================================
// GET CART
// =========================================

// GET /api/v1/cart
func (h *CartHandler) Get(
	w http.ResponseWriter,
	r *http.Request,
) {
	customerID, ok := getAuthenticatedCustomerID(w, r)
	if !ok {
		return
	}

	cart, err := h.service.Get(r.Context(), customerID)
	if err != nil {
		cartError(w, err)
		return
	}

	writeCustomerJSON(w, http.StatusOK, map[string]any{
		"data": cart,
	})
}

// =========================================
// ADD ITEM
// =========================================

// POST /api/v1/cart/items
func (h *CartHandler) AddItem(
	w http.ResponseWriter,
	r *http.Request,
) {
	customerID, ok := getAuthenticatedCustomerID(w, r)
	if !ok {
		return
	}

	var input AddCartItemRequest

	if !decodeCartRequest(w, r, &input) {
		return
	}

	item, err := h.service.AddItem(
		r.Context(),
		customerID,
		input.ProductID,
		input.VariantID,
		input.Quantity,
	)

	if err != nil {
		cartError(w, err)
		return
	}

	writeCustomerJSON(w, http.StatusCreated, map[string]any{
		"message": "product added to cart successfully",
		"data":    item,
	})
}

// =========================================
// UPDATE ITEM QUANTITY
// =========================================

// PATCH /api/v1/cart/items/{id}
func (h *CartHandler) UpdateQuantity(
	w http.ResponseWriter,
	r *http.Request,
) {
	customerID, ok := getAuthenticatedCustomerID(w, r)
	if !ok {
		return
	}

	itemID, ok := parseCartItemID(w, r)
	if !ok {
		return
	}

	var input UpdateCartItemRequest

	if !decodeCartRequest(w, r, &input) {
		return
	}

	item, err := h.service.UpdateQuantity(
		r.Context(),
		customerID,
		itemID,
		input.Quantity,
	)

	if err != nil {
		cartError(w, err)
		return
	}

	writeCustomerJSON(w, http.StatusOK, map[string]any{
		"message": "cart item quantity updated successfully",
		"data":    item,
	})
}

// =========================================
// REMOVE ITEM
// =========================================

// DELETE /api/v1/cart/items/{id}
func (h *CartHandler) RemoveItem(
	w http.ResponseWriter,
	r *http.Request,
) {
	customerID, ok := getAuthenticatedCustomerID(w, r)
	if !ok {
		return
	}

	itemID, ok := parseCartItemID(w, r)
	if !ok {
		return
	}

	err := h.service.RemoveItem(
		r.Context(),
		customerID,
		itemID,
	)

	if err != nil {
		cartError(w, err)
		return
	}

	writeCustomerJSON(w, http.StatusOK, map[string]string{
		"message": "cart item removed successfully",
	})
}

// =========================================
// CLEAR CART
// =========================================

// DELETE /api/v1/cart/items
func (h *CartHandler) Clear(
	w http.ResponseWriter,
	r *http.Request,
) {
	customerID, ok := getAuthenticatedCustomerID(w, r)
	if !ok {
		return
	}

	err := h.service.Clear(r.Context(), customerID)
	if err != nil {
		cartError(w, err)
		return
	}

	writeCustomerJSON(w, http.StatusOK, map[string]string{
		"message": "cart cleared successfully",
	})
}
