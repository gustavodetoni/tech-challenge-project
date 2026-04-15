package admin

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/soat-architecture/tech-challenge-project/internal/domain/client"
	"github.com/soat-architecture/tech-challenge-project/internal/interfaces/repository"
	"github.com/soat-architecture/tech-challenge-project/pkg/br/document"
)

type ClientService struct {
	repo repository.ClientRepository
}

func NewClientService(repo repository.ClientRepository) *ClientService {
	return &ClientService{repo: repo}
}

func (s *ClientService) Create(ctx context.Context, in client.Client) (*client.Client, error) {
	err := verifyClient(&in)
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

func (s *ClientService) Update(ctx context.Context, id string, in client.Client) (*client.Client, error) {
	err := verifyClient(&in)
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
		return document.IsValidCPF(doc)
	case client.DocumentTypeCNPJ:
		return document.IsValidCNPJ(doc)
	default:
		return false
	}
}

func verifyClient(in *client.Client) error {
	in.Name = strings.TrimSpace(in.Name)
	in.DocumentNumber = document.Normalize(in.DocumentNumber)
	if in.Name == "" {
		return errors.New("name is required")
	}
	if !isValidDocument(in.DocumentType, in.DocumentNumber) {
		return errors.New("invalid document")
	}
	return nil
}
