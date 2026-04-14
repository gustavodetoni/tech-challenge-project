package mapper

import (
	"github.com/soat-architecture/tech-challenge-project/internal/domain/vehicle"
	"github.com/soat-architecture/tech-challenge-project/internal/interfaces/http/dto"
)

func VehicleResponseFromDomain(v vehicle.Vehicle) dto.VehicleResponse {
	return dto.VehicleResponse{
		ID:              v.ID,
		ClientID:        v.ClientID,
		Plate:           v.Plate,
		Brand:           v.Brand,
		Model:           v.Model,
		ManufactureYear: v.ManufactureYear,
		ModelYear:       v.ModelYear,
		Color:           v.Color,
		Mileage:         v.Mileage,
		Chassis:         v.Chassis,
		Notes:           v.Notes,
	}
}
