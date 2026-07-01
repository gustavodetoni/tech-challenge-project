package admin

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/brazilian-utils/go/licenseplate"
	"github.com/google/uuid"

	repository "github.com/soat-architecture/tech-challenge-project/internal/application/port"
	"github.com/soat-architecture/tech-challenge-project/internal/domain/vehicle"
	"github.com/soat-architecture/tech-challenge-project/pkg/br/plate"
)

type VehicleAdminUseCase struct {
	repo repository.VehicleRepository
}

func NewVehicleAdminUseCase(repo repository.VehicleRepository) *VehicleAdminUseCase {
	return &VehicleAdminUseCase{repo: repo}
}

func NewVehicleService(repo repository.VehicleRepository) *VehicleAdminUseCase {
	return NewVehicleAdminUseCase(repo)
}

type VehicleService = VehicleAdminUseCase

type CreateVehicleInput struct {
	ClientID        string
	Plate           string
	Brand           string
	Model           string
	ManufactureYear *int
	ModelYear       int
	Color           *string
	Mileage         *int
	Chassis         *string
	Notes           *string
}

type UpdateVehicleInput = CreateVehicleInput

func (s *VehicleAdminUseCase) CreateFromInput(ctx context.Context, input CreateVehicleInput) (*vehicle.Vehicle, error) {
	return s.Create(ctx, vehicleFromInput(input))
}

func (s *VehicleAdminUseCase) UpdateFromInput(ctx context.Context, id string, input UpdateVehicleInput) (*vehicle.Vehicle, error) {
	return s.Update(ctx, id, vehicleFromInput(input))
}

func (s *VehicleAdminUseCase) Create(ctx context.Context, in vehicle.Vehicle) (*vehicle.Vehicle, error) {
	err := verifyVehicle(&in)
	if err != nil {
		return nil, err
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

func (s *VehicleAdminUseCase) Update(ctx context.Context, id string, in vehicle.Vehicle) (*vehicle.Vehicle, error) {
	err := verifyVehicle(&in)
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

func (s *VehicleAdminUseCase) Delete(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}

func (s *VehicleAdminUseCase) FindByID(ctx context.Context, id string) (*vehicle.Vehicle, error) {
	return s.repo.FindByID(ctx, id)
}

func (s *VehicleAdminUseCase) ListByClientID(ctx context.Context, clientID string, limit, offset int) ([]vehicle.Vehicle, error) {
	return s.repo.ListByClientID(ctx, clientID, limit, offset)
}

func vehicleFromInput(input CreateVehicleInput) vehicle.Vehicle {
	return vehicle.Vehicle{
		ClientID:        input.ClientID,
		Plate:           input.Plate,
		Brand:           input.Brand,
		Model:           input.Model,
		ManufactureYear: input.ManufactureYear,
		ModelYear:       input.ModelYear,
		Color:           input.Color,
		Mileage:         input.Mileage,
		Chassis:         input.Chassis,
		Notes:           input.Notes,
	}
}

func verifyVehicle(in *vehicle.Vehicle) error {
	in.Plate = plate.Normalize(in.Plate)
	in.Brand = strings.TrimSpace(in.Brand)
	in.Model = strings.TrimSpace(in.Model)
	if in.ClientID == "" {
		return errors.New("client_id is required")
	}
	if !licenseplate.IsValid(in.Plate, "") {
		return errors.New("invalid plate")
	}
	if in.Brand == "" || in.Model == "" {
		return errors.New("brand and model are required")
	}
	if in.ModelYear < 1900 || in.ModelYear > 2100 {
		return errors.New("invalid model_year")
	}
	return nil
}
