package mapper

import (
	"github.com/soat-architecture/tech-challenge-project/internal/domain/service"
	"github.com/soat-architecture/tech-challenge-project/internal/interfaces/http/dto"
)

func ServiceResponseFromDomain(s service.Service) dto.ServiceResponse {
	return dto.ServiceResponse{
		ID:               s.ID,
		Name:             s.Name,
		Description:      s.Description,
		BasePriceCents:   s.BasePriceCents,
		EstimatedMinutes: s.EstimatedMinutes,
		Active:           s.Active,
	}
}
