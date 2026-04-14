package mapper

import (
	"github.com/soat-architecture/tech-challenge-project/internal/domain/part"
	"github.com/soat-architecture/tech-challenge-project/internal/interfaces/http/dto"
)

func PartResponseFromDomain(p part.Part) dto.PartResponse {
	return dto.PartResponse{
		ID:             p.ID,
		SKU:            p.SKU,
		Name:           p.Name,
		Description:    p.Description,
		UnitPriceCents: p.UnitPriceCents,
		StockQuantity:  p.StockQuantity,
		Active:         p.Active,
	}
}
