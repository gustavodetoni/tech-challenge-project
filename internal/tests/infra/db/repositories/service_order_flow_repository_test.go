package repositories

import (
	"context"
	"testing"
	"time"

	"github.com/jackc/pgconn"
	repositories2 "github.com/soat-architecture/tech-challenge-project/internal/infra/db/repositories"
	"github.com/stretchr/testify/require"

	repository "github.com/soat-architecture/tech-challenge-project/internal/application/port"
	"github.com/soat-architecture/tech-challenge-project/internal/domain/order"
)

func TestServiceOrderFlowRepository_CreateDraft_Conflict(t *testing.T) {
	gdb := newTestGormDB(t, []dbOp{
		{
			kind:         dbOpExec,
			wantContains: []string{`INSERT`, `service_orders`},
			err:          &pgconn.PgError{Code: "23505"},
		},
	})
	repo := repositories2.NewServiceOrderFlowRepository(gdb)

	_, _, err := repo.CreateDraft(context.Background(), repository.CreateServiceOrderDraftParams{
		ServiceOrder: order.ServiceOrder{ID: "so1", Code: "C-1", ClientID: "c1", VehicleID: "v1", Status: order.StatusReceived, OpenedAt: time.Now().UTC()},
		Budget:       order.Budget{ID: "b1", Version: 1, Status: order.BudgetStatusDraft, TotalAmountCents: 100},
	})
	require.ErrorIs(t, err, repository.ErrConflict)
}

func TestServiceOrderFlowRepository_CreateDraft_Success_WithItems(t *testing.T) {
	gdb := newTestGormDB(t, []dbOp{
		{kind: dbOpExec, wantContains: []string{`INSERT`, `service_orders`}, rowsAffected: 1},
		{kind: dbOpExec, wantContains: []string{`INSERT`, `service_order_status_history`}, rowsAffected: 1},
		{kind: dbOpExec, wantContains: []string{`INSERT`, `budgets`}, rowsAffected: 1},
		{kind: dbOpExec, wantContains: []string{`INSERT`, `budget_services`}, rowsAffected: 1},
		{kind: dbOpExec, wantContains: []string{`INSERT`, `budget_parts`}, rowsAffected: 1},
	})
	repo := repositories2.NewServiceOrderFlowRepository(gdb)

	so, b, err := repo.CreateDraft(context.Background(), repository.CreateServiceOrderDraftParams{
		ServiceOrder: order.ServiceOrder{ID: "so1", Code: "C-1", ClientID: "c1", VehicleID: "v1", Status: order.StatusReceived, OpenedAt: time.Now().UTC()},
		Budget:       order.Budget{ID: "b1", Version: 1, Status: order.BudgetStatusDraft, TotalAmountCents: 100},
		BudgetServices: []order.BudgetServiceItem{{
			ServiceID:       "s1",
			Description:     "Alinhamento",
			Quantity:        1,
			UnitPriceCents:  100,
			TotalPriceCents: 100,
		}},
		BudgetParts: []order.BudgetPartItem{{
			PartID:          "p1",
			Description:     "Filtro",
			Quantity:        1,
			UnitPriceCents:  10,
			TotalPriceCents: 10,
		}},
	})
	require.NoError(t, err)
	require.Equal(t, "so1", so.ID)
	require.Equal(t, "b1", b.ID)
}

func TestServiceOrderFlowRepository_CreateBudgetRevision_NotFoundAndConflict(t *testing.T) {
	gdb := newTestGormDB(t, []dbOp{
		{
			kind:         dbOpQuery,
			wantContains: []string{`FROM "service_orders"`},
			columns:      []string{"id", "status"},
			rows:         [][]any{},
		},
		{
			kind:         dbOpQuery,
			wantContains: []string{`FROM "service_orders"`},
			columns:      []string{"id", "status"},
			rows:         [][]any{{"so1", "RECEIVED"}},
		},
	})
	repo := repositories2.NewServiceOrderFlowRepository(gdb)

	_, err := repo.CreateBudgetRevision(context.Background(), repository.CreateBudgetRevisionParams{
		ServiceOrderID: "so1",
		Budget:         order.Budget{ID: "b1", Status: order.BudgetStatusDraft, TotalAmountCents: 100},
	})
	require.ErrorIs(t, err, repository.ErrNotFound)

	_, err = repo.CreateBudgetRevision(context.Background(), repository.CreateBudgetRevisionParams{
		ServiceOrderID: "so1",
		Budget:         order.Budget{ID: "b1", Status: order.BudgetStatusDraft, TotalAmountCents: 100},
	})
	require.ErrorIs(t, err, repository.ErrConflict)
}

func TestServiceOrderFlowRepository_CreateBudgetRevision_Success(t *testing.T) {
	now := time.Now().UTC()
	gdb := newTestGormDB(t, []dbOp{
		{
			kind:         dbOpQuery,
			wantContains: []string{`FROM "service_orders"`},
			columns:      []string{"id", "status"},
			rows:         [][]any{{"so1", "IN_DIAGNOSIS"}},
		},
		{
			kind:         dbOpQuery,
			wantContains: []string{`FROM "budgets"`},
			columns:      []string{"id", "service_order_id", "version", "status", "total_amount_cents", "created_at", "updated_at"},
			rows:         [][]any{{"b0", "so1", int64(1), "DRAFT", int64(100), now, now}},
		},
		{kind: dbOpExec, wantContains: []string{`INSERT`, `budgets`}, rowsAffected: 1},
		{kind: dbOpExec, wantContains: []string{`INSERT`, `budget_services`}, rowsAffected: 1},
	})
	repo := repositories2.NewServiceOrderFlowRepository(gdb)

	out, err := repo.CreateBudgetRevision(context.Background(), repository.CreateBudgetRevisionParams{
		ServiceOrderID: "so1",
		Budget:         order.Budget{ID: "b1", Status: order.BudgetStatusDraft, TotalAmountCents: 100},
		BudgetServices: []order.BudgetServiceItem{{
			ServiceID:       "s1",
			Description:     "Alinhamento",
			Quantity:        1,
			UnitPriceCents:  100,
			TotalPriceCents: 100,
		}},
	})
	require.NoError(t, err)
	require.GreaterOrEqual(t, out.Version, 2)
}

