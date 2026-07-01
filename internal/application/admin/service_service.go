package admin

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"

	repository "github.com/soat-architecture/tech-challenge-project/internal/application/port"
	"github.com/soat-architecture/tech-challenge-project/internal/domain/service"
)

type ServiceCatalogUseCase struct {
	repo repository.ServiceRepository
}

func NewServiceCatalogUseCase(repo repository.ServiceRepository) *ServiceCatalogUseCase {
	return &ServiceCatalogUseCase{repo: repo}
}

type CreateServiceInput struct {
	Name             string
	Description      *string
	BasePriceCents   int64
	EstimatedMinutes int
	Active           bool
}

type UpdateServiceInput = CreateServiceInput

func (s *ServiceCatalogUseCase) CreateFromInput(ctx context.Context, input CreateServiceInput) (*service.Service, error) {
	return s.Create(ctx, serviceFromInput(input))
}

func (s *ServiceCatalogUseCase) UpdateFromInput(ctx context.Context, id string, input UpdateServiceInput) (*service.Service, error) {
	return s.Update(ctx, id, serviceFromInput(input))
}

func (s *ServiceCatalogUseCase) Create(ctx context.Context, in service.Service) (*service.Service, error) {
	err := verifyService(&in)
	if err != nil {
		return nil, err
	}
	now := time.Now().UTC()
	in.ID = uuid.NewString()
	in.Active = true
	in.CreatedAt = now
	in.UpdatedAt = now

	if err := s.repo.Create(ctx, &in); err != nil {
		return nil, err
	}
	return &in, nil
}

func (s *ServiceCatalogUseCase) Update(ctx context.Context, id string, in service.Service) (*service.Service, error) {
	err := verifyService(&in)
	if err != nil {
		return nil, err
	}
	in.ID = id
	in.UpdatedAt = time.Now().UTC()
	if err := s.repo.Update(ctx, &in); err != nil {
		return nil, err
	}
	return s.repo.FindByID(ctx, id)
}

func (s *ServiceCatalogUseCase) Delete(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}

func (s *ServiceCatalogUseCase) FindByID(ctx context.Context, id string) (*service.Service, error) {
	return s.repo.FindByID(ctx, id)
}

func (s *ServiceCatalogUseCase) List(ctx context.Context, limit, offset int) ([]service.Service, error) {
	return s.repo.List(ctx, limit, offset)
}

func serviceFromInput(input CreateServiceInput) service.Service {
	return service.Service{
		Name:             input.Name,
		Description:      input.Description,
		BasePriceCents:   input.BasePriceCents,
		EstimatedMinutes: input.EstimatedMinutes,
		Active:           input.Active,
	}
}

func verifyService(in *service.Service) error {
	in.Name = strings.TrimSpace(in.Name)
	if in.Name == "" {
		return errors.New("name is required")
	}
	if in.BasePriceCents < 0 || in.EstimatedMinutes < 0 {
		return errors.New("invalid price/estimated minutes")
	}
	return nil
}
