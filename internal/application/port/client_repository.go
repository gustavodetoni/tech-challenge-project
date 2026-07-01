package port

import (
	"context"

	"github.com/soat-architecture/tech-challenge-project/internal/domain/client"
)

type ClientRepository interface {
	Create(ctx context.Context, c *client.Client) error
	Update(ctx context.Context, c *client.Client) error
	Delete(ctx context.Context, id string) error
	FindByID(ctx context.Context, id string) (*client.Client, error)
	FindByDocument(ctx context.Context, documentNumber string) (*client.Client, error)
	List(ctx context.Context, limit, offset int) ([]client.Client, error)
}