func TestServiceOrderFlowRepository_CreateBudgetRevision_BudgetQueryError(t *testing.T) {
	gdb := newTestGormDB(t, []dbOp{
		{
			kind:         dbOpQuery,
			wantContains: []string{`FROM "service_orders"`},
			columns:      []string{"id", "status"},
			rows:         [][]any{{"so1", "IN_DIAGNOSIS"}},
		},
		{
			kind:         dbOpQuery,
			wantContains: []string{`FROM "budgets"`},
			err:          &pgconn.PgError{Code: "XX000"},
		},
	})
	repo := repositories2.NewServiceOrderFlowRepository(gdb)

	_, err := repo.CreateBudgetRevision(context.Background(), repository.CreateBudgetRevisionParams{
		ServiceOrderID: "so1",
		Budget:         order.Budget{ID: "b1", Status: order.BudgetStatusDraft, TotalAmountCents: 100},
	})
	require.Error(t, err)
}

func TestServiceOrderFlowRepository_SendLatestBudget_BudgetNotDraft(t *testing.T) {
	now := time.Now().UTC()
	gdb := newTestGormDB(t, []dbOp{
		{
			kind:         dbOpQuery,
			wantContains: []string{`FROM "budgets"`},
			columns:      []string{"id", "service_order_id", "version", "status", "total_amount_cents", "created_at", "updated_at"},
			rows:         [][]any{{"b1", "so1", int64(1), "SENT", int64(100), now, now}},
		},
	})
	repo := repositories2.NewServiceOrderFlowRepository(gdb)

	err := repo.SendLatestBudget(context.Background(), "so1", nil)
	require.Error(t, err)
}

func TestServiceOrderFlowRepository_SendLatestBudget_Success(t *testing.T) {
	now := time.Now().UTC()
	gdb := newTestGormDB(t, []dbOp{
		{
			kind:         dbOpQuery,
			wantContains: []string{`FROM "budgets"`},
			columns:      []string{"id", "service_order_id", "version", "status", "total_amount_cents", "created_at", "updated_at"},
			rows:         [][]any{{"b1", "so1", int64(1), "DRAFT", int64(100), now, now}},
		},
		{kind: dbOpExec, wantContains: []string{`UPDATE`, `budgets`}, rowsAffected: 1},
		{
			kind:         dbOpQuery,
			wantContains: []string{`FROM "service_orders"`, `SELECT`, `status`},
			columns:      []string{"status"},
			rows:         [][]any{{"IN_DIAGNOSIS"}},
		},
		{kind: dbOpExec, wantContains: []string{`UPDATE`, `service_orders`}, rowsAffected: 1},
		{kind: dbOpExec, wantContains: []string{`INSERT`, `service_order_status_history`}, rowsAffected: 1},
	})
	repo := repositories2.NewServiceOrderFlowRepository(gdb)

	err := repo.SendLatestBudget(context.Background(), "so1", nil)
	require.NoError(t, err)
}

func TestServiceOrderFlowRepository_SendLatestBudget_NoBudget_NotFound(t *testing.T) {
	gdb := newTestGormDB(t, []dbOp{
		{
			kind:         dbOpQuery,
			wantContains: []string{`FROM "budgets"`},
			columns:      []string{"id"},
			rows:         [][]any{},
		},
	})
	repo := repositories2.NewServiceOrderFlowRepository(gdb)

	err := repo.SendLatestBudget(context.Background(), "so1", nil)
	require.ErrorIs(t, err, repository.ErrNotFound)
}

func TestServiceOrderFlowRepository_ApproveLatestBudgetByCode_NotFound(t *testing.T) {
	gdb := newTestGormDB(t, []dbOp{
		{
			kind:         dbOpQuery,
			wantContains: []string{`FROM service_orders so`, `c.document_number`},
			columns:      []string{"id"},
			rows:         [][]any{},
		},
	})
	repo := repositories2.NewServiceOrderFlowRepository(gdb)

	err := repo.ApproveLatestBudgetByCode(context.Background(), "C-1", "46420082412", nil)
	require.ErrorIs(t, err, repository.ErrNotFound)
}

func TestServiceOrderFlowRepository_ApproveLatestBudgetByCode_BudgetNotFound(t *testing.T) {
	gdb := newTestGormDB(t, []dbOp{
		{
			kind:         dbOpQuery,
			wantContains: []string{`FROM service_orders so`, `c.document_number`},
			columns:      []string{"id"},
			rows:         [][]any{{"so1"}},
		},
		{
			kind:         dbOpQuery,
			wantContains: []string{`FROM "budgets"`},
			columns:      []string{"id"},
			rows:         [][]any{},
		},
	})
	repo := repositories2.NewServiceOrderFlowRepository(gdb)

	err := repo.ApproveLatestBudgetByCode(context.Background(), "C-1", "46420082412", nil)
	require.ErrorIs(t, err, repository.ErrNotFound)
}

