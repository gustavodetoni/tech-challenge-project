package admin

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/brazilian-utils/go/licenseplate"
	"github.com/google/uuid"

	"github.com/soat-architecture/tech-challenge-project/internal/domain/vehicle"
	"github.com/soat-architecture/tech-challenge-project/internal/interfaces/repository"
	"github.com/soat-architecture/tech-challenge-project/pkg/br/plate"
)

type VehicleService struct {
	repo repository.VehicleRepository
}

func NewVehicleService(repo repository.VehicleRepository) *VehicleService {
	return &VehicleService{repo: repo}
}

func (s *VehicleService) Create(ctx context.Context, in vehicle.Vehicle) (*vehicle.Vehicle, error) {
	in.Plate = plate.Normalize(in.Plate)
	in.Brand = strings.TrimSpace(in.Brand)
	in.Model = strings.TrimSpace(in.Model)
	if in.ClientID == "" {
		return nil, errors.New("client_id is required")
	}
	if !licenseplate.IsValid(in.Plate, "") {
		return nil, errors.New("invalid plate")
	}
	if in.Brand == "" || in.Model == "" {
		return nil, errors.New("brand and model are required")
	}
	if in.ModelYear < 1900 || in.ModelYear > 2100 {
		return nil, errors.New("invalid model_year")
	}

	now := time.Now().UTC()
	in.ID = uuid.NewString()
	in.CreatedAt = now
	in.UpdatedAt = now

	if err := s.repo.Create(ctx, &in); err != nil {
		return nil, err
	}
	return &in, nil
}

func (s *VehicleService) Update(ctx context.Context, id string, in vehicle.Vehicle) (*vehicle.Vehicle, error) {
	in.Plate = plate.Normalize(in.Plate)
	in.Brand = strings.TrimSpace(in.Brand)
	in.Model = strings.TrimSpace(in.Model)
	if in.ClientID == "" {
		return nil, errors.New("client_id is required")
	}
	if !licenseplate.IsValid(in.Plate, "") {
		return nil, errors.New("invalid plate")
	}
	if in.Brand == "" || in.Model == "" {
		return nil, errors.New("brand and model are required")
	}
	if in.ModelYear < 1900 || in.ModelYear > 2100 {
		return nil, errors.New("invalid model_year")
	}

	in.ID = id
	in.UpdatedAt = time.Now().UTC()
	if err := s.repo.Update(ctx, &in); err != nil {
		return nil, err
	}
	return s.repo.FindByID(ctx, id)
}

func (s *VehicleService) Delete(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}

func (s *VehicleService) FindByID(ctx context.Context, id string) (*vehicle.Vehicle, error) {
	return s.repo.FindByID(ctx, id)
}

func (s *VehicleService) ListByClientID(ctx context.Context, clientID string, limit, offset int) ([]vehicle.Vehicle, error) {
	return s.repo.ListByClientID(ctx, clientID, limit, offset)
}
