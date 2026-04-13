package repository

import (
	"context"

	"github.com/soat-architecture/tech-challenge-project/internal/domain/vehicle"
)

type VehicleRepository interface {
	Create(ctx context.Context, v *vehicle.Vehicle) error
	Update(ctx context.Context, v *vehicle.Vehicle) error
	Delete(ctx context.Context, id string) error
	FindByID(ctx context.Context, id string) (*vehicle.Vehicle, error)
	ListByClientID(ctx context.Context, clientID string, limit, offset int) ([]vehicle.Vehicle, error)
}

