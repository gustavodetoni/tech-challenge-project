package repositories

import (
	"context"
	"testing"
	"time"

	"github.com/jackc/pgconn"
	repositories2 "github.com/soat-architecture/tech-challenge-project/internal/infra/db/repositories"
	"github.com/stretchr/testify/require"

	repository "github.com/soat-architecture/tech-challenge-project/internal/application/port"
	"github.com/soat-architecture/tech-challenge-project/internal/domain/service"
)

func TestServiceRepository_Create_NilService(t *testing.T) {
	repo := repositories2.NewServiceRepository(nil)
	err := repo.Create(context.Background(), nil)
	require.Error(t, err)
}

func TestServiceRepository_Create_Conflict(t *testing.T) {
	gdb := newTestGormDB(t, []dbOp{
		{
			kind:         dbOpExec,
			wantContains: []string{`INSERT`, `services`},
			err:          &pgconn.PgError{Code: "23505"},
		},
	})
	repo := repositories2.NewServiceRepository(gdb)

	now := time.Now().UTC()
	err := repo.Create(context.Background(), &service.Service{
		ID:               "s1",
		Name:             "Alinhamento",
		BasePriceCents:   100,
		EstimatedMinutes: 10,
		Active:           true,
		CreatedAt:        now,
		UpdatedAt:        now,
	})
	require.ErrorIs(t, err, repository.ErrConflict)
}

func TestServiceRepository_FindByIDs_Empty(t *testing.T) {
	repo := repositories2.NewServiceRepository(nil)
	out, err := repo.FindByIDs(context.Background(), nil)
	require.NoError(t, err)
	require.Empty(t, out)
}

func TestServiceRepository_Update_Delete_Find_List(t *testing.T) {
	now := time.Now().UTC()
	gdb := newTestGormDB(t, []dbOp{
		{kind: dbOpExec, wantContains: []string{`UPDATE`, `services`}, rowsAffected: 0}, // Update -> not found
		{kind: dbOpExec, wantContains: []string{`UPDATE`, `services`}, rowsAffected: 1}, // Update -> ok
		{kind: dbOpExec, wantContains: []string{`UPDATE`, `services`}, rowsAffected: 0}, // Delete -> not found
		{kind: dbOpExec, wantContains: []string{`UPDATE`, `services`}, rowsAffected: 1}, // Delete -> ok
		{
			kind:         dbOpQuery,
			wantContains: []string{`FROM "services"`},
			columns:      []string{"id", "name", "description", "base_price_cents", "estimated_minutes", "active", "created_at", "updated_at"},
			rows:         [][]any{{"s1", "Alinhamento", nil, int64(100), int64(10), true, now, now}},
		}, // FindByID -> ok
		{
			kind:         dbOpQuery,
			wantContains: []string{`FROM "services"`},
			columns:      []string{"id"},
			rows:         [][]any{},
		}, // List -> empty
	})
	repo := repositories2.NewServiceRepository(gdb)

	err := repo.Update(context.Background(), &service.Service{ID: "s1", Name: "Alinhamento", BasePriceCents: 100, EstimatedMinutes: 10, Active: true, UpdatedAt: now})
	require.ErrorIs(t, err, repository.ErrNotFound)

	err = repo.Update(context.Background(), &service.Service{ID: "s1", Name: "Alinhamento", BasePriceCents: 100, EstimatedMinutes: 10, Active: true, UpdatedAt: now})
	require.NoError(t, err)

	err = repo.Delete(context.Background(), "s1")
	require.ErrorIs(t, err, repository.ErrNotFound)

	err = repo.Delete(context.Background(), "s1")
	require.NoError(t, err)

	got, err := repo.FindByID(context.Background(), "s1")
	require.NoError(t, err)
	require.Equal(t, "s1", got.ID)

	list, err := repo.List(context.Background(), 0, -1)
	require.NoError(t, err)
	require.Empty(t, list)
}

func TestServiceRepository_FindByID_NotFound(t *testing.T) {
	gdb := newTestGormDB(t, []dbOp{
		{
			kind:         dbOpQuery,
			wantContains: []string{`FROM "services"`},
			columns:      []string{"id"},
			rows:         [][]any{},
		},
	})
	repo := repositories2.NewServiceRepository(gdb)

	_, err := repo.FindByID(context.Background(), "s1")
	require.ErrorIs(t, err, repository.ErrNotFound)
}

func TestServiceRepository_FindByIDs_And_List_Success(t *testing.T) {
	now := time.Now().UTC()
	gdb := newTestGormDB(t, []dbOp{
		{
			kind:         dbOpQuery,
			wantContains: []string{`FROM "services"`},
			columns:      []string{"id", "name", "base_price_cents", "estimated_minutes", "active", "created_at", "updated_at"},
			rows: [][]any{
				{"s1", "S1", int64(100), int64(10), true, now, now},
				{"s2", "S2", int64(200), int64(20), false, now, now},
			},
		},
		{
			kind:         dbOpQuery,
			wantContains: []string{`FROM "services"`},
			columns:      []string{"id", "name", "base_price_cents", "estimated_minutes", "active", "created_at", "updated_at"},
			rows: [][]any{
				{"s1", "S1", int64(100), int64(10), true, now, now},
			},
		},
	})
	repo := repositories2.NewServiceRepository(gdb)

	got, err := repo.FindByIDs(context.Background(), []string{"s1", "s2"})
	require.NoError(t, err)
	require.Len(t, got, 2)

	list, err := repo.List(context.Background(), 10, 0)
	require.NoError(t, err)
	require.Len(t, list, 1)
}
