package admin

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/soat-architecture/tech-challenge-project/internal/domain/part"
	"github.com/soat-architecture/tech-challenge-project/internal/interfaces/repository"
)

type PartService struct {
	repo repository.PartRepository
}

func NewPartService(repo repository.PartRepository) *PartService { return &PartService{repo: repo} }

func (s *PartService) Create(ctx context.Context, in part.Part) (*part.Part, error) {
	in.SKU = strings.TrimSpace(in.SKU)
	in.Name = strings.TrimSpace(in.Name)
	if in.SKU == "" || in.Name == "" {
		return nil, errors.New("sku and name are required")
	}
	if in.UnitPriceCents < 0 || in.StockQuantity < 0 {
		return nil, errors.New("invalid price/stock")
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

func (s *PartService) Update(ctx context.Context, id string, in part.Part) (*part.Part, error) {
	in.SKU = strings.TrimSpace(in.SKU)
	in.Name = strings.TrimSpace(in.Name)
	if in.SKU == "" || in.Name == "" {
		return nil, errors.New("sku and name are required")
	}
	if in.UnitPriceCents < 0 {
		return nil, errors.New("invalid price")
	}

	in.ID = id
	in.UpdatedAt = time.Now().UTC()
	if err := s.repo.Update(ctx, &in); err != nil {
		return nil, err
	}
	return s.repo.FindByID(ctx, id)
}

func (s *PartService) Delete(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}

func (s *PartService) FindByID(ctx context.Context, id string) (*part.Part, error) {
	return s.repo.FindByID(ctx, id)
}

func (s *PartService) List(ctx context.Context, limit, offset int) ([]part.Part, error) {
	return s.repo.List(ctx, limit, offset)
}

func (s *PartService) AdjustStock(ctx context.Context, partID string, movementType part.StockMovementType, quantity int, notes *string, createdByUserID *string) (*part.Part, error) {
	return s.repo.AdjustStock(ctx, partID, movementType, quantity, notes, createdByUserID)
}