func TestServiceOrderFlowRepository_ApproveLatestBudgetByCode_BudgetNotSent(t *testing.T) {
	now := time.Now().UTC()
	gdb := newTestGormDB(t, []dbOp{
		{
			kind:         dbOpQuery,
			wantContains: []string{`FROM service_orders so`, `c.document_number`},
			columns:      []string{"id"},
			rows:         [][]any{{"so1"}},
		},
		{
			kind:         dbOpQuery,
			wantContains: []string{`FROM "budgets"`},
			columns:      []string{"id", "service_order_id", "version", "status", "total_amount_cents", "created_at", "updated_at"},
			rows:         [][]any{{"b1", "so1", int64(1), "DRAFT", int64(100), now, now}},
		},
	})
	repo := repositories2.NewServiceOrderFlowRepository(gdb)

	err := repo.ApproveLatestBudgetByCode(context.Background(), "C-1", "46420082412", nil)
	require.Error(t, err)
}

func TestServiceOrderFlowRepository_ApproveLatestBudgetByCode_Success_NoItems(t *testing.T) {
	now := time.Now().UTC()
	gdb := newTestGormDB(t, []dbOp{
		{
			kind:         dbOpQuery,
			wantContains: []string{`FROM service_orders so`, `c.document_number`},
			columns:      []string{"id"},
			rows:         [][]any{{"so1"}},
		},
		{
			kind:         dbOpQuery,
			wantContains: []string{`FROM "budgets"`},
			columns:      []string{"id", "service_order_id", "version", "status", "total_amount_cents", "created_at", "updated_at"},
			rows:         [][]any{{"b1", "so1", int64(1), "SENT", int64(100), now, now}},
		},
		{kind: dbOpExec, wantContains: []string{`UPDATE`, `budgets`}, rowsAffected: 1},
		{
			kind:         dbOpQuery,
			wantContains: []string{`FROM "budget_services"`},
			columns:      []string{"id"},
			rows:         [][]any{},
		},
		{
			kind:         dbOpQuery,
			wantContains: []string{`FROM "budget_parts"`},
			columns:      []string{"id"},
			rows:         [][]any{},
		},
		{
			kind:         dbOpQuery,
			wantContains: []string{`FROM "service_orders"`, `SELECT`, `status`},
			columns:      []string{"status"},
			rows:         [][]any{{"WAITING_APPROVAL"}},
		},
		{kind: dbOpExec, wantContains: []string{`UPDATE`, `service_orders`}, rowsAffected: 1},
		{kind: dbOpExec, wantContains: []string{`INSERT`, `service_order_status_history`}, rowsAffected: 1},
	})
	repo := repositories2.NewServiceOrderFlowRepository(gdb)

	err := repo.ApproveLatestBudgetByCode(context.Background(), "C-1", "46420082412", nil)
	require.NoError(t, err)
}

func TestServiceOrderFlowRepository_ApproveLatestBudgetByCode_Success_WithServiceAndPart(t *testing.T) {
	now := time.Now().UTC()
	gdb := newTestGormDB(t, []dbOp{
		{
			kind:         dbOpQuery,
			wantContains: []string{`FROM service_orders so`, `c.document_number`},
			columns:      []string{"id"},
			rows:         [][]any{{"so1"}},
		},
		{
			kind:         dbOpQuery,
			wantContains: []string{`FROM "budgets"`},
			columns:      []string{"id", "service_order_id", "version", "status", "total_amount_cents", "created_at", "updated_at"},
			rows:         [][]any{{"b1", "so1", int64(1), "SENT", int64(110), now, now}},
		},
		{kind: dbOpExec, wantContains: []string{`UPDATE`, `budgets`}, rowsAffected: 1},
		{
			kind:         dbOpQuery,
			wantContains: []string{`FROM "budget_services"`},
			columns:      []string{"id", "budget_id", "service_id", "description", "quantity", "unit_price_cents", "total_price_cents", "created_at", "updated_at", "deleted_at"},
			rows: [][]any{
				{"bs1", "b1", repositories2.PtrString("s1"), "Svc", int64(1), int64(100), int64(100), now, now, nil},
			},
		},
		{kind: dbOpExec, wantContains: []string{`INSERT`, `service_order_services`}, rowsAffected: 1},
		{
			kind:         dbOpQuery,
			wantContains: []string{`FROM "budget_parts"`},
			columns:      []string{"id", "budget_id", "part_id", "description", "quantity", "unit_price_cents", "total_price_cents", "created_at", "updated_at", "deleted_at"},
			rows: [][]any{
				{"bp1", "b1", repositories2.PtrString("p1"), "Part", int64(1), int64(10), int64(10), now, now, nil},
			},
		},
		{
			kind:         dbOpQuery,
			wantContains: []string{`FROM "parts"`},
			columns:      []string{"id", "stock_quantity", "updated_at", "deleted_at"},
			rows:         [][]any{{"p1", int64(10), now, nil}},
		},
		{kind: dbOpExec, wantContains: []string{`UPDATE`, `parts`}, rowsAffected: 1},
		{kind: dbOpExec, wantContains: []string{`INSERT`, `stock_movements`}, rowsAffected: 1},
		{kind: dbOpExec, wantContains: []string{`INSERT`, `service_order_parts`}, rowsAffected: 1},
		{
			kind:         dbOpQuery,
			wantContains: []string{`FROM "service_orders"`, `SELECT`, `status`},
			columns:      []string{"status"},
			rows:         [][]any{{"WAITING_APPROVAL"}},
		},
		{kind: dbOpExec, wantContains: []string{`UPDATE`, `service_orders`}, rowsAffected: 1},
		{kind: dbOpExec, wantContains: []string{`INSERT`, `service_order_status_history`}, rowsAffected: 1},
	})
	repo := repositories2.NewServiceOrderFlowRepository(gdb)

	err := repo.ApproveLatestBudgetByCode(context.Background(), "C-1", "46420082412", nil)
	require.NoError(t, err)
}

