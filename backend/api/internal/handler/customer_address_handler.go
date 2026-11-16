package handler

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"strconv"
	"time"

	"hafiz-pintar-store/backend/api/internal/database/sqlc"
	"hafiz-pintar-store/backend/api/internal/repository"
	"hafiz-pintar-store/backend/api/internal/service"

	"github.com/jackc/pgx/v5"
)

type CustomerAddressHandler struct {
	service *service.CustomerAddressService
}

func NewCustomerAddressHandler(
	svc *service.CustomerAddressService,
) *CustomerAddressHandler {
	return &CustomerAddressHandler{
		service: svc,
	}
}

type CustomerAddressRequest struct {
	Label          string  `json:"label"`
	RecipientName  string  `json:"recipient_name"`
	RecipientPhone string  `json:"recipient_phone"`
	AddressLine    string  `json:"address_line"`
	Subdistrict    *string `json:"subdistrict"`
	District       *string `json:"district"`
	City           string  `json:"city"`
	Province       string  `json:"province"`
	PostalCode     *string `json:"postal_code"`
	IsDefault      bool    `json:"is_default"`
}

type CustomerAddressResponse struct {
	ID             int64   `json:"id"`
	CustomerID     int64   `json:"customer_id"`
	Label          string  `json:"label"`
	RecipientName  string  `json:"recipient_name"`
	RecipientPhone string  `json:"recipient_phone"`
	AddressLine    string  `json:"address_line"`
	Subdistrict    *string `json:"subdistrict"`
	District       *string `json:"district"`
	City           string  `json:"city"`
	Province       string  `json:"province"`
	PostalCode     *string `json:"postal_code"`
	IsDefault      bool    `json:"is_default"`
	CreatedAt      string  `json:"created_at"`
	UpdatedAt      string  `json:"updated_at"`
}

func mapCustomerAddress(
	address sqlc.CustomerAddress,
) CustomerAddressResponse {
	return CustomerAddressResponse{
		ID:             address.ID,
		CustomerID:     address.CustomerID,
		Label:          address.Label,
		RecipientName:  address.RecipientName,
		RecipientPhone: address.RecipientPhone,
		AddressLine:    address.AddressLine,
		Subdistrict:    address.Subdistrict,
		District:       address.District,
		City:           address.City,
		Province:       address.Province,
		PostalCode:     address.PostalCode,
		IsDefault:      address.IsDefault,
		CreatedAt:      address.CreatedAt.Time.UTC().Format(time.RFC3339),
		UpdatedAt:      address.UpdatedAt.Time.UTC().Format(time.RFC3339),
	}
}

func getAuthenticatedCustomerID(
	w http.ResponseWriter,
	r *http.Request,
) (int64, bool) {
	customerID, ok := CustomerIDFromContext(r.Context())

	if !ok || customerID <= 0 {
		customerError(
			w,
			http.StatusUnauthorized,
			"customer authentication is required",
		)
		return 0, false
	}

	return customerID, true
}

func parseAddressID(
	w http.ResponseWriter,
	r *http.Request,
) (int64, bool) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)

	if err != nil || id <= 0 {
		customerError(
			w,
			http.StatusBadRequest,
			"invalid address ID",
		)
		return 0, false
	}

	return id, true
}

func decodeCustomerAddressRequest(
	w http.ResponseWriter,
	r *http.Request,
) (CustomerAddressRequest, bool) {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	defer r.Body.Close()

	var input CustomerAddressRequest

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&input); err != nil {
		customerError(
			w,
			http.StatusBadRequest,
			"invalid request body",
		)
		return input, false
	}

	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		customerError(
			w,
			http.StatusBadRequest,
			"request body must contain only one JSON object",
		)
		return input, false
	}

	return input, true
}

func customerAddressError(
	w http.ResponseWriter,
	err error,
) {
	switch {
	case errors.Is(err, pgx.ErrNoRows),
		errors.Is(err, repository.ErrCustomerAddressNotFound):
		customerError(
			w,
			http.StatusNotFound,
			"customer address not found",
		)

	case errors.Is(err, service.ErrInvalidAddressCustomerID),
		errors.Is(err, service.ErrInvalidAddressID),
		errors.Is(err, service.ErrInvalidAddressLabel),
		errors.Is(err, service.ErrInvalidRecipientName),
		errors.Is(err, service.ErrInvalidRecipientPhone),
		errors.Is(err, service.ErrInvalidAddressLine),
		errors.Is(err, service.ErrInvalidAddressCity),
		errors.Is(err, service.ErrInvalidAddressProvince):
		customerError(
			w,
			http.StatusBadRequest,
			err.Error(),
		)

	default:
		log.Printf("Customer address operation failed: %v", err)
		customerError(
			w,
			http.StatusInternalServerError,
			"failed to process customer address",
		)
	}
}

