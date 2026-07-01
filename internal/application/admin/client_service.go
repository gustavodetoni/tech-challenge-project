package admin

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"

	repository "github.com/soat-architecture/tech-challenge-project/internal/application/port"
	"github.com/soat-architecture/tech-challenge-project/internal/domain/client"
	"github.com/soat-architecture/tech-challenge-project/pkg/br/document"
)

type ClientAdminUseCase struct {
	repo repository.ClientRepository
}

func NewClientAdminUseCase(repo repository.ClientRepository) *ClientAdminUseCase {
	return &ClientAdminUseCase{repo: repo}
}

func NewClientService(repo repository.ClientRepository) *ClientAdminUseCase {
	return NewClientAdminUseCase(repo)
}

type ClientService = ClientAdminUseCase

type CreateClientInput struct {
	DocumentType   string
	DocumentNumber string
	Name           string
	Email          *string
	Phone          *string
}

type UpdateClientInput = CreateClientInput

func (s *ClientAdminUseCase) CreateFromInput(ctx context.Context, input CreateClientInput) (*client.Client, error) {
	return s.Create(ctx, clientFromInput(input))
}

func (s *ClientAdminUseCase) UpdateFromInput(ctx context.Context, id string, input UpdateClientInput) (*client.Client, error) {
	return s.Update(ctx, id, clientFromInput(input))
}

func (s *ClientAdminUseCase) Create(ctx context.Context, in client.Client) (*client.Client, error) {
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

func (s *ClientAdminUseCase) Update(ctx context.Context, id string, in client.Client) (*client.Client, error) {
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

func (s *ClientAdminUseCase) Delete(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}

func (s *ClientAdminUseCase) FindByID(ctx context.Context, id string) (*client.Client, error) {
	return s.repo.FindByID(ctx, id)
}

func (s *ClientAdminUseCase) List(ctx context.Context, limit, offset int) ([]client.Client, error) {
	return s.repo.List(ctx, limit, offset)
}

func clientFromInput(input CreateClientInput) client.Client {
	return client.Client{
		DocumentType:   client.DocumentType(input.DocumentType),
		DocumentNumber: input.DocumentNumber,
		Name:           input.Name,
		Email:          input.Email,
		Phone:          input.Phone,
	}
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