func TestServiceOrderFlowRepository_ApproveLatestBudgetByCode_BudgetServicesQueryError(t *testing.T) {
	now := time.Now().UTC()
	gdb := newTestGormDB(t, []dbOp{
		{
			kind:         dbOpQuery,
			wantContains: []string{`FROM service_orders so`, `c.document_number`},
			columns:      []string{"id"},
			rows:         [][]any{{"so1"}},
		},
		{
			kind:         dbOpQuery,
			wantContains: []string{`FROM "budgets"`},
			columns:      []string{"id", "service_order_id", "version", "status", "total_amount_cents", "created_at", "updated_at"},
			rows:         [][]any{{"b1", "so1", int64(1), "SENT", int64(100), now, now}},
		},
		{kind: dbOpExec, wantContains: []string{`UPDATE`, `budgets`}, rowsAffected: 1},
		{
			kind:         dbOpQuery,
			wantContains: []string{`FROM "budget_services"`},
			err:          &pgconn.PgError{Code: "XX000"},
		},
	})
	repo := repositories2.NewServiceOrderFlowRepository(gdb)

	err := repo.ApproveLatestBudgetByCode(context.Background(), "C-1", "46420082412", nil)
	require.Error(t, err)
}

func TestServiceOrderFlowRepository_ApproveLatestBudgetByCode_BudgetPartsQueryError(t *testing.T) {
	now := time.Now().UTC()
	gdb := newTestGormDB(t, []dbOp{
		{
			kind:         dbOpQuery,
			wantContains: []string{`FROM service_orders so`, `c.document_number`},
			columns:      []string{"id"},
			rows:         [][]any{{"so1"}},
		},
		{
			kind:         dbOpQuery,
			wantContains: []string{`FROM "budgets"`},
			columns:      []string{"id", "service_order_id", "version", "status", "total_amount_cents", "created_at", "updated_at"},
			rows:         [][]any{{"b1", "so1", int64(1), "SENT", int64(100), now, now}},
		},
		{kind: dbOpExec, wantContains: []string{`UPDATE`, `budgets`}, rowsAffected: 1},
		{
			kind:         dbOpQuery,
			wantContains: []string{`FROM "budget_services"`},
			columns:      []string{"id"},
			rows:         [][]any{},
		},
		{
			kind:         dbOpQuery,
			wantContains: []string{`FROM "budget_parts"`},
			err:          &pgconn.PgError{Code: "XX000"},
		},
	})
	repo := repositories2.NewServiceOrderFlowRepository(gdb)

	err := repo.ApproveLatestBudgetByCode(context.Background(), "C-1", "46420082412", nil)
	require.Error(t, err)
}

func TestServiceOrderFlowRepository_ApproveLatestBudgetByCode_PartIDNil_SkipsAndTransitions(t *testing.T) {
	now := time.Now().UTC()
	gdb := newTestGormDB(t, []dbOp{
		{
			kind:         dbOpQuery,
			wantContains: []string{`FROM service_orders so`, `c.document_number`},
			columns:      []string{"id"},
			rows:         [][]any{{"so1"}},
		},
		{
			kind:         dbOpQuery,
			wantContains: []string{`FROM "budgets"`},
			columns:      []string{"id", "service_order_id", "version", "status", "total_amount_cents", "created_at", "updated_at"},
			rows:         [][]any{{"b1", "so1", int64(1), "SENT", int64(100), now, now}},
		},
		{kind: dbOpExec, wantContains: []string{`UPDATE`, `budgets`}, rowsAffected: 1},
		{
			kind:         dbOpQuery,
			wantContains: []string{`FROM "budget_services"`},
			columns:      []string{"id"},
			rows:         [][]any{},
		},
		{
			kind:         dbOpQuery,
			wantContains: []string{`FROM "budget_parts"`},
			columns:      []string{"id", "budget_id", "part_id", "description", "quantity", "unit_price_cents", "total_price_cents", "created_at", "updated_at", "deleted_at"},
			rows: [][]any{
				{"bp1", "b1", (*string)(nil), "Part", int64(1), int64(10), int64(10), now, now, nil},
			},
		},
		{
			kind:         dbOpQuery,
			wantContains: []string{`FROM "service_orders"`, `SELECT`, `status`},
			columns:      []string{"status"},
			rows:         [][]any{{"WAITING_APPROVAL"}},
		},
		{kind: dbOpExec, wantContains: []string{`UPDATE`, `service_orders`}, rowsAffected: 1},
		{kind: dbOpExec, wantContains: []string{`INSERT`, `service_order_status_history`}, rowsAffected: 1},
	})
	repo := repositories2.NewServiceOrderFlowRepository(gdb)

	err := repo.ApproveLatestBudgetByCode(context.Background(), "C-1", "46420082412", nil)
	require.NoError(t, err)
}

