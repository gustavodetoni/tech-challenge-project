package repository

import (
	"context"
	"errors"

	"github.com/soat-architecture/tech-challenge-project/internal/domain/user"
)

var (
	ErrNotFound = errors.New("not found")
	ErrConflict = errors.New("conflict")
)

type UserRepository interface {
	Create(ctx context.Context, u *user.User) error
	FindByEmail(ctx context.Context, email string) (*user.User, error)
	FindByID(ctx context.Context, id string) (*user.User, error)
}
