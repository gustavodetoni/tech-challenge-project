package repositories

import (
	"context"
	"testing"
	"time"

	"github.com/jackc/pgconn"
	repositories2 "github.com/soat-architecture/tech-challenge-project/internal/infra/db/repositories"
	"github.com/stretchr/testify/require"

	repository "github.com/soat-architecture/tech-challenge-project/internal/application/port"
	"github.com/soat-architecture/tech-challenge-project/internal/domain/user"
)

func TestUserRepository_Create_NilUser(t *testing.T) {
	repo := repositories2.NewUserRepository(nil)
	err := repo.Create(context.Background(), nil)
	require.Error(t, err)
}

func TestUserRepository_Create_Conflict(t *testing.T) {
	gdb := newTestGormDB(t, []dbOp{
		{
			kind:         dbOpExec,
			wantContains: []string{`INSERT`, `users`},
			err:          &pgconn.PgError{Code: "23505"},
		},
	})
	repo := repositories2.NewUserRepository(gdb)

	now := time.Now().UTC()
	err := repo.Create(context.Background(), &user.User{
		ID:           "u1",
		Name:         "Maria",
		Email:        "maria@example.com",
		PasswordHash: "hash",
		Role:         user.RoleViewer,
		CreatedAt:    now,
		UpdatedAt:    now,
	})
	require.ErrorIs(t, err, repository.ErrConflict)
}

func TestUserRepository_FindByEmail_FindByID_UpdateRole(t *testing.T) {
	now := time.Now().UTC()
	gdb := newTestGormDB(t, []dbOp{
		{
			kind:         dbOpQuery,
			wantContains: []string{`FROM "users"`, `lower(email)`},
			columns:      []string{"id", "name", "email", "password_hash", "role", "created_at", "updated_at"},
			rows:         [][]any{{"u1", "Maria", "maria@example.com", "hash", "VIEWER", now, now}},
		},
		{
			kind:         dbOpQuery,
			wantContains: []string{`FROM "users"`},
			columns:      []string{"id"},
			rows:         [][]any{},
		},
		{
			kind:         dbOpExec,
			wantContains: []string{`UPDATE`, `users`},
			rowsAffected: 0,
		},
		{
			kind:         dbOpExec,
			wantContains: []string{`UPDATE`, `users`},
			rowsAffected: 1,
		},
	})
	repo := repositories2.NewUserRepository(gdb)

	u, err := repo.FindByEmail(context.Background(), "MARIA@EXAMPLE.COM")
	require.NoError(t, err)
	require.Equal(t, "u1", u.ID)

	_, err = repo.FindByID(context.Background(), "u2")
	require.ErrorIs(t, err, repository.ErrNotFound)

	err = repo.UpdateRole(context.Background(), "u2", user.RoleManager)
	require.ErrorIs(t, err, repository.ErrNotFound)

	err = repo.UpdateRole(context.Background(), "u1", user.RoleManager)
	require.NoError(t, err)
}

func TestUserRepository_FindByEmail_NotFound_And_FindByID_Success(t *testing.T) {
	now := time.Now().UTC()
	gdb := newTestGormDB(t, []dbOp{
		{
			kind:         dbOpQuery,
			wantContains: []string{`FROM "users"`, `lower(email)`},
			columns:      []string{"id"},
			rows:         [][]any{},
		},
		{
			kind:         dbOpQuery,
			wantContains: []string{`FROM "users"`},
			columns:      []string{"id", "name", "email", "password_hash", "role", "created_at", "updated_at"},
			rows:         [][]any{{"u1", "Maria", "maria@example.com", "hash", "ADMIN", now, now}},
		},
	})
	repo := repositories2.NewUserRepository(gdb)

	_, err := repo.FindByEmail(context.Background(), "maria@example.com")
	require.ErrorIs(t, err, repository.ErrNotFound)

	u, err := repo.FindByID(context.Background(), "u1")
	require.NoError(t, err)
	require.Equal(t, user.RoleAdmin, u.Role)
}