// GET /api/v1/customer-addresses
func (h *CustomerAddressHandler) List(
	w http.ResponseWriter,
	r *http.Request,
) {
	customerID, ok := getAuthenticatedCustomerID(w, r)
	if !ok {
		return
	}

	addresses, err := h.service.List(r.Context(), customerID)
	if err != nil {
		customerAddressError(w, err)
		return
	}

	response := make(
		[]CustomerAddressResponse,
		0,
		len(addresses),
	)

	for _, address := range addresses {
		response = append(response, mapCustomerAddress(address))
	}

	writeCustomerJSON(w, http.StatusOK, map[string]any{
		"data":  response,
		"total": len(response),
	})
}

// GET /api/v1/customer-addresses/{id}
func (h *CustomerAddressHandler) GetByID(
	w http.ResponseWriter,
	r *http.Request,
) {
	customerID, ok := getAuthenticatedCustomerID(w, r)
	if !ok {
		return
	}

	id, ok := parseAddressID(w, r)
	if !ok {
		return
	}

	address, err := h.service.GetByID(
		r.Context(),
		id,
		customerID,
	)
	if err != nil {
		customerAddressError(w, err)
		return
	}

	writeCustomerJSON(w, http.StatusOK, map[string]any{
		"data": mapCustomerAddress(address),
	})
}

// POST /api/v1/customer-addresses
func (h *CustomerAddressHandler) Create(
	w http.ResponseWriter,
	r *http.Request,
) {
	customerID, ok := getAuthenticatedCustomerID(w, r)
	if !ok {
		return
	}

	input, ok := decodeCustomerAddressRequest(w, r)
	if !ok {
		return
	}

	address, err := h.service.Create(
		r.Context(),
		sqlc.CreateCustomerAddressParams{
			CustomerID:     customerID,
			Label:          input.Label,
			RecipientName:  input.RecipientName,
			RecipientPhone: input.RecipientPhone,
			AddressLine:    input.AddressLine,
			Subdistrict:    input.Subdistrict,
			District:       input.District,
			City:           input.City,
			Province:       input.Province,
			PostalCode:     input.PostalCode,
			IsDefault:      input.IsDefault,
		},
	)
	if err != nil {
		customerAddressError(w, err)
		return
	}

	writeCustomerJSON(w, http.StatusCreated, map[string]any{
		"message": "customer address created successfully",
		"data":    mapCustomerAddress(address),
	})
}

// PUT /api/v1/customer-addresses/{id}
func (h *CustomerAddressHandler) Update(
	w http.ResponseWriter,
	r *http.Request,
) {
	customerID, ok := getAuthenticatedCustomerID(w, r)
	if !ok {
		return
	}

	id, ok := parseAddressID(w, r)
	if !ok {
		return
	}

	input, ok := decodeCustomerAddressRequest(w, r)
	if !ok {
		return
	}

	address, err := h.service.Update(
		r.Context(),
		sqlc.UpdateCustomerAddressParams{
			ID:             id,
			CustomerID:     customerID,
			Label:          input.Label,
			RecipientName:  input.RecipientName,
			RecipientPhone: input.RecipientPhone,
			AddressLine:    input.AddressLine,
			Subdistrict:    input.Subdistrict,
			District:       input.District,
			City:           input.City,
			Province:       input.Province,
			PostalCode:     input.PostalCode,
		},
	)
	if err != nil {
		customerAddressError(w, err)
		return
	}

	writeCustomerJSON(w, http.StatusOK, map[string]any{
		"message": "customer address updated successfully",
		"data":    mapCustomerAddress(address),
	})
}

// DELETE /api/v1/customer-addresses/{id}
func (h *CustomerAddressHandler) Delete(
	w http.ResponseWriter,
	r *http.Request,
) {
	customerID, ok := getAuthenticatedCustomerID(w, r)
	if !ok {
		return
	}

	id, ok := parseAddressID(w, r)
	if !ok {
		return
	}

	_, err := h.service.Delete(
		r.Context(),
		id,
		customerID,
	)
	if err != nil {
		customerAddressError(w, err)
		return
	}

	writeCustomerJSON(w, http.StatusOK, map[string]string{
		"message": "customer address deleted successfully",
	})
}

// PATCH /api/v1/customer-addresses/{id}/default
func (h *CustomerAddressHandler) SetDefault(
	w http.ResponseWriter,
	r *http.Request,
) {
	customerID, ok := getAuthenticatedCustomerID(w, r)
	if !ok {
		return
	}

	id, ok := parseAddressID(w, r)
	if !ok {
		return
	}

	err := h.service.SetDefault(
		r.Context(),
		id,
		customerID,
	)
	if err != nil {
		customerAddressError(w, err)
		return
	}

	writeCustomerJSON(w, http.StatusOK, map[string]string{
		"message": "default address updated successfully",
	})
}