func TestServiceOrderFlowRepository_ApproveLatestBudgetByCode_ServiceInsertError(t *testing.T) {
	now := time.Now().UTC()
	gdb := newTestGormDB(t, []dbOp{
		{
			kind:         dbOpQuery,
			wantContains: []string{`FROM service_orders so`, `c.document_number`},
			columns:      []string{"id"},
			rows:         [][]any{{"so1"}},
		},
		{
			kind:         dbOpQuery,
			wantContains: []string{`FROM "budgets"`},
			columns:      []string{"id", "service_order_id", "version", "status", "total_amount_cents", "created_at", "updated_at"},
			rows:         [][]any{{"b1", "so1", int64(1), "SENT", int64(110), now, now}},
		},
		{kind: dbOpExec, wantContains: []string{`UPDATE`, `budgets`}, rowsAffected: 1},
		{
			kind:         dbOpQuery,
			wantContains: []string{`FROM "budget_services"`},
			columns:      []string{"id", "budget_id", "service_id", "description", "quantity", "unit_price_cents", "total_price_cents", "created_at", "updated_at", "deleted_at"},
			rows: [][]any{
				{"bs1", "b1", repositories2.PtrString("s1"), "Svc", int64(1), int64(100), int64(100), now, now, nil},
			},
		},
		{kind: dbOpExec, wantContains: []string{`INSERT`, `service_order_services`}, err: &pgconn.PgError{Code: "XX000"}},
	})
	repo := repositories2.NewServiceOrderFlowRepository(gdb)

	err := repo.ApproveLatestBudgetByCode(context.Background(), "C-1", "46420082412", nil)
	require.Error(t, err)
}

func TestServiceOrderFlowRepository_ApproveLatestBudgetByCode_InsufficientStock(t *testing.T) {
	now := time.Now().UTC()
	gdb := newTestGormDB(t, []dbOp{
		{
			kind:         dbOpQuery,
			wantContains: []string{`FROM service_orders so`, `c.document_number`},
			columns:      []string{"id"},
			rows:         [][]any{{"so1"}},
		},
		{
			kind:         dbOpQuery,
			wantContains: []string{`FROM "budgets"`},
			columns:      []string{"id", "service_order_id", "version", "status", "total_amount_cents", "created_at", "updated_at"},
			rows:         [][]any{{"b1", "so1", int64(1), "SENT", int64(10), now, now}},
		},
		{kind: dbOpExec, wantContains: []string{`UPDATE`, `budgets`}, rowsAffected: 1},
		{
			kind:         dbOpQuery,
			wantContains: []string{`FROM "budget_services"`},
			columns:      []string{"id"},
			rows:         [][]any{},
		},
		{
			kind:         dbOpQuery,
			wantContains: []string{`FROM "budget_parts"`},
			columns:      []string{"id", "budget_id", "part_id", "description", "quantity", "unit_price_cents", "total_price_cents", "created_at", "updated_at", "deleted_at"},
			rows: [][]any{
				{"bp1", "b1", repositories2.PtrString("p1"), "Part", int64(2), int64(10), int64(20), now, now, nil},
			},
		},
		{
			kind:         dbOpQuery,
			wantContains: []string{`FROM "parts"`},
			columns:      []string{"id", "stock_quantity", "updated_at", "deleted_at"},
			rows:         [][]any{{"p1", int64(1), now, nil}},
		},
	})
	repo := repositories2.NewServiceOrderFlowRepository(gdb)

	err := repo.ApproveLatestBudgetByCode(context.Background(), "C-1", "46420082412", nil)
	require.Error(t, err)
}

func TestServiceOrderFlowRepository_ApproveLatestBudgetByCode_PartNotFound(t *testing.T) {
	now := time.Now().UTC()
	gdb := newTestGormDB(t, []dbOp{
		{
			kind:         dbOpQuery,
			wantContains: []string{`FROM service_orders so`, `c.document_number`},
			columns:      []string{"id"},
			rows:         [][]any{{"so1"}},
		},
		{
			kind:         dbOpQuery,
			wantContains: []string{`FROM "budgets"`},
			columns:      []string{"id", "service_order_id", "version", "status", "total_amount_cents", "created_at", "updated_at"},
			rows:         [][]any{{"b1", "so1", int64(1), "SENT", int64(10), now, now}},
		},
		{kind: dbOpExec, wantContains: []string{`UPDATE`, `budgets`}, rowsAffected: 1},
		{
			kind:         dbOpQuery,
			wantContains: []string{`FROM "budget_services"`},
			columns:      []string{"id"},
			rows:         [][]any{},
		},
		{
			kind:         dbOpQuery,
			wantContains: []string{`FROM "budget_parts"`},
			columns:      []string{"id", "budget_id", "part_id", "description", "quantity", "unit_price_cents", "total_price_cents", "created_at", "updated_at", "deleted_at"},
			rows: [][]any{
				{"bp1", "b1", repositories2.PtrString("p1"), "Part", int64(1), int64(10), int64(10), now, now, nil},
			},
		},
		{
			kind:         dbOpQuery,
			wantContains: []string{`FROM "parts"`},
			columns:      []string{"id", "stock_quantity", "updated_at", "deleted_at"},
			rows:         [][]any{},
		},
	})
	repo := repositories2.NewServiceOrderFlowRepository(gdb)

	err := repo.ApproveLatestBudgetByCode(context.Background(), "C-1", "46420082412", nil)
	require.ErrorIs(t, err, repository.ErrNotFound)
}

