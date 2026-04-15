package repositories

import (
	"context"
	"testing"
	"time"

	"github.com/jackc/pgconn"
	repositories2 "github.com/soat-architecture/tech-challenge-project/internal/infra/db/repositories"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/soat-architecture/tech-challenge-project/internal/domain/order"
	"github.com/soat-architecture/tech-challenge-project/internal/interfaces/repository"
)

func TestServiceOrderRepository_FindByID_NotFound(t *testing.T) {
	gdb := newTestGormDB(t, []dbOp{
		{
			kind:         dbOpQuery,
			wantContains: []string{`FROM "service_orders"`},
			columns:      []string{"id"},
			rows:         [][]any{},
		},
	})
	repo := repositories2.NewServiceOrderRepository(gdb)

	_, err := repo.FindByID(context.Background(), "so1")
	require.ErrorIs(t, err, repository.ErrNotFound)
}

func TestServiceOrderRepository_FindByID_Success(t *testing.T) {
	openedAt := time.Date(2026, 4, 1, 10, 0, 0, 0, time.UTC)
	gdb := newTestGormDB(t, []dbOp{
		{
			kind:         dbOpQuery,
			wantContains: []string{`FROM "service_orders"`},
			columns:      []string{"id", "code", "client_id", "vehicle_id", "status", "opened_at", "created_at", "updated_at"},
			rows:         [][]any{{"so1", "C-1", "c1", "v1", "RECEIVED", openedAt, openedAt, openedAt}},
		},
	})
	repo := repositories2.NewServiceOrderRepository(gdb)

	so, err := repo.FindByID(context.Background(), "so1")
	require.NoError(t, err)
	require.Equal(t, "C-1", so.Code)
}

func TestServiceOrderRepository_List_WithAndWithoutStatus(t *testing.T) {
	openedAt := time.Date(2026, 4, 1, 10, 0, 0, 0, time.UTC)
	gdb := newTestGormDB(t, []dbOp{
		{
			kind:         dbOpQuery,
			wantContains: []string{`FROM service_orders so`, `join clients`, `join vehicles`},
			columns:      []string{"id", "code", "status", "opened_at", "client_id", "client_name", "vehicle_id", "plate"},
			rows: [][]any{
				{"so1", "C-1", "RECEIVED", openedAt, "c1", "Maria", "v1", "ABC1D23"},
			},
		},
		{
			kind:         dbOpQuery,
			wantContains: []string{`FROM service_orders so`, `so.status =`},
			columns:      []string{"id", "code", "status", "opened_at", "client_id", "client_name", "vehicle_id", "plate"},
			rows: [][]any{
				{"so2", "C-2", "RECEIVED", openedAt, "c2", "Joao", "v2", "DEF1G23"},
			},
		},
	})
	repo := repositories2.NewServiceOrderRepository(gdb)

	items, err := repo.List(context.Background(), 10, 0, nil)
	require.NoError(t, err)
	require.Len(t, items, 1)

	s := order.StatusReceived
	items2, err := repo.List(context.Background(), 10, 0, &s)
	require.NoError(t, err)
	require.Len(t, items2, 1)
}

func TestServiceOrderRepository_List_QueryError(t *testing.T) {
	gdb := newTestGormDB(t, []dbOp{
		{
			kind:         dbOpQuery,
			wantContains: []string{`FROM service_orders so`},
			err:          &pgconn.PgError{Code: "XX000"},
		},
	})
	repo := repositories2.NewServiceOrderRepository(gdb)

	_, err := repo.List(context.Background(), 10, 0, nil)
	require.Error(t, err)
}

func TestServiceOrderRepository_GetDetailByID_NoBudget(t *testing.T) {
	openedAt := time.Date(2026, 4, 1, 10, 0, 0, 0, time.UTC)
	gdb := newTestGormDB(t, []dbOp{
		{
			kind:         dbOpQuery,
			wantContains: []string{`FROM "service_orders"`},
			columns: []string{
				"id", "code", "client_id", "vehicle_id", "assigned_user_id", "status", "opened_at", "execution_started_at", "finished_at", "delivered_at", "created_at", "updated_at", "deleted_at",
			},
			rows: [][]any{
				{"so1", "C-1", "c1", "v1", nil, "RECEIVED", openedAt, nil, nil, nil, openedAt, openedAt, nil},
			},
		},
		{
			kind:         dbOpQuery,
			wantContains: []string{`FROM "budgets"`},
			columns:      []string{"id"},
			rows:         [][]any{},
		},
		{
			kind:         dbOpQuery,
			wantContains: []string{`FROM "service_order_status_history"`},
			columns:      []string{"from_status", "to_status", "changed_by_user_id", "reason", "changed_at", "deleted_at"},
			rows:         [][]any{},
		},
	})
	repo := repositories2.NewServiceOrderRepository(gdb)

	detail, err := repo.GetDetailByID(context.Background(), "so1")
	require.NoError(t, err)
	require.Equal(t, "so1", detail.ServiceOrder.ID)
	require.Nil(t, detail.LatestBudget)
	require.Empty(t, detail.BudgetServices)
	require.Empty(t, detail.BudgetParts)
	require.Empty(t, detail.StatusHistory)
}

