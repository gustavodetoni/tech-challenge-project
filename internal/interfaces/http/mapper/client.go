package mapper

import (
	"github.com/soat-architecture/tech-challenge-project/internal/domain/client"
	"github.com/soat-architecture/tech-challenge-project/internal/interfaces/http/dto"
)

func ClientResponseFromDomain(cl client.Client) dto.ClientResponse {
	return dto.ClientResponse{
		ID:             cl.ID,
		DocumentType:   string(cl.DocumentType),
		DocumentNumber: cl.DocumentNumber,
		Name:           cl.Name,
		Email:          cl.Email,
		Phone:          cl.Phone,
	}
}