func TestServiceOrderFlowRepository_ApproveLatestBudgetByCode_PartUpdateError(t *testing.T) {
	now := time.Now().UTC()
	gdb := newTestGormDB(t, []dbOp{
		{
			kind:         dbOpQuery,
			wantContains: []string{`FROM service_orders so`, `c.document_number`},
			columns:      []string{"id"},
			rows:         [][]any{{"so1"}},
		},
		{
			kind:         dbOpQuery,
			wantContains: []string{`FROM "budgets"`},
			columns:      []string{"id", "service_order_id", "version", "status", "total_amount_cents", "created_at", "updated_at"},
			rows:         [][]any{{"b1", "so1", int64(1), "SENT", int64(10), now, now}},
		},
		{kind: dbOpExec, wantContains: []string{`UPDATE`, `budgets`}, rowsAffected: 1},
		{
			kind:         dbOpQuery,
			wantContains: []string{`FROM "budget_services"`},
			columns:      []string{"id"},
			rows:         [][]any{},
		},
		{
			kind:         dbOpQuery,
			wantContains: []string{`FROM "budget_parts"`},
			columns:      []string{"id", "budget_id", "part_id", "description", "quantity", "unit_price_cents", "total_price_cents", "created_at", "updated_at", "deleted_at"},
			rows: [][]any{
				{"bp1", "b1", repositories2.PtrString("p1"), "Part", int64(1), int64(10), int64(10), now, now, nil},
			},
		},
		{
			kind:         dbOpQuery,
			wantContains: []string{`FROM "parts"`},
			columns:      []string{"id", "stock_quantity", "updated_at", "deleted_at"},
			rows:         [][]any{{"p1", int64(10), now, nil}},
		},
		{kind: dbOpExec, wantContains: []string{`UPDATE`, `parts`}, err: &pgconn.PgError{Code: "XX000"}},
	})
	repo := repositories2.NewServiceOrderFlowRepository(gdb)

	err := repo.ApproveLatestBudgetByCode(context.Background(), "C-1", "46420082412", nil)
	require.Error(t, err)
}

func TestServiceOrderFlowRepository_ApproveLatestBudgetByCode_StockMovementInsertError(t *testing.T) {
	now := time.Now().UTC()
	gdb := newTestGormDB(t, []dbOp{
		{
			kind:         dbOpQuery,
			wantContains: []string{`FROM service_orders so`, `c.document_number`},
			columns:      []string{"id"},
			rows:         [][]any{{"so1"}},
		},
		{
			kind:         dbOpQuery,
			wantContains: []string{`FROM "budgets"`},
			columns:      []string{"id", "service_order_id", "version", "status", "total_amount_cents", "created_at", "updated_at"},
			rows:         [][]any{{"b1", "so1", int64(1), "SENT", int64(10), now, now}},
		},
		{kind: dbOpExec, wantContains: []string{`UPDATE`, `budgets`}, rowsAffected: 1},
		{
			kind:         dbOpQuery,
			wantContains: []string{`FROM "budget_services"`},
			columns:      []string{"id"},
			rows:         [][]any{},
		},
		{
			kind:         dbOpQuery,
			wantContains: []string{`FROM "budget_parts"`},
			columns:      []string{"id", "budget_id", "part_id", "description", "quantity", "unit_price_cents", "total_price_cents", "created_at", "updated_at", "deleted_at"},
			rows: [][]any{
				{"bp1", "b1", repositories2.PtrString("p1"), "Part", int64(1), int64(10), int64(10), now, now, nil},
			},
		},
		{
			kind:         dbOpQuery,
			wantContains: []string{`FROM "parts"`},
			columns:      []string{"id", "stock_quantity", "updated_at", "deleted_at"},
			rows:         [][]any{{"p1", int64(10), now, nil}},
		},
		{kind: dbOpExec, wantContains: []string{`UPDATE`, `parts`}, rowsAffected: 1},
		{kind: dbOpExec, wantContains: []string{`INSERT`, `stock_movements`}, err: &pgconn.PgError{Code: "XX000"}},
	})
	repo := repositories2.NewServiceOrderFlowRepository(gdb)

	err := repo.ApproveLatestBudgetByCode(context.Background(), "C-1", "46420082412", nil)
	require.Error(t, err)
}

func TestServiceOrderFlowRepository_RejectLatestBudgetByCode_Success(t *testing.T) {
	now := time.Now().UTC()
	gdb := newTestGormDB(t, []dbOp{
		{
			kind:         dbOpQuery,
			wantContains: []string{`FROM service_orders so`, `c.document_number`},
			columns:      []string{"id"},
			rows:         [][]any{{"so1"}},
		},
		{
			kind:         dbOpQuery,
			wantContains: []string{`FROM "budgets"`},
			columns:      []string{"id", "service_order_id", "version", "status", "total_amount_cents", "created_at", "updated_at"},
			rows:         [][]any{{"b1", "so1", int64(1), "SENT", int64(100), now, now}},
		},
		{kind: dbOpExec, wantContains: []string{`UPDATE`, `budgets`}, rowsAffected: 1},
		{
			kind:         dbOpQuery,
			wantContains: []string{`FROM "service_orders"`, `SELECT`, `status`},
			columns:      []string{"status"},
			rows:         [][]any{{"WAITING_APPROVAL"}},
		},
		{kind: dbOpExec, wantContains: []string{`UPDATE`, `service_orders`}, rowsAffected: 1},
		{kind: dbOpExec, wantContains: []string{`INSERT`, `service_order_status_history`}, rowsAffected: 1},
	})
	repo := repositories2.NewServiceOrderFlowRepository(gdb)

	err := repo.RejectLatestBudgetByCode(context.Background(), "C-1", "46420082412", "too expensive")
	require.NoError(t, err)
}

