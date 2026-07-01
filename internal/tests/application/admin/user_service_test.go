package admin_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/soat-architecture/tech-challenge-project/internal/application/admin"
	repository "github.com/soat-architecture/tech-challenge-project/internal/application/port"
	"github.com/soat-architecture/tech-challenge-project/internal/domain/user"
	repomocks "github.com/soat-architecture/tech-challenge-project/internal/tests/interfaces/repository/mocks"
)

func TestUserService_UpdateRole_Success(t *testing.T) {
	t.Parallel()

	repo := new(repomocks.UserRepository)
	svc := admin.NewUserService(repo)

	repo.On("UpdateRole", mock.Anything, "u1", user.RoleManager).Return(nil).Once()

	err := svc.UpdateRole(context.Background(), "u1", user.RoleManager)
	require.NoError(t, err)
	repo.AssertExpectations(t)
}

func TestUserService_UpdateRole_InvalidID(t *testing.T) {
	t.Parallel()

	repo := new(repomocks.UserRepository)
	svc := admin.NewUserService(repo)

	err := svc.UpdateRole(context.Background(), "   ", user.RoleManager)
	require.Error(t, err)
	assert.ErrorIs(t, err, admin.ErrInvalidInput)
}

func TestUserService_UpdateRole_InvalidRole(t *testing.T) {
	t.Parallel()

	repo := new(repomocks.UserRepository)
	svc := admin.NewUserService(repo)

	err := svc.UpdateRole(context.Background(), "u1", user.Role("NOPE"))
	require.Error(t, err)
	assert.ErrorIs(t, err, admin.ErrInvalidInput)
}

func TestUserService_UpdateRole_NotFound(t *testing.T) {
	t.Parallel()

	repo := new(repomocks.UserRepository)
	svc := admin.NewUserService(repo)

	repo.On("UpdateRole", mock.Anything, "u1", user.RoleManager).Return(repository.ErrNotFound).Once()

	err := svc.UpdateRole(context.Background(), "u1", user.RoleManager)
	require.Error(t, err)
	assert.ErrorIs(t, err, repository.ErrNotFound)
	repo.AssertExpectations(t)
}
