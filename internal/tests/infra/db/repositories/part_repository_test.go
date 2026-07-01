package repositories

import (
	"context"
	"testing"
	"time"

	"github.com/jackc/pgconn"
	repositories2 "github.com/soat-architecture/tech-challenge-project/internal/infra/db/repositories"
	"github.com/stretchr/testify/require"

	repository "github.com/soat-architecture/tech-challenge-project/internal/application/port"
	"github.com/soat-architecture/tech-challenge-project/internal/domain/part"
)

func TestPartRepository_AdjustStock_QuantityMustBePositive(t *testing.T) {
	repo := repositories2.NewPartRepository(nil)
	_, err := repo.AdjustStock(context.Background(), "p1", part.StockMovementIn, 0, nil, nil)
	require.Error(t, err)
}

func TestPartRepository_Create_NilPart(t *testing.T) {
	repo := repositories2.NewPartRepository(nil)
	err := repo.Create(context.Background(), nil)
	require.Error(t, err)
}

func TestPartRepository_FindByIDs_Empty(t *testing.T) {
	repo := repositories2.NewPartRepository(nil)
	out, err := repo.FindByIDs(context.Background(), nil)
	require.NoError(t, err)
	require.Empty(t, out)
}

func TestPartRepository_Create_Conflict(t *testing.T) {
	gdb := newTestGormDB(t, []dbOp{
		{
			kind:         dbOpExec,
			wantContains: []string{`INSERT`, `parts`},
			err:          &pgconn.PgError{Code: "23505"},
		},
	})
	repo := repositories2.NewPartRepository(gdb)

	now := time.Now().UTC()
	err := repo.Create(context.Background(), &part.Part{ID: "p1", SKU: "SKU-1", Name: "Filtro", UnitPriceCents: 5000, StockQuantity: 10, Active: true, CreatedAt: now, UpdatedAt: now})
	require.ErrorIs(t, err, repository.ErrConflict)
}

func TestPartRepository_AdjustStock_InvalidMovementType(t *testing.T) {
	now := time.Now().UTC()
	gdb := newTestGormDB(t, []dbOp{
		{
			kind:         dbOpQuery,
			wantContains: []string{`FROM "parts"`},
			columns: []string{
				"id", "sku", "name", "description", "unit_price_cents", "stock_quantity", "active", "created_at", "updated_at", "deleted_at",
			},
			rows: [][]any{
				{"p1", "SKU-1", "Filtro", nil, int64(5000), int64(10), true, now, now, nil},
			},
		},
	})
	repo := repositories2.NewPartRepository(gdb)

	_, err := repo.AdjustStock(context.Background(), "p1", part.StockMovementType("NOPE"), 1, nil, nil)
	require.Error(t, err)
}

func TestPartRepository_AdjustStock_InsufficientStock(t *testing.T) {
	now := time.Now().UTC()
	gdb := newTestGormDB(t, []dbOp{
		{
			kind:         dbOpQuery,
			wantContains: []string{`FROM "parts"`},
			columns: []string{
				"id", "sku", "name", "description", "unit_price_cents", "stock_quantity", "active", "created_at", "updated_at", "deleted_at",
			},
			rows: [][]any{
				{"p1", "SKU-1", "Filtro", nil, int64(5000), int64(2), true, now, now, nil},
			},
		},
	})
	repo := repositories2.NewPartRepository(gdb)

	_, err := repo.AdjustStock(context.Background(), "p1", part.StockMovementOut, 5, nil, nil)
	require.Error(t, err)
}

func TestPartRepository_AdjustStock_NotFound(t *testing.T) {
	gdb := newTestGormDB(t, []dbOp{
		{
			kind:         dbOpQuery,
			wantContains: []string{`FROM "parts"`},
			columns:      []string{"id"},
			rows:         [][]any{},
		},
	})
	repo := repositories2.NewPartRepository(gdb)

	_, err := repo.AdjustStock(context.Background(), "p1", part.StockMovementIn, 1, nil, nil)
	require.ErrorIs(t, err, repository.ErrNotFound)
}

func TestPartRepository_AdjustStock_Success_In(t *testing.T) {
	now := time.Now().UTC()
	gdb := newTestGormDB(t, []dbOp{
		{
			kind:         dbOpQuery,
			wantContains: []string{`FROM "parts"`},
			columns: []string{
				"id", "sku", "name", "description", "unit_price_cents", "stock_quantity", "active", "created_at", "updated_at", "deleted_at",
			},
			rows: [][]any{
				{"p1", "SKU-1", "Filtro", nil, int64(5000), int64(10), true, now, now, nil},
			},
		},
		{
			kind:         dbOpExec,
			wantContains: []string{`UPDATE`, `parts`},
			rowsAffected: 1,
		},
		{
			kind:         dbOpExec,
			wantContains: []string{`INSERT`, `stock_movements`},
			rowsAffected: 1,
		},
	})
	repo := repositories2.NewPartRepository(gdb)

	updated, err := repo.AdjustStock(context.Background(), "p1", part.StockMovementIn, 5, nil, nil)
	require.NoError(t, err)
	require.Equal(t, 15, updated.StockQuantity)
}

