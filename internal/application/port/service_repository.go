package port

import (
	"context"

	"github.com/soat-architecture/tech-challenge-project/internal/domain/service"
)

type ServiceRepository interface {
	Create(ctx context.Context, s *service.Service) error
	Update(ctx context.Context, s *service.Service) error
	Delete(ctx context.Context, id string) error
	FindByID(ctx context.Context, id string) (*service.Service, error)
	FindByIDs(ctx context.Context, ids []string) ([]service.Service, error)
	List(ctx context.Context, limit, offset int) ([]service.Service, error)
}
