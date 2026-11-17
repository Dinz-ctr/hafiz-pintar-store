package service

import (
	"context"
	"errors"
	"strings"

	"hafiz-pintar-store/backend/api/internal/repository"
)

var (
	ErrInvalidAdminOrderID     = errors.New("invalid admin order ID")
	ErrInvalidAdminOrderStatus = errors.New("invalid admin order status")
	ErrInvalidAdminOrderFilter = errors.New("invalid admin order filter")
	ErrInvalidOrderTransition  = errors.New("invalid order status transition")
	ErrInvalidShippingDetails  = errors.New("invalid shipping details")
)

type AdminOrderService struct {
	repo *repository.AdminOrderRepository
}

func NewAdminOrderService(
	repo *repository.AdminOrderRepository,
) *AdminOrderService {
	return &AdminOrderService{repo: repo}
}

type AdminOrderListResult struct {
	Data       []repository.AdminOrderSummary `json:"data"`
	Page       int32                          `json:"page"`
	Limit      int32                          `json:"limit"`
	Total      int64                          `json:"total"`
	TotalPages int64                          `json:"total_pages"`
	Keyword    string                         `json:"keyword"`
	Status     string                         `json:"status"`
}

func (s *AdminOrderService) List(
	ctx context.Context,
	page int32,
	limit int32,
	keyword string,
	status string,
) (AdminOrderListResult, error) {
	if page <= 0 {
		page = 1
	}
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	keyword = strings.TrimSpace(keyword)
	status = strings.ToLower(strings.TrimSpace(status))

	if status != "" && !isValidAdminOrderStatus(status) {
		return AdminOrderListResult{}, ErrInvalidAdminOrderFilter
	}

	offset := (page - 1) * limit

	orders, total, err := s.repo.List(
		ctx,
		repository.AdminOrderListParams{
			Keyword: keyword,
			Status:  status,
			Limit:   limit,
			Offset:  offset,
		},
	)
	if err != nil {
		return AdminOrderListResult{}, err
	}

	totalPages := int64(0)
	if total > 0 {
		totalPages = (total + int64(limit) - 1) / int64(limit)
	}

	return AdminOrderListResult{
		Data:       orders,
		Page:       page,
		Limit:      limit,
		Total:      total,
		TotalPages: totalPages,
		Keyword:    keyword,
		Status:     status,
	}, nil
}

func (s *AdminOrderService) GetByID(
	ctx context.Context,
	orderID int64,
) (repository.AdminOrderDetail, error) {
	if orderID <= 0 {
		return repository.AdminOrderDetail{}, ErrInvalidAdminOrderID
	}

	return s.repo.GetByID(ctx, orderID)
}

func (s *AdminOrderService) UpdateStatus(
	ctx context.Context,
	orderID int64,
	newStatus string,
) (repository.AdminOrderSummary, error) {
	if orderID <= 0 {
		return repository.AdminOrderSummary{}, ErrInvalidAdminOrderID
	}

	newStatus = strings.ToLower(strings.TrimSpace(newStatus))
	if !isValidAdminOrderStatus(newStatus) {
		return repository.AdminOrderSummary{}, ErrInvalidAdminOrderStatus
	}

	currentStatus, err := s.repo.GetStatus(ctx, orderID)
	if err != nil {
		return repository.AdminOrderSummary{}, err
	}

	if !canTransitionOrderStatus(currentStatus, newStatus) {
		return repository.AdminOrderSummary{}, ErrInvalidOrderTransition
	}

	// Pembatalan menggunakan transaksi khusus:
	// restore stok + inventory movement + update status.
	if newStatus == "cancelled" {
		return s.repo.CancelOrder(
			ctx,
			orderID,
			currentStatus,
		)
	}

	return s.repo.UpdateStatus(
		ctx,
		orderID,
		currentStatus,
		newStatus,
	)
}

// UpdateShipping menyimpan kurir dan nomor resi,
// sekaligus mengubah status pesanan menjadi shipped.
func (s *AdminOrderService) UpdateShipping(
	ctx context.Context,
	orderID int64,
	courierName string,
	trackingNumber string,
) (repository.AdminOrderDetail, error) {
	if orderID <= 0 {
		return repository.AdminOrderDetail{}, ErrInvalidAdminOrderID
	}

	courierName = strings.TrimSpace(courierName)
	trackingNumber = strings.TrimSpace(trackingNumber)

	if courierName == "" || trackingNumber == "" {
		return repository.AdminOrderDetail{}, ErrInvalidShippingDetails
	}

	if len(courierName) > 100 || len(trackingNumber) > 100 {
		return repository.AdminOrderDetail{}, ErrInvalidShippingDetails
	}

	return s.repo.UpdateShipping(
		ctx,
		orderID,
		courierName,
		trackingNumber,
	)
}

func isValidAdminOrderStatus(status string) bool {
	switch status {
	case "pending",
		"confirmed",
		"processing",
		"shipped",
		"completed",
		"cancelled":
		return true
	default:
		return false
	}
}

func canTransitionOrderStatus(current, next string) bool {
	switch current {
	case "pending":
		return next == "confirmed" || next == "cancelled"

	case "confirmed":
		return next == "processing" || next == "cancelled"

	case "processing":
		// Pengiriman harus melalui endpoint UpdateShipping.
		return false

	case "shipped":
		return next == "completed"

	default:
		return false
	}
}