func TestPartRepository_AdjustStock_Success_OutAndAdjustment(t *testing.T) {
	now := time.Now().UTC()
	gdb := newTestGormDB(t, []dbOp{
		{
			kind:         dbOpQuery,
			wantContains: []string{`FROM "parts"`},
			columns:      []string{"id", "sku", "name", "unit_price_cents", "stock_quantity", "active", "created_at", "updated_at"},
			rows:         [][]any{{"p1", "SKU-1", "Filtro", int64(5000), int64(10), true, now, now}},
		},
		{kind: dbOpExec, wantContains: []string{`UPDATE`, `parts`}, rowsAffected: 1},
		{kind: dbOpExec, wantContains: []string{`INSERT`, `stock_movements`}, rowsAffected: 1},
		{
			kind:         dbOpQuery,
			wantContains: []string{`FROM "parts"`},
			columns:      []string{"id", "sku", "name", "unit_price_cents", "stock_quantity", "active", "created_at", "updated_at"},
			rows:         [][]any{{"p1", "SKU-1", "Filtro", int64(5000), int64(7), true, now, now}},
		},
		{kind: dbOpExec, wantContains: []string{`UPDATE`, `parts`}, rowsAffected: 1},
		{kind: dbOpExec, wantContains: []string{`INSERT`, `stock_movements`}, rowsAffected: 1},
	})
	repo := repositories2.NewPartRepository(gdb)

	updated, err := repo.AdjustStock(context.Background(), "p1", part.StockMovementOut, 3, nil, nil)
	require.NoError(t, err)
	require.Equal(t, 7, updated.StockQuantity)

	updated, err = repo.AdjustStock(context.Background(), "p1", part.StockMovementAdjustment, 9, nil, nil)
	require.NoError(t, err)
	require.Equal(t, 9, updated.StockQuantity)
}

func TestPartRepository_CRUD_And_List_Normalization(t *testing.T) {
	now := time.Now().UTC()
	gdb := newTestGormDB(t, []dbOp{
		{kind: dbOpExec, wantContains: []string{`UPDATE`, `parts`}, rowsAffected: 0}, // Update -> not found
		{kind: dbOpExec, wantContains: []string{`UPDATE`, `parts`}, rowsAffected: 1}, // Update -> ok
		{kind: dbOpExec, wantContains: []string{`UPDATE`, `parts`}, rowsAffected: 0}, // Delete -> not found
		{kind: dbOpExec, wantContains: []string{`UPDATE`, `parts`}, rowsAffected: 1}, // Delete -> ok
		{
			kind:         dbOpQuery,
			wantContains: []string{`FROM "parts"`},
			columns:      []string{"id"},
			rows:         [][]any{},
		}, // FindByID -> not found
		{
			kind:         dbOpQuery,
			wantContains: []string{`FROM "parts"`},
			columns: []string{
				"id", "sku", "name", "description", "unit_price_cents", "stock_quantity", "active", "created_at", "updated_at", "deleted_at",
			},
			rows: [][]any{
				{"p1", "SKU-1", "Filtro", nil, int64(5000), int64(10), true, now, now, nil},
			},
		}, // FindByIDs -> one
		{
			kind:         dbOpQuery,
			wantContains: []string{`FROM "parts"`},
			columns:      []string{"id"},
			rows:         [][]any{},
		}, // List -> empty
	})
	repo := repositories2.NewPartRepository(gdb)

	err := repo.Update(context.Background(), &part.Part{ID: "p1", SKU: "SKU-1", Name: "Filtro", UnitPriceCents: 5000, Active: true, UpdatedAt: now})
	require.ErrorIs(t, err, repository.ErrNotFound)

	err = repo.Update(context.Background(), &part.Part{ID: "p1", SKU: "SKU-1", Name: "Filtro", UnitPriceCents: 5000, Active: true, UpdatedAt: now})
	require.NoError(t, err)

	err = repo.Delete(context.Background(), "p1")
	require.ErrorIs(t, err, repository.ErrNotFound)

	err = repo.Delete(context.Background(), "p1")
	require.NoError(t, err)

	_, err = repo.FindByID(context.Background(), "p1")
	require.ErrorIs(t, err, repository.ErrNotFound)

	pts, err := repo.FindByIDs(context.Background(), []string{"p1"})
	require.NoError(t, err)
	require.Len(t, pts, 1)

	list, err := repo.List(context.Background(), 0, -1)
	require.NoError(t, err)
	require.Empty(t, list)
}

func TestPartRepository_FindByID_Success(t *testing.T) {
	now := time.Now().UTC()
	gdb := newTestGormDB(t, []dbOp{
		{
			kind:         dbOpQuery,
			wantContains: []string{`FROM "parts"`},
			columns: []string{
				"id", "sku", "name", "description", "unit_price_cents", "stock_quantity", "active", "created_at", "updated_at", "deleted_at",
			},
			rows: [][]any{
				{"p1", "SKU-1", "Filtro", nil, int64(5000), int64(10), true, now, now, nil},
			},
		},
	})
	repo := repositories2.NewPartRepository(gdb)

	out, err := repo.FindByID(context.Background(), "p1")
	require.NoError(t, err)
	require.Equal(t, "p1", out.ID)
}
