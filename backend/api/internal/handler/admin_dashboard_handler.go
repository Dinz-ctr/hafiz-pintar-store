package handler

import (
	"log"
	"net/http"

	"hafiz-pintar-store/backend/api/internal/service"
)

type AdminDashboardHandler struct {
	service *service.AdminDashboardService
}

func NewAdminDashboardHandler(
	svc *service.AdminDashboardService,
) *AdminDashboardHandler {
	return &AdminDashboardHandler{
		service: svc,
	}
}

// =========================================
// GET ADMIN DASHBOARD
// =========================================

// GET /api/v1/admin/dashboard
func (h *AdminDashboardHandler) GetDashboard(
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

	// Pastikan request memiliki identitas Admin.
	adminID, ok := AdminIDFromContext(r.Context())
	if !ok || adminID <= 0 {
		customerError(
			w,
			http.StatusUnauthorized,
			"admin authentication required",
		)
		return
	}

	result, err := h.service.GetDashboard(r.Context())
	if err != nil {
		log.Printf(
			"Admin dashboard failed for admin %d: %v",
			adminID,
			err,
		)

		customerError(
			w,
			http.StatusInternalServerError,
			"failed to retrieve admin dashboard",
		)
		return
	}

	writeCustomerJSON(
		w,
		http.StatusOK,
		map[string]any{
			"message": "admin dashboard retrieved successfully",
			"data":    result,
		},
	)
}
