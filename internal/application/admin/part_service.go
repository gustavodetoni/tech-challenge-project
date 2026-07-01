package admin

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"

	repository "github.com/soat-architecture/tech-challenge-project/internal/application/port"
	"github.com/soat-architecture/tech-challenge-project/internal/domain/part"
)

type PartService struct {
	repo repository.PartRepository
}

func NewPartService(repo repository.PartRepository) *PartService { return &PartService{repo: repo} }

type CreatePartInput struct {
	SKU            string
	Name           string
	Description    *string
	UnitPriceCents int64
	StockQuantity  int
	Active         bool
}

type UpdatePartInput = CreatePartInput

func (s *PartService) CreateFromInput(ctx context.Context, input CreatePartInput) (*part.Part, error) {
	return s.Create(ctx, partFromInput(input))
}

func (s *PartService) UpdateFromInput(ctx context.Context, id string, input UpdatePartInput) (*part.Part, error) {
	return s.Update(ctx, id, partFromInput(input))
}

func (s *PartService) Create(ctx context.Context, in part.Part) (*part.Part, error) {
	err := verifyPart(&in, true)
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

func (s *PartService) Update(ctx context.Context, id string, in part.Part) (*part.Part, error) {
	err := verifyPart(&in, false)
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

func partFromInput(input CreatePartInput) part.Part {
	return part.Part{
		SKU:            input.SKU,
		Name:           input.Name,
		Description:    input.Description,
		UnitPriceCents: input.UnitPriceCents,
		StockQuantity:  input.StockQuantity,
		Active:         input.Active,
	}
}

func verifyPart(in *part.Part, verifyStock bool) error {
	in.SKU = strings.TrimSpace(in.SKU)
	in.Name = strings.TrimSpace(in.Name)
	if in.SKU == "" || in.Name == "" {
		return errors.New("sku and name are required")
	}
	if in.UnitPriceCents < 0 {
		return errors.New("invalid price")
	}
	if verifyStock && in.StockQuantity < 0 {
		return errors.New("invalid stock")
	}
	return nil
}
