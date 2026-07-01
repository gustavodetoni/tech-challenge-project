package admin

import (
	"context"
	"time"

	repository "github.com/soat-architecture/tech-challenge-project/internal/application/port"
	"github.com/soat-architecture/tech-challenge-project/internal/domain/order"
)

type ServiceOrderService struct {
	repo repository.ServiceOrderRepository
}

func NewServiceOrderService(repo repository.ServiceOrderRepository) *ServiceOrderService {
	return &ServiceOrderService{repo: repo}
}

func (s *ServiceOrderService) List(ctx context.Context, limit, offset int, status *order.Status) ([]order.ServiceOrderSummary, error) {
	return s.repo.List(ctx, limit, offset, status)
}

func (s *ServiceOrderService) FindByID(ctx context.Context, id string) (*order.ServiceOrder, error) {
	return s.repo.FindByID(ctx, id)
}

func (s *ServiceOrderService) GetDetailByID(ctx context.Context, id string) (*order.ServiceOrderDetail, error) {
	return s.repo.GetDetailByID(ctx, id)
}

func (s *ServiceOrderService) AverageExecutionMinutes(ctx context.Context, from, to *time.Time) (float64, error) {
	return s.repo.AverageExecutionMinutes(ctx, from, to)
}

func (s *ServiceOrderService) AverageServiceExecutionMinutes(ctx context.Context, serviceID *string, from, to *time.Time) ([]repository.ServiceExecutionAverage, error) {
	return s.repo.AverageServiceExecutionMinutes(ctx, serviceID, from, to)
}
