package repository

import (
	"context"

	"github.com/soat-architecture/tech-challenge-project/internal/domain/part"
)

type PartRepository interface {
	Create(ctx context.Context, p *part.Part) error
	Update(ctx context.Context, p *part.Part) error
	Delete(ctx context.Context, id string) error
	FindByID(ctx context.Context, id string) (*part.Part, error)
	List(ctx context.Context, limit, offset int) ([]part.Part, error)
	AdjustStock(ctx context.Context, partID string, movementType part.StockMovementType, quantity int, notes *string, createdByUserID *string) (*part.Part, error)
}

