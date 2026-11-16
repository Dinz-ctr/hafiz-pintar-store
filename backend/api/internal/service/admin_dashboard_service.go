package service

import (
	"context"

	"hafiz-pintar-store/backend/api/internal/repository"
)

type AdminDashboardService struct {
	repo *repository.AdminDashboardRepository
}

func NewAdminDashboardService(
	repo *repository.AdminDashboardRepository,
) *AdminDashboardService {
	return &AdminDashboardService{
		repo: repo,
	}
}

type AdminDashboardResult struct {
	Summary      repository.AdminDashboardSummary `json:"summary"`
	RecentOrders []repository.AdminRecentOrder    `json:"recent_orders"`
}

// GET ADMIN DASHBOARD
func (s *AdminDashboardService) GetDashboard(
	ctx context.Context,
) (AdminDashboardResult, error) {
	summary, err := s.repo.GetSummary(ctx)
	if err != nil {
		return AdminDashboardResult{}, err
	}

	recentOrders, err := s.repo.GetRecentOrders(ctx, 5)
	if err != nil {
		return AdminDashboardResult{}, err
	}

	if recentOrders == nil {
		recentOrders = make([]repository.AdminRecentOrder, 0)
	}

	return AdminDashboardResult{
		Summary:      summary,
		RecentOrders: recentOrders,
	}, nil
}