func TestServiceOrderFlowRepository_RejectLatestBudgetByCode_NotFound(t *testing.T) {
	gdb := newTestGormDB(t, []dbOp{
		{
			kind:         dbOpQuery,
			wantContains: []string{`FROM service_orders so`, `c.document_number`},
			columns:      []string{"id"},
			rows:         [][]any{},
		},
	})
	repo := repositories2.NewServiceOrderFlowRepository(gdb)

	err := repo.RejectLatestBudgetByCode(context.Background(), "C-1", "46420082412", "too expensive")
	require.ErrorIs(t, err, repository.ErrNotFound)
}

func TestServiceOrderFlowRepository_RejectLatestBudgetByCode_BudgetNotFound(t *testing.T) {
	gdb := newTestGormDB(t, []dbOp{
		{
			kind:         dbOpQuery,
			wantContains: []string{`FROM service_orders so`, `c.document_number`},
			columns:      []string{"id"},
			rows:         [][]any{{"so1"}},
		},
		{
			kind:         dbOpQuery,
			wantContains: []string{`FROM "budgets"`},
			columns:      []string{"id"},
			rows:         [][]any{},
		},
	})
	repo := repositories2.NewServiceOrderFlowRepository(gdb)

	err := repo.RejectLatestBudgetByCode(context.Background(), "C-1", "46420082412", "too expensive")
	require.ErrorIs(t, err, repository.ErrNotFound)
}

func TestServiceOrderFlowRepository_RejectLatestBudgetByCode_BudgetNotSent(t *testing.T) {
	now := time.Now().UTC()
	gdb := newTestGormDB(t, []dbOp{
		{
			kind:         dbOpQuery,
			wantContains: []string{`FROM service_orders so`, `c.document_number`},
			columns:      []string{"id"},
			rows:         [][]any{{"so1"}},
		},
		{
			kind:         dbOpQuery,
			wantContains: []string{`FROM "budgets"`},
			columns:      []string{"id", "service_order_id", "version", "status", "total_amount_cents", "created_at", "updated_at"},
			rows:         [][]any{{"b1", "so1", int64(1), "DRAFT", int64(100), now, now}},
		},
	})
	repo := repositories2.NewServiceOrderFlowRepository(gdb)

	err := repo.RejectLatestBudgetByCode(context.Background(), "C-1", "46420082412", "too expensive")
	require.Error(t, err)
}

func TestServiceOrderFlowRepository_RejectLatestBudgetByCode_BudgetUpdateError(t *testing.T) {
	now := time.Now().UTC()
	gdb := newTestGormDB(t, []dbOp{
		{
			kind:         dbOpQuery,
			wantContains: []string{`FROM service_orders so`, `c.document_number`},
			columns:      []string{"id"},
			rows:         [][]any{{"so1"}},
		},
		{
			kind:         dbOpQuery,
			wantContains: []string{`FROM "budgets"`},
			columns:      []string{"id", "service_order_id", "version", "status", "total_amount_cents", "created_at", "updated_at"},
			rows:         [][]any{{"b1", "so1", int64(1), "SENT", int64(100), now, now}},
		},
		{kind: dbOpExec, wantContains: []string{`UPDATE`, `budgets`}, err: &pgconn.PgError{Code: "XX000"}},
	})
	repo := repositories2.NewServiceOrderFlowRepository(gdb)

	err := repo.RejectLatestBudgetByCode(context.Background(), "C-1", "46420082412", "too expensive")
	require.Error(t, err)
}

func TestServiceOrderFlowRepository_RejectLatestBudgetByCode_InvalidTransition(t *testing.T) {
	now := time.Now().UTC()
	gdb := newTestGormDB(t, []dbOp{
		{
			kind:         dbOpQuery,
			wantContains: []string{`FROM service_orders so`, `c.document_number`},
			columns:      []string{"id"},
			rows:         [][]any{{"so1"}},
		},
		{
			kind:         dbOpQuery,
			wantContains: []string{`FROM "budgets"`},
			columns:      []string{"id", "service_order_id", "version", "status", "total_amount_cents", "created_at", "updated_at"},
			rows:         [][]any{{"b1", "so1", int64(1), "SENT", int64(100), now, now}},
		},
		{kind: dbOpExec, wantContains: []string{`UPDATE`, `budgets`}, rowsAffected: 1},
		{
			kind:         dbOpQuery,
			wantContains: []string{`FROM "service_orders"`, `SELECT`, `status`},
			columns:      []string{"status"},
			rows:         [][]any{{"RECEIVED"}},
		},
	})
	repo := repositories2.NewServiceOrderFlowRepository(gdb)

	err := repo.RejectLatestBudgetByCode(context.Background(), "C-1", "46420082412", "too expensive")
	require.Error(t, err)
}

func TestServiceOrderFlowRepository_GetClientViewByCode_NotFound(t *testing.T) {
	gdb := newTestGormDB(t, []dbOp{
		{
			kind:         dbOpQuery,
			wantContains: []string{`FROM service_orders so`, `join clients`, `join vehicles`},
			columns:      []string{"id"},
			rows:         [][]any{},
		},
	})
	repo := repositories2.NewServiceOrderFlowRepository(gdb)

	_, err := repo.GetClientViewByCode(context.Background(), "C-1", "46420082412")
	require.ErrorIs(t, err, repository.ErrNotFound)
}

