package handler

import (
	"errors"
	"log"
	"net/http"
	"strings"

	"hafiz-pintar-store/backend/api/internal/repository"
	"hafiz-pintar-store/backend/api/internal/service"
)

type AdminHandler struct {
	service *service.AdminService
}

func NewAdminHandler(svc *service.AdminService) *AdminHandler {
	return &AdminHandler{
		service: svc,
	}
}

// =========================================
// REQUEST DTO
// =========================================

type AdminLoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// =========================================
// ADMIN LOGIN
// =========================================

// POST /api/v1/admin/auth/login
func (h *AdminHandler) Login(
	w http.ResponseWriter,
	r *http.Request,
) {
	if r.Method != http.MethodPost {
		customerError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var input AdminLoginRequest

	if !decodeCartRequest(w, r, &input) {
		return
	}

	input.Email = strings.TrimSpace(input.Email)

	result, err := h.service.Login(
		r.Context(),
		input.Email,
		input.Password,
	)

	if err != nil {
		switch {
		case errors.Is(err, service.ErrAdminInvalidCredentials):
			customerError(
				w,
				http.StatusUnauthorized,
				"invalid email or password",
			)

		case errors.Is(err, service.ErrAdminInactive):
			customerError(
				w,
				http.StatusForbidden,
				"admin account is inactive",
			)

		case errors.Is(err, service.ErrAdminTokenConfig):
			customerError(
				w,
				http.StatusInternalServerError,
				"admin authentication is not configured",
			)

		default:
			log.Printf("Admin login failed: %v", err)
			customerError(
				w,
				http.StatusInternalServerError,
				"failed to login admin",
			)
		}
		return
	}

	writeCustomerJSON(w, http.StatusOK, map[string]any{
		"message": "admin login successful",
		"data":    result,
	})
}

// =========================================
// GET ADMIN PROFILE
// =========================================

// GET /api/v1/admin/me
func (h *AdminHandler) Me(
	w http.ResponseWriter,
	r *http.Request,
) {
	if r.Method != http.MethodGet {
		customerError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	adminID, ok := AdminIDFromContext(r.Context())
	if !ok || adminID <= 0 {
		customerError(
			w,
			http.StatusUnauthorized,
			"admin authentication required",
		)
		return
	}

	admin, err := h.service.GetByID(r.Context(), adminID)
	if err != nil {
		if errors.Is(err, repository.ErrAdminNotFound) {
			customerError(
				w,
				http.StatusUnauthorized,
				"admin account not found",
			)
			return
		}

		log.Printf("Get admin profile failed: %v", err)
		customerError(
			w,
			http.StatusInternalServerError,
			"failed to retrieve admin profile",
		)
		return
	}

	writeCustomerJSON(w, http.StatusOK, map[string]any{
		"message": "admin profile retrieved successfully",
		"data":    admin,
	})
}
