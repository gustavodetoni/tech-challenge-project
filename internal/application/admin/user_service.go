package admin

import (
	"context"
	"fmt"
	"strings"

	repository "github.com/soat-architecture/tech-challenge-project/internal/application/port"
	"github.com/soat-architecture/tech-challenge-project/internal/domain/user"
	"github.com/soat-architecture/tech-challenge-project/internal/infra/expections"
)

var ErrInvalidInput = expections.New(expections.CodeValidation, "invalid input")

type UserAdminUseCase struct {
	repo repository.UserRepository
}

func NewUserAdminUseCase(repo repository.UserRepository) *UserAdminUseCase {
	return &UserAdminUseCase{repo: repo}
}

func NewUserService(repo repository.UserRepository) *UserAdminUseCase {
	return NewUserAdminUseCase(repo)
}

type UserService = UserAdminUseCase

type UpdateUserRoleInput struct {
	UserID string
	Role   string
}

func (s *UserAdminUseCase) UpdateRoleFromInput(ctx context.Context, input UpdateUserRoleInput) error {
	return s.UpdateRole(ctx, input.UserID, user.Role(input.Role))
}

func (s *UserAdminUseCase) UpdateRole(ctx context.Context, id string, role user.Role) error {
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
