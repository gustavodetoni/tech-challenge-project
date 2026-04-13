package admin

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/brazilian-utils/go/cnpj"
	"github.com/brazilian-utils/go/cpf"
	"github.com/google/uuid"

	"github.com/soat-architecture/tech-challenge-project/internal/domain/client"
	"github.com/soat-architecture/tech-challenge-project/internal/interfaces/repository"
)

type ClientService struct {
	repo repository.ClientRepository
}

func NewClientService(repo repository.ClientRepository) *ClientService {
	return &ClientService{repo: repo}
}

func (s *ClientService) Create(ctx context.Context, in client.Client) (*client.Client, error) {
	in.Name = strings.TrimSpace(in.Name)
	in.DocumentNumber = strings.TrimSpace(in.DocumentNumber)
	if in.Name == "" {
		return nil, errors.New("name is required")
	}
	if !isValidDocument(in.DocumentType, in.DocumentNumber) {
		return nil, errors.New("invalid document")
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

func (s *ClientService) Update(ctx context.Context, id string, in client.Client) (*client.Client, error) {
	in.Name = strings.TrimSpace(in.Name)
	in.DocumentNumber = strings.TrimSpace(in.DocumentNumber)
	if in.Name == "" {
		return nil, errors.New("name is required")
	}
	if !isValidDocument(in.DocumentType, in.DocumentNumber) {
		return nil, errors.New("invalid document")
	}

	in.ID = id
	in.UpdatedAt = time.Now().UTC()
	if err := s.repo.Update(ctx, &in); err != nil {
		return nil, err
	}

	return s.repo.FindByID(ctx, id)
}

func (s *ClientService) Delete(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}

func (s *ClientService) FindByID(ctx context.Context, id string) (*client.Client, error) {
	return s.repo.FindByID(ctx, id)
}

func (s *ClientService) List(ctx context.Context, limit, offset int) ([]client.Client, error) {
	return s.repo.List(ctx, limit, offset)
}

func isValidDocument(docType client.DocumentType, doc string) bool {
	switch docType {
	case client.DocumentTypeCPF:
		return cpf.IsValid(doc)
	case client.DocumentTypeCNPJ:
		return cnpj.IsValid(doc)
	default:
		return false
	}
}
