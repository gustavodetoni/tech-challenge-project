package repository

import (
	"context"
	"time"

	"github.com/soat-architecture/tech-challenge-project/internal/domain/order"
)

type ServiceOrderRepository interface {
	List(ctx context.Context, limit, offset int, status *order.Status) ([]order.ServiceOrderSummary, error)
	FindByID(ctx context.Context, id string) (*order.ServiceOrder, error)
	GetDetailByID(ctx context.Context, id string) (*order.ServiceOrderDetail, error)
	AverageExecutionMinutes(ctx context.Context, from, to *time.Time) (float64, error)
	AverageServiceExecutionMinutes(ctx context.Context, serviceID *string, from, to *time.Time) ([]ServiceExecutionAverage, error)
}

type ServiceExecutionAverage struct {
	ServiceID      *string
	Description    string
	AverageMinutes float64
	SampleCount    int64
}
