package admin

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/soat-architecture/tech-challenge-project/internal/domain/service"
	"github.com/soat-architecture/tech-challenge-project/internal/interfaces/repository"
)

type ServiceService struct {
	repo repository.ServiceRepository
}

func NewServiceService(repo repository.ServiceRepository) *ServiceService {
	return &ServiceService{repo: repo}
}

func (s *ServiceService) Create(ctx context.Context, in service.Service) (*service.Service, error) {
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

func (s *ServiceService) Update(ctx context.Context, id string, in service.Service) (*service.Service, error) {
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

func (s *ServiceService) Delete(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}

func (s *ServiceService) FindByID(ctx context.Context, id string) (*service.Service, error) {
	return s.repo.FindByID(ctx, id)
}

func (s *ServiceService) List(ctx context.Context, limit, offset int) ([]service.Service, error) {
	return s.repo.List(ctx, limit, offset)
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