func TestServiceOrderRepository_FindByID_QueryError(t *testing.T) {
	gdb := newTestGormDB(t, []dbOp{
		{
			kind:         dbOpQuery,
			wantContains: []string{`FROM "service_orders"`},
			err:          &pgconn.PgError{Code: "XX000"},
		},
	})
	repo := repositories2.NewServiceOrderRepository(gdb)

	_, err := repo.FindByID(context.Background(), "so1")
	require.Error(t, err)
}

func TestServiceOrderRepository_GetDetailByID_BudgetQueryError(t *testing.T) {
	openedAt := time.Date(2026, 4, 1, 10, 0, 0, 0, time.UTC)
	gdb := newTestGormDB(t, []dbOp{
		{
			kind:         dbOpQuery,
			wantContains: []string{`FROM "service_orders"`},
			columns:      []string{"id", "code", "client_id", "vehicle_id", "status", "opened_at", "created_at", "updated_at"},
			rows:         [][]any{{"so1", "C-1", "c1", "v1", "RECEIVED", openedAt, openedAt, openedAt}},
		},
		{
			kind:         dbOpQuery,
			wantContains: []string{`FROM "budgets"`},
			err:          assert.AnError,
		},
	})
	repo := repositories2.NewServiceOrderRepository(gdb)

	_, err := repo.GetDetailByID(context.Background(), "so1")
	require.Error(t, err)
}

func TestServiceOrderRepository_GetDetailByID_BudgetServicesQueryError(t *testing.T) {
	openedAt := time.Date(2026, 4, 1, 10, 0, 0, 0, time.UTC)
	gdb := newTestGormDB(t, []dbOp{
		{
			kind:         dbOpQuery,
			wantContains: []string{`FROM "service_orders"`},
			columns:      []string{"id", "code", "client_id", "vehicle_id", "status", "opened_at", "created_at", "updated_at"},
			rows:         [][]any{{"so1", "C-1", "c1", "v1", "WAITING_APPROVAL", openedAt, openedAt, openedAt}},
		},
		{
			kind:         dbOpQuery,
			wantContains: []string{`FROM "budgets"`},
			columns:      []string{"id", "service_order_id", "version", "status", "total_amount_cents", "created_at", "updated_at"},
			rows:         [][]any{{"b1", "so1", int64(2), "SENT", int64(110), openedAt, openedAt}},
		},
		{
			kind:         dbOpQuery,
			wantContains: []string{`FROM "budget_services"`},
			err:          &pgconn.PgError{Code: "XX000"},
		},
	})
	repo := repositories2.NewServiceOrderRepository(gdb)

	_, err := repo.GetDetailByID(context.Background(), "so1")
	require.Error(t, err)
}

func TestServiceOrderRepository_GetDetailByID_BudgetPartsQueryError(t *testing.T) {
	openedAt := time.Date(2026, 4, 1, 10, 0, 0, 0, time.UTC)
	gdb := newTestGormDB(t, []dbOp{
		{
			kind:         dbOpQuery,
			wantContains: []string{`FROM "service_orders"`},
			columns:      []string{"id", "code", "client_id", "vehicle_id", "status", "opened_at", "created_at", "updated_at"},
			rows:         [][]any{{"so1", "C-1", "c1", "v1", "WAITING_APPROVAL", openedAt, openedAt, openedAt}},
		},
		{
			kind:         dbOpQuery,
			wantContains: []string{`FROM "budgets"`},
			columns:      []string{"id", "service_order_id", "version", "status", "total_amount_cents", "created_at", "updated_at"},
			rows:         [][]any{{"b1", "so1", int64(2), "SENT", int64(110), openedAt, openedAt}},
		},
		{
			kind:         dbOpQuery,
			wantContains: []string{`FROM "budget_services"`},
			columns:      []string{"service_id", "description", "quantity", "unit_price_cents", "total_price_cents", "deleted_at"},
			rows:         [][]any{},
		},
		{
			kind:         dbOpQuery,
			wantContains: []string{`FROM "budget_parts"`},
			err:          &pgconn.PgError{Code: "XX000"},
		},
	})
	repo := repositories2.NewServiceOrderRepository(gdb)

	_, err := repo.GetDetailByID(context.Background(), "so1")
	require.Error(t, err)
}

