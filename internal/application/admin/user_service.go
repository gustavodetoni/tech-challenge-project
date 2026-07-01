package admin

import (
	"context"
	"errors"
	"fmt"
	"strings"

	repository "github.com/soat-architecture/tech-challenge-project/internal/application/port"
	"github.com/soat-architecture/tech-challenge-project/internal/domain/user"
)

var ErrInvalidInput = errors.New("invalid input")

type UserService struct {
	repo repository.UserRepository
}

func NewUserService(repo repository.UserRepository) *UserService { return &UserService{repo: repo} }

func (s *UserService) UpdateRole(ctx context.Context, id string, role user.Role) error {
	id = strings.TrimSpace(id)
	if id == "" {
		return fmt.Errorf("%w: id is required", ErrInvalidInput)
	}
	if !isValidRole(role) {
		return fmt.Errorf("%w: invalid role", ErrInvalidInput)
	}
	return s.repo.UpdateRole(ctx, id, role)
}

func isValidRole(role user.Role) bool {
	switch role {
	case user.RoleAdmin, user.RoleManager, user.RoleMechanic, user.RoleAttendant, user.RoleViewer:
		return true
	default:
		return false
	}
}
