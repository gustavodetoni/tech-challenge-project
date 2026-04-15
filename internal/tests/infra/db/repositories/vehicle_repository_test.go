package repositories

import (
	"context"
	"testing"
	"time"

	"github.com/jackc/pgconn"
	repositories2 "github.com/soat-architecture/tech-challenge-project/internal/infra/db/repositories"
	"github.com/stretchr/testify/require"

	"github.com/soat-architecture/tech-challenge-project/internal/domain/vehicle"
	"github.com/soat-architecture/tech-challenge-project/internal/interfaces/repository"
)

func TestVehicleRepository_Create_NilVehicle(t *testing.T) {
	repo := repositories2.NewVehicleRepository(nil)
	err := repo.Create(context.Background(), nil)
	require.Error(t, err)
}

func TestVehicleRepository_Create_Conflict(t *testing.T) {
	gdb := newTestGormDB(t, []dbOp{
		{
			kind:         dbOpExec,
			wantContains: []string{`INSERT`, `vehicles`},
			err:          &pgconn.PgError{Code: "23505"},
		},
	})
	repo := repositories2.NewVehicleRepository(gdb)

	now := time.Now().UTC()
	err := repo.Create(context.Background(), &vehicle.Vehicle{
		ID:        "v1",
		ClientID:  "c1",
		Plate:     "ABC1D23",
		Brand:     "Fiat",
		Model:     "Uno",
		ModelYear: 2015,
		CreatedAt: now,
		UpdatedAt: now,
	})
	require.ErrorIs(t, err, repository.ErrConflict)
}

func TestVehicleRepository_Update_Delete_Find_List(t *testing.T) {
	now := time.Now().UTC()
	gdb := newTestGormDB(t, []dbOp{
		{kind: dbOpExec, wantContains: []string{`UPDATE`, `vehicles`}, rowsAffected: 0}, // Update -> not found
		{kind: dbOpExec, wantContains: []string{`UPDATE`, `vehicles`}, rowsAffected: 1}, // Update -> ok
		{kind: dbOpExec, wantContains: []string{`UPDATE`, `vehicles`}, rowsAffected: 0}, // Delete -> not found
		{kind: dbOpExec, wantContains: []string{`UPDATE`, `vehicles`}, rowsAffected: 1}, // Delete -> ok
		{
			kind:         dbOpQuery,
			wantContains: []string{`FROM "vehicles"`},
			columns:      []string{"id", "client_id", "plate", "brand", "model", "model_year", "created_at", "updated_at"},
			rows:         [][]any{{"v1", "c1", "ABC1D23", "Fiat", "Uno", int64(2015), now, now}},
		}, // FindByID -> ok
		{
			kind:         dbOpQuery,
			wantContains: []string{`FROM "vehicles"`},
			columns:      []string{"id"},
			rows:         [][]any{},
		}, // FindByPlate -> not found
		{
			kind:         dbOpQuery,
			wantContains: []string{`FROM "vehicles"`},
			columns:      []string{"id", "client_id", "plate", "brand", "model", "model_year", "created_at", "updated_at"},
			rows:         [][]any{{"v1", "c1", "ABC1D23", "Fiat", "Uno", int64(2015), now, now}},
		}, // FindByPlate -> ok
		{
			kind:         dbOpQuery,
			wantContains: []string{`FROM "vehicles"`},
			columns:      []string{"id", "client_id", "plate", "brand", "model", "model_year", "created_at", "updated_at"},
			rows:         [][]any{{"v2", "c1", "DEF1G23", "Fiat", "Uno", int64(2016), now, now}},
		}, // ListByClientID -> ok
	})
	repo := repositories2.NewVehicleRepository(gdb)

	err := repo.Update(context.Background(), &vehicle.Vehicle{ID: "v1", ClientID: "c1", Plate: "ABC1D23", Brand: "Fiat", Model: "Uno", ModelYear: 2015, UpdatedAt: now})
	require.ErrorIs(t, err, repository.ErrNotFound)

	err = repo.Update(context.Background(), &vehicle.Vehicle{ID: "v1", ClientID: "c1", Plate: "ABC1D23", Brand: "Fiat", Model: "Uno", ModelYear: 2015, UpdatedAt: now})
	require.NoError(t, err)

	err = repo.Delete(context.Background(), "v1")
	require.ErrorIs(t, err, repository.ErrNotFound)

	err = repo.Delete(context.Background(), "v1")
	require.NoError(t, err)

	v, err := repo.FindByID(context.Background(), "v1")
	require.NoError(t, err)
	require.Equal(t, "v1", v.ID)

	_, err = repo.FindByPlate(context.Background(), "ABC1D23")
	require.ErrorIs(t, err, repository.ErrNotFound)

	v2, err := repo.FindByPlate(context.Background(), "ABC1D23")
	require.NoError(t, err)
	require.Equal(t, "v1", v2.ID)

	list, err := repo.ListByClientID(context.Background(), "c1", 0, -1)
	require.NoError(t, err)
	require.Len(t, list, 1)
}

func TestVehicleRepository_FindByID_NotFound(t *testing.T) {
	gdb := newTestGormDB(t, []dbOp{
		{
			kind:         dbOpQuery,
			wantContains: []string{`FROM "vehicles"`},
			columns:      []string{"id"},
			rows:         [][]any{},
		},
	})
	repo := repositories2.NewVehicleRepository(gdb)

	_, err := repo.FindByID(context.Background(), "v1")
	require.ErrorIs(t, err, repository.ErrNotFound)
}

func TestVehicleRepository_FindByID_QueryError(t *testing.T) {
	gdb := newTestGormDB(t, []dbOp{
		{
			kind:         dbOpQuery,
			wantContains: []string{`FROM "vehicles"`},
			err:          &pgconn.PgError{Code: "XX000"},
		},
	})
	repo := repositories2.NewVehicleRepository(gdb)

	_, err := repo.FindByID(context.Background(), "v1")
	require.Error(t, err)
}