func TestServiceOrderFlowRepository_GetClientViewByCode_Success(t *testing.T) {
	openedAt := time.Date(2026, 4, 1, 10, 0, 0, 0, time.UTC)
	gdb := newTestGormDB(t, []dbOp{
		{
			kind:         dbOpQuery,
			wantContains: []string{`FROM service_orders so`, `join clients`, `join vehicles`},
			columns:      []string{"id", "code", "status", "client_id", "vehicle_id", "opened_at", "customer_complaint", "plate", "brand", "model", "manufacture_year", "model_year", "color"},
			rows: [][]any{
				{"so1", "C-1", "RECEIVED", "c1", "v1", openedAt, nil, "ABC1D23", "Fiat", "Uno", nil, int64(2015), nil},
			},
		},
		{
			kind:         dbOpQuery,
			wantContains: []string{`FROM "budgets"`},
			columns:      []string{"id", "service_order_id", "version", "status", "total_amount_cents", "sent_at", "approved_at", "rejected_at", "approved_by_name", "rejection_reason"},
			rows:         [][]any{{"b1", "so1", int64(1), "DRAFT", int64(110), nil, nil, nil, nil, nil}},
		},
		{
			kind:         dbOpQuery,
			wantContains: []string{`FROM "budget_services"`},
			columns:      []string{"service_id", "description", "quantity", "unit_price_cents", "total_price_cents"},
			rows:         [][]any{},
		},
		{
			kind:         dbOpQuery,
			wantContains: []string{`FROM "budget_parts"`},
			columns:      []string{"part_id", "description", "quantity", "unit_price_cents", "total_price_cents"},
			rows:         [][]any{},
		},
		{
			kind:         dbOpQuery,
			wantContains: []string{`FROM "service_order_status_history"`},
			columns:      []string{"from_status", "to_status", "changed_at", "changed_by_user_id", "reason"},
			rows:         [][]any{},
		},
	})
	repo := repositories2.NewServiceOrderFlowRepository(gdb)

	view, err := repo.GetClientViewByCode(context.Background(), "C-1", "46420082412")
	require.NoError(t, err)
	require.Equal(t, "C-1", view.Code)
	require.Equal(t, order.BudgetStatusDraft, view.BudgetStatus)
}

func TestServiceOrderFlowRepository_StartDiagnosis_NotFound(t *testing.T) {
	gdb := newTestGormDB(t, []dbOp{
		{
			kind:         dbOpQuery,
			wantContains: []string{`FROM "service_orders"`},
			columns:      []string{"status"},
			rows:         [][]any{},
		},
	})
	repo := repositories2.NewServiceOrderFlowRepository(gdb)

	err := repo.StartDiagnosis(context.Background(), "so1", nil)
	require.ErrorIs(t, err, repository.ErrNotFound)
}

func TestServiceOrderFlowRepository_StartDiagnosis_Success(t *testing.T) {
	gdb := newTestGormDB(t, []dbOp{
		{
			kind:         dbOpQuery,
			wantContains: []string{`FROM "service_orders"`, `SELECT`, `status`},
			columns:      []string{"status"},
			rows:         [][]any{{"RECEIVED"}},
		},
		{
			kind:         dbOpExec,
			wantContains: []string{`UPDATE`, `service_orders`},
			rowsAffected: 1,
		},
		{
			kind:         dbOpExec,
			wantContains: []string{`INSERT`, `service_order_status_history`},
			rowsAffected: 1,
		},
	})
	repo := repositories2.NewServiceOrderFlowRepository(gdb)

	err := repo.StartDiagnosis(context.Background(), "so1", nil)
	require.NoError(t, err)
}

func TestServiceOrderFlowRepository_Finish_SameStatus_NoOp(t *testing.T) {
	gdb := newTestGormDB(t, []dbOp{
		{
			kind:         dbOpQuery,
			wantContains: []string{`FROM "service_orders"`, `SELECT`, `status`},
			columns:      []string{"status"},
			rows:         [][]any{{"FINISHED"}},
		},
	})
	repo := repositories2.NewServiceOrderFlowRepository(gdb)

	err := repo.Finish(context.Background(), "so1", nil)
	require.NoError(t, err)
}

func TestServiceOrderFlowRepository_Deliver_InvalidTransition(t *testing.T) {
	gdb := newTestGormDB(t, []dbOp{
		{
			kind:         dbOpQuery,
			wantContains: []string{`FROM "service_orders"`, `SELECT`, `status`},
			columns:      []string{"status"},
			rows:         [][]any{{"RECEIVED"}},
		},
	})
	repo := repositories2.NewServiceOrderFlowRepository(gdb)

	err := repo.Deliver(context.Background(), "so1", nil)
	require.Error(t, err)
}

func TestIsAllowedTransition(t *testing.T) {
	require.True(t, repositories2.IsAllowedTransition(order.StatusReceived, order.StatusInDiagnosis))
	require.True(t, repositories2.IsAllowedTransition(order.StatusReceived, order.StatusWaitingApproval))

	require.True(t, repositories2.IsAllowedTransition(order.StatusInDiagnosis, order.StatusWaitingApproval))
	require.False(t, repositories2.IsAllowedTransition(order.StatusInDiagnosis, order.StatusInProgress))

	require.True(t, repositories2.IsAllowedTransition(order.StatusWaitingApproval, order.StatusInProgress))
	require.True(t, repositories2.IsAllowedTransition(order.StatusWaitingApproval, order.StatusInDiagnosis))
	require.False(t, repositories2.IsAllowedTransition(order.StatusWaitingApproval, order.StatusFinished))

	require.True(t, repositories2.IsAllowedTransition(order.StatusInProgress, order.StatusFinished))
	require.True(t, repositories2.IsAllowedTransition(order.StatusFinished, order.StatusDelivered))

	require.False(t, repositories2.IsAllowedTransition(order.StatusCanceled, order.StatusDelivered))
}