func TestServiceOrderRepository_GetDetailByID_HistoryQueryError(t *testing.T) {
	openedAt := time.Date(2026, 4, 1, 10, 0, 0, 0, time.UTC)
	gdb := newTestGormDB(t, []dbOp{
		{
			kind:         dbOpQuery,
			wantContains: []string{`FROM "service_orders"`},
			columns:      []string{"id", "code", "client_id", "vehicle_id", "status", "opened_at", "created_at", "updated_at"},
			rows:         [][]any{{"so1", "C-1", "c1", "v1", "WAITING_APPROVAL", openedAt, openedAt, openedAt}},
		},
		{
			kind:         dbOpQuery,
			wantContains: []string{`FROM "budgets"`},
			columns:      []string{"id", "service_order_id", "version", "status", "total_amount_cents", "created_at", "updated_at"},
			rows:         [][]any{{"b1", "so1", int64(2), "SENT", int64(110), openedAt, openedAt}},
		},
		{
			kind:         dbOpQuery,
			wantContains: []string{`FROM "budget_services"`},
			columns:      []string{"service_id", "description", "quantity", "unit_price_cents", "total_price_cents", "deleted_at"},
			rows:         [][]any{},
		},
		{
			kind:         dbOpQuery,
			wantContains: []string{`FROM "budget_parts"`},
			columns:      []string{"part_id", "description", "quantity", "unit_price_cents", "total_price_cents", "deleted_at"},
			rows:         [][]any{},
		},
		{
			kind:         dbOpQuery,
			wantContains: []string{`FROM "service_order_status_history"`},
			err:          &pgconn.PgError{Code: "XX000"},
		},
	})
	repo := repositories2.NewServiceOrderRepository(gdb)

	_, err := repo.GetDetailByID(context.Background(), "so1")
	require.Error(t, err)
}

func TestServiceOrderRepository_GetDetailByID_WithBudgetAndHistory(t *testing.T) {
	openedAt := time.Date(2026, 4, 1, 10, 0, 0, 0, time.UTC)
	sentAt := time.Date(2026, 4, 2, 10, 0, 0, 0, time.UTC)
	changedAt := time.Date(2026, 4, 3, 10, 0, 0, 0, time.UTC)

	gdb := newTestGormDB(t, []dbOp{
		{
			kind:         dbOpQuery,
			wantContains: []string{`FROM "service_orders"`},
			columns:      []string{"id", "code", "client_id", "vehicle_id", "status", "opened_at", "created_at", "updated_at"},
			rows:         [][]any{{"so1", "C-1", "c1", "v1", "WAITING_APPROVAL", openedAt, openedAt, openedAt}},
		},
		{
			kind:         dbOpQuery,
			wantContains: []string{`FROM "budgets"`},
			columns:      []string{"id", "service_order_id", "version", "status", "total_amount_cents", "sent_at", "created_at", "updated_at"},
			rows:         [][]any{{"b1", "so1", int64(2), "SENT", int64(110), sentAt, openedAt, openedAt}},
		},
		{
			kind:         dbOpQuery,
			wantContains: []string{`FROM "budget_services"`},
			columns:      []string{"service_id", "description", "quantity", "unit_price_cents", "total_price_cents", "deleted_at"},
			rows: [][]any{
				{nil, "Alinhamento", int64(1), int64(100), int64(100), nil},
			},
		},
		{
			kind:         dbOpQuery,
			wantContains: []string{`FROM "budget_parts"`},
			columns:      []string{"part_id", "description", "quantity", "unit_price_cents", "total_price_cents", "deleted_at"},
			rows: [][]any{
				{repositories2.PtrString("p1"), "Filtro", int64(1), int64(10), int64(10), nil},
			},
		},
		{
			kind:         dbOpQuery,
			wantContains: []string{`FROM "service_order_status_history"`},
			columns:      []string{"from_status", "to_status", "changed_by_user_id", "reason", "changed_at", "deleted_at"},
			rows: [][]any{
				{nil, "WAITING_APPROVAL", repositories2.PtrString("u1"), nil, changedAt, nil},
			},
		},
	})
	repo := repositories2.NewServiceOrderRepository(gdb)

	detail, err := repo.GetDetailByID(context.Background(), "so1")
	require.NoError(t, err)
	require.NotNil(t, detail.LatestBudget)
	require.Len(t, detail.BudgetServices, 1)
	require.Len(t, detail.BudgetParts, 1)
	require.Len(t, detail.StatusHistory, 1)
}

func TestServiceOrderRepository_AverageExecutionMinutes_WithFromTo(t *testing.T) {
	from := time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2026, 4, 10, 0, 0, 0, 0, time.UTC)

	gdb := newTestGormDB(t, []dbOp{
		{
			kind:         dbOpQuery,
			wantContains: []string{`avg_minutes`},
			columns:      []string{"avg_minutes"},
			rows: [][]any{
				{12.5},
			},
		},
	})
	repo := repositories2.NewServiceOrderRepository(gdb)

	avg, err := repo.AverageExecutionMinutes(context.Background(), &from, &to)
	require.NoError(t, err)
	require.Equal(t, 12.5, avg)
}

func TestServiceOrderRepository_AverageExecutionMinutes_QueryError(t *testing.T) {
	gdb := newTestGormDB(t, []dbOp{
		{
			kind:         dbOpQuery,
			wantContains: []string{`avg_minutes`},
			err:          &pgconn.PgError{Code: "XX000"},
		},
	})
	repo := repositories2.NewServiceOrderRepository(gdb)

	_, err := repo.AverageExecutionMinutes(context.Background(), nil, nil)
	require.Error(t, err)
}
