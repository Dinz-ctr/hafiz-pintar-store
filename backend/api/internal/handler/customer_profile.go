package handler

import (
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5"
)

// Me mengambil profil customer yang sedang login.
func (h *CustomerHandler) Me(
	w http.ResponseWriter,
	r *http.Request,
) {
	// Ambil customer ID dari JWT middleware.
	customerID, ok := CustomerIDFromContext(r.Context())
	if !ok || customerID <= 0 {
		customerError(
			w,
			http.StatusUnauthorized,
			"customer authentication is required",
		)
		return
	}

	// Ambil profil berdasarkan customer ID dari token.
	customer, err := h.service.GetByID(
		r.Context(),
		customerID,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			customerError(
				w,
				http.StatusNotFound,
				"customer not found",
			)
			return
		}

		log.Printf("Get customer profile failed: %v", err)

		customerError(
			w,
			http.StatusInternalServerError,
			"failed to retrieve customer profile",
		)
		return
	}

	if !customer.IsActive {
		customerError(
			w,
			http.StatusForbidden,
			"customer account is inactive",
		)
		return
	}

	response := CustomerResponse{
		ID:        customer.ID,
		FullName:  customer.FullName,
		Email:     customer.Email,
		Phone:     customer.Phone,
		IsActive:  customer.IsActive,
		CreatedAt: customer.CreatedAt.Time.UTC().Format(time.RFC3339),
		UpdatedAt: customer.UpdatedAt.Time.UTC().Format(time.RFC3339),
	}

	writeCustomerJSON(
		w,
		http.StatusOK,
		map[string]any{
			"message": "customer profile retrieved successfully",
			"data":    response,
		},
	)
}
