package repositories

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/soat-architecture/tech-challenge-project/internal/domain/client"
	"github.com/soat-architecture/tech-challenge-project/internal/domain/order"
	"github.com/soat-architecture/tech-challenge-project/internal/domain/part"
	"github.com/soat-architecture/tech-challenge-project/internal/domain/service"
	"github.com/soat-architecture/tech-challenge-project/internal/domain/user"
	"github.com/soat-architecture/tech-challenge-project/internal/domain/vehicle"
	"github.com/soat-architecture/tech-challenge-project/internal/interfaces/repository"
)

func TestStore_NewStore_InitializesMaps(t *testing.T) {
	t.Parallel()

	s := NewStore()
	require.NotNil(t, s)
	assert.NotNil(t, s.usersByID)
	assert.NotNil(t, s.clientsByID)
	assert.NotNil(t, s.vehiclesByID)
	assert.NotNil(t, s.servicesByID)
	assert.NotNil(t, s.partsByID)
	assert.NotNil(t, s.serviceOrdersByID)
	assert.NotNil(t, s.budgetsByID)
	assert.NotNil(t, s.statusHistoryBySO)
}

func TestUserRepository_CRUDAndUniqueness(t *testing.T) {
	t.Parallel()

	s := NewStore()
	repo := NewUserRepository(s)

	err := repo.Create(context.Background(), nil)
	require.Error(t, err)

	err = repo.Create(context.Background(), &user.User{ID: "u1", Email: "   "})
	require.Error(t, err)

	now := time.Now().UTC()
	u := &user.User{ID: "u1", Email: "  ADMIN@EXAMPLE.COM  ", Name: "Admin", Role: user.RoleAdmin, CreatedAt: now, UpdatedAt: now}
	require.NoError(t, repo.Create(context.Background(), u))

	err = repo.Create(context.Background(), &user.User{ID: "u2", Email: "admin@example.com"})
	require.ErrorIs(t, err, repository.ErrConflict)

	gotByEmail, err := repo.FindByEmail(context.Background(), " admin@example.com ")
	require.NoError(t, err)
	assert.Equal(t, "u1", gotByEmail.ID)

	gotByID, err := repo.FindByID(context.Background(), "u1")
	require.NoError(t, err)
	assert.Equal(t, "  ADMIN@EXAMPLE.COM  ", gotByID.Email)

	_, err = repo.FindByID(context.Background(), "missing")
	require.ErrorIs(t, err, repository.ErrNotFound)

	err = repo.UpdateRole(context.Background(), "missing", user.RoleViewer)
	require.ErrorIs(t, err, repository.ErrNotFound)

	require.NoError(t, repo.UpdateRole(context.Background(), "u1", user.RoleViewer))
	gotByID, err = repo.FindByID(context.Background(), "u1")
	require.NoError(t, err)
	assert.Equal(t, user.RoleViewer, gotByID.Role)
	assert.False(t, gotByID.UpdatedAt.Before(now))
}

func TestClientRepository_CRUD_List_FindByDocument(t *testing.T) {
	t.Parallel()

	s := NewStore()
	repo := NewClientRepository(s)

	err := repo.Create(context.Background(), nil)
	require.Error(t, err)

	err = repo.Create(context.Background(), &client.Client{ID: "c1", DocumentNumber: "   "})
	require.Error(t, err)

	t0 := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
	t1 := time.Date(2026, 1, 2, 12, 0, 0, 0, time.UTC)
	c1 := &client.Client{ID: "c1", DocumentType: client.DocumentTypeCPF, DocumentNumber: "111", Name: "A", CreatedAt: t0, UpdatedAt: t0}
	c2 := &client.Client{ID: "c2", DocumentType: client.DocumentTypeCPF, DocumentNumber: "222", Name: "B", CreatedAt: t1, UpdatedAt: t1}
	require.NoError(t, repo.Create(context.Background(), c1))
	require.NoError(t, repo.Create(context.Background(), c2))

	err = repo.Create(context.Background(), &client.Client{ID: "c3", DocumentNumber: "111"})
	require.ErrorIs(t, err, repository.ErrConflict)

	got, err := repo.FindByDocument(context.Background(), "111")
	require.NoError(t, err)
	assert.Equal(t, "c1", got.ID)

	_, err = repo.FindByDocument(context.Background(), "missing")
	require.ErrorIs(t, err, repository.ErrNotFound)

	out, err := repo.List(context.Background(), 10, 0)
	require.NoError(t, err)
	require.Len(t, out, 2)
	assert.Equal(t, "c2", out[0].ID)
	assert.Equal(t, "c1", out[1].ID)

	err = repo.Update(context.Background(), &client.Client{ID: "missing", DocumentNumber: "999"})
	require.ErrorIs(t, err, repository.ErrNotFound)

	err = repo.Update(context.Background(), &client.Client{ID: "c1", DocumentNumber: "222"})
	require.ErrorIs(t, err, repository.ErrConflict)

	require.NoError(t, repo.Update(context.Background(), &client.Client{ID: "c1", DocumentNumber: "333", Name: "A2", CreatedAt: t0, UpdatedAt: t0}))
	got, err = repo.FindByID(context.Background(), "c1")
	require.NoError(t, err)
	assert.Equal(t, "333", got.DocumentNumber)
	assert.Equal(t, "A2", got.Name)

	err = repo.Delete(context.Background(), "missing")
	require.ErrorIs(t, err, repository.ErrNotFound)

	require.NoError(t, repo.Delete(context.Background(), "c1"))
	_, err = repo.FindByID(context.Background(), "c1")
	require.ErrorIs(t, err, repository.ErrNotFound)
}

func TestVehicleRepository_CRUD_ListByClientID_FindByPlate(t *testing.T) {
	t.Parallel()

	s := NewStore()
	repo := NewVehicleRepository(s)

	err := repo.Create(context.Background(), nil)
	require.Error(t, err)

	err = repo.Create(context.Background(), &vehicle.Vehicle{ID: "v1", Plate: "   "})
	require.Error(t, err)

	t0 := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
	v1 := &vehicle.Vehicle{ID: "v1", ClientID: "c1", Plate: "AAA0001", Brand: "X", Model: "Y", ModelYear: 2020, CreatedAt: t0, UpdatedAt: t0}
	v2 := &vehicle.Vehicle{ID: "v2", ClientID: "c1", Plate: "BBB0002", Brand: "X", Model: "Y", ModelYear: 2021, CreatedAt: t0.Add(1 * time.Hour), UpdatedAt: t0.Add(1 * time.Hour)}
	v3 := &vehicle.Vehicle{ID: "v3", ClientID: "c2", Plate: "CCC0003", Brand: "X", Model: "Y", ModelYear: 2022, CreatedAt: t0.Add(2 * time.Hour), UpdatedAt: t0.Add(2 * time.Hour)}
	require.NoError(t, repo.Create(context.Background(), v1))
	require.NoError(t, repo.Create(context.Background(), v2))
	require.NoError(t, repo.Create(context.Background(), v3))

	err = repo.Create(context.Background(), &vehicle.Vehicle{ID: "v4", Plate: "AAA0001"})
	require.ErrorIs(t, err, repository.ErrConflict)

	got, err := repo.FindByPlate(context.Background(), "AAA0001")
	require.NoError(t, err)
	assert.Equal(t, "v1", got.ID)

	_, err = repo.FindByPlate(context.Background(), "missing")
	require.ErrorIs(t, err, repository.ErrNotFound)

	out, err := repo.ListByClientID(context.Background(), "c1", 10, 0)
	require.NoError(t, err)
	require.Len(t, out, 2)
	assert.Equal(t, "v2", out[0].ID)
	assert.Equal(t, "v1", out[1].ID)

	err = repo.Update(context.Background(), &vehicle.Vehicle{ID: "missing", Plate: "ZZZ0000"})
	require.ErrorIs(t, err, repository.ErrNotFound)

	err = repo.Update(context.Background(), &vehicle.Vehicle{ID: "v1", Plate: "BBB0002"})
	require.ErrorIs(t, err, repository.ErrConflict)

	require.NoError(t, repo.Update(context.Background(), &vehicle.Vehicle{ID: "v1", ClientID: "c1", Plate: "AAA0001", Brand: "Z", Model: "Y", ModelYear: 2020}))
	got, err = repo.FindByID(context.Background(), "v1")
	require.NoError(t, err)
	assert.Equal(t, "Z", got.Brand)

	err = repo.Delete(context.Background(), "missing")
	require.ErrorIs(t, err, repository.ErrNotFound)

	require.NoError(t, repo.Delete(context.Background(), "v1"))
	_, err = repo.FindByID(context.Background(), "v1")
	require.ErrorIs(t, err, repository.ErrNotFound)
}

func TestServiceRepository_CRUD_FindByIDs_NormalizesName(t *testing.T) {
	t.Parallel()

	s := NewStore()
	repo := NewServiceRepository(s)

	err := repo.Create(context.Background(), nil)
	require.Error(t, err)

	err = repo.Create(context.Background(), &service.Service{ID: "s1", Name: "   "})
	require.Error(t, err)

	t0 := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
	s1 := &service.Service{ID: "s1", Name: "  Troca de Óleo  ", BasePriceCents: 1000, EstimatedMinutes: 30, Active: true, CreatedAt: t0, UpdatedAt: t0}
	require.NoError(t, repo.Create(context.Background(), s1))

	err = repo.Create(context.Background(), &service.Service{ID: "s2", Name: "troca de óleo"})
	require.ErrorIs(t, err, repository.ErrConflict)

	err = repo.Update(context.Background(), &service.Service{ID: "missing", Name: "x"})
	require.ErrorIs(t, err, repository.ErrNotFound)

	s2 := &service.Service{ID: "s2", Name: "Alinhamento", BasePriceCents: 2000, EstimatedMinutes: 40, Active: true, CreatedAt: t0.Add(1 * time.Hour), UpdatedAt: t0.Add(1 * time.Hour)}
	require.NoError(t, repo.Create(context.Background(), s2))

	err = repo.Update(context.Background(), &service.Service{ID: "s2", Name: "TROCA DE ÓLEO"})
	require.ErrorIs(t, err, repository.ErrConflict)

	out, err := repo.FindByIDs(context.Background(), []string{"s2", "missing", "s2", "s1"})
	require.NoError(t, err)
	require.Len(t, out, 2)

	list, err := repo.List(context.Background(), 10, 0)
	require.NoError(t, err)
	require.Len(t, list, 2)
	assert.Equal(t, "s2", list[0].ID)

	err = repo.Delete(context.Background(), "missing")
	require.ErrorIs(t, err, repository.ErrNotFound)
	require.NoError(t, repo.Delete(context.Background(), "s1"))
	_, err = repo.FindByID(context.Background(), "s1")
	require.ErrorIs(t, err, repository.ErrNotFound)
}

func TestPartRepository_CRUD_FindByIDs_AdjustStock(t *testing.T) {
	t.Parallel()

	s := NewStore()
	repo := NewPartRepository(s)

	err := repo.Create(context.Background(), nil)
	require.Error(t, err)

	err = repo.Create(context.Background(), &part.Part{ID: "p1", SKU: "   "})
	require.Error(t, err)

	t0 := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
	p1 := &part.Part{ID: "p1", SKU: "  SKU-1  ", Name: "Filtro", UnitPriceCents: 1000, StockQuantity: 10, Active: true, CreatedAt: t0, UpdatedAt: t0}
	require.NoError(t, repo.Create(context.Background(), p1))

	err = repo.Create(context.Background(), &part.Part{ID: "p2", SKU: "sku-1"})
	require.ErrorIs(t, err, repository.ErrConflict)

	err = repo.Update(context.Background(), &part.Part{ID: "missing", SKU: "x"})
	require.ErrorIs(t, err, repository.ErrNotFound)

	p2 := &part.Part{ID: "p2", SKU: "SKU-2", Name: "Óleo", UnitPriceCents: 5000, StockQuantity: 1, Active: true, CreatedAt: t0.Add(1 * time.Hour), UpdatedAt: t0.Add(1 * time.Hour)}
	require.NoError(t, repo.Create(context.Background(), p2))

	err = repo.Update(context.Background(), &part.Part{ID: "p2", SKU: "SKU-1"})
	require.ErrorIs(t, err, repository.ErrConflict)

	require.NoError(t, repo.Update(context.Background(), &part.Part{ID: "p2", SKU: "SKU-2", Name: "Óleo Sintético", UnitPriceCents: 6000, StockQuantity: 1, Active: true, CreatedAt: p2.CreatedAt, UpdatedAt: p2.UpdatedAt}))
	gotPart, err := repo.FindByID(context.Background(), "p2")
	require.NoError(t, err)
	assert.Equal(t, "Óleo Sintético", gotPart.Name)

	out, err := repo.FindByIDs(context.Background(), []string{"p2", "missing", "p2", "p1"})
	require.NoError(t, err)
	require.Len(t, out, 2)

	_, err = repo.AdjustStock(context.Background(), "p1", part.StockMovementIn, 0, nil, nil)
	require.Error(t, err)

	_, err = repo.AdjustStock(context.Background(), "missing", part.StockMovementIn, 1, nil, nil)
	require.ErrorIs(t, err, repository.ErrNotFound)

	_, err = repo.AdjustStock(context.Background(), "p1", "NOPE", 1, nil, nil)
	require.Error(t, err)

	_, err = repo.AdjustStock(context.Background(), "p2", part.StockMovementOut, 10, nil, nil)
	require.Error(t, err)

	got, err := repo.AdjustStock(context.Background(), "p1", part.StockMovementOut, 2, nil, nil)
	require.NoError(t, err)
	assert.Equal(t, 8, got.StockQuantity)

	got, err = repo.AdjustStock(context.Background(), "p1", part.StockMovementAdjustment, 7, nil, nil)
	require.NoError(t, err)
	assert.Equal(t, 7, got.StockQuantity)

	list, err := repo.List(context.Background(), 10, 0)
	require.NoError(t, err)
	require.Len(t, list, 2)
	assert.ElementsMatch(t, []string{"p1", "p2"}, []string{list[0].ID, list[1].ID})
}

func TestServiceOrderRepository_List_Detail_AverageExecutionMinutes(t *testing.T) {
	t.Parallel()

	s := NewStore()

	s.clientsByID["c1"] = &client.Client{ID: "c1", Name: "Maria", DocumentNumber: "111"}
	s.vehiclesByID["v1"] = &vehicle.Vehicle{ID: "v1", Plate: "AAA0001"}

	now := time.Now().UTC()
	openedAt := now.Add(-10 * time.Hour)
	execStart := openedAt.Add(1 * time.Hour)
	finishedAt := execStart.Add(90 * time.Minute)
	so := &order.ServiceOrder{ID: "so1", Code: "OS-1", ClientID: "c1", VehicleID: "v1", Status: order.StatusFinished, OpenedAt: openedAt, ExecutionStart: &execStart, FinishedAt: &finishedAt}
	s.serviceOrdersByID["so1"] = so
	s.serviceOrdersByCode["OS-1"] = so

	so2 := &order.ServiceOrder{ID: "so2", Code: "OS-2", ClientID: "c1", VehicleID: "v1", Status: order.StatusReceived, OpenedAt: openedAt.Add(-1 * time.Hour)}
	s.serviceOrdersByID["so2"] = so2
	s.serviceOrdersByCode["OS-2"] = so2

	s.budgetsByServiceOrder["so1"] = []*order.Budget{
		{ID: "b1", ServiceOrderID: "so1", Version: 1, Status: order.BudgetStatusDraft, TotalAmountCents: 100},
		{ID: "b2", ServiceOrderID: "so1", Version: 2, Status: order.BudgetStatusSent, TotalAmountCents: 200},
	}
	s.budgetServicesByBudget["b2"] = []order.BudgetServiceItem{{ServiceID: "s1", Description: "X", Quantity: 1}}
	s.budgetPartsByBudget["b2"] = []order.BudgetPartItem{{PartID: "p1", Description: "Y", Quantity: 2}}
	s.statusHistoryBySO["so1"] = []order.StatusHistoryEntry{{ToStatus: order.StatusFinished, ChangedAt: now}}

	repo := NewServiceOrderRepository(s)

	_, err := repo.FindByID(context.Background(), "missing")
	require.ErrorIs(t, err, repository.ErrNotFound)

	gotSO, err := repo.FindByID(context.Background(), "so1")
	require.NoError(t, err)
	assert.Equal(t, "OS-1", gotSO.Code)

	status := order.StatusFinished
	sums, err := repo.List(context.Background(), 1, 0, &status)
	require.NoError(t, err)
	require.Len(t, sums, 1)
	assert.Equal(t, "so1", sums[0].ID)
	assert.Equal(t, "Maria", sums[0].ClientName)
	assert.Equal(t, "AAA0001", sums[0].Plate)

	sums, err = repo.List(context.Background(), -1, -5, nil)
	require.NoError(t, err)
	require.Len(t, sums, 2)

	sums, err = repo.List(context.Background(), 10, 999, nil)
	require.NoError(t, err)
	assert.Empty(t, sums)

	detail, err := repo.GetDetailByID(context.Background(), "so1")
	require.NoError(t, err)
	require.NotNil(t, detail)
	require.NotNil(t, detail.LatestBudget)
	assert.Equal(t, "b2", detail.LatestBudget.ID)
	assert.Len(t, detail.BudgetServices, 1)
	assert.Len(t, detail.BudgetParts, 1)
	assert.Len(t, detail.StatusHistory, 1)

	_, err = repo.GetDetailByID(context.Background(), "missing")
	require.ErrorIs(t, err, repository.ErrNotFound)

	avg, err := repo.AverageExecutionMinutes(context.Background(), nil, nil)
	require.NoError(t, err)
	assert.InDelta(t, 90.0, avg, 0.0001)

	from := now.Add(-2 * time.Hour)
	avg, err = repo.AverageExecutionMinutes(context.Background(), &from, nil)
	require.NoError(t, err)
	assert.Equal(t, 0.0, avg)
}

func TestServiceOrderFlowRepository_HappyPath_AndErrors(t *testing.T) {
	t.Parallel()

	s := NewStore()
	partsRepo := NewPartRepository(s)
	flow := NewServiceOrderFlowRepository(s, partsRepo)

	c := &client.Client{ID: "c1", DocumentNumber: "111", Name: "Maria"}
	s.clientsByID[c.ID] = c
	s.clientsByDocument[c.DocumentNumber] = c
	v := &vehicle.Vehicle{ID: "v1", ClientID: "c1", Plate: "AAA0001"}
	s.vehiclesByID[v.ID] = v
	s.vehiclesByPlate[v.Plate] = v

	// Seed parts (for budget approval stock consumption)
	p1 := &part.Part{ID: "p1", SKU: "sku-1", Name: "Filtro", UnitPriceCents: 1000, StockQuantity: 1, Active: true}
	require.NoError(t, partsRepo.Create(context.Background(), p1))

	so, b, err := flow.CreateDraft(context.Background(), repository.CreateServiceOrderDraftParams{
		ServiceOrder: order.ServiceOrder{ClientID: "c1", VehicleID: "v1"},
		Budget:       order.Budget{TotalAmountCents: 100},
		BudgetParts:  []order.BudgetPartItem{{PartID: "p1", Quantity: 1}},
	})
	require.NoError(t, err)
	require.NotNil(t, so)
	require.NotNil(t, b)
	assert.NotEmpty(t, so.ID)
	assert.NotEmpty(t, so.Code)
	assert.Equal(t, order.StatusReceived, so.Status)
	assert.Equal(t, order.BudgetStatusDraft, b.Status)

	_, _, err = flow.CreateDraft(context.Background(), repository.CreateServiceOrderDraftParams{
		ServiceOrder: order.ServiceOrder{ID: "so2", Code: so.Code, ClientID: "c1", VehicleID: "v1"},
		Budget:       order.Budget{ID: "b2", TotalAmountCents: 100},
	})
	require.ErrorIs(t, err, repository.ErrConflict)

	_, err = flow.CreateBudgetRevision(context.Background(), repository.CreateBudgetRevisionParams{ServiceOrderID: "missing"})
	require.ErrorIs(t, err, repository.ErrNotFound)

	_, err = flow.CreateBudgetRevision(context.Background(), repository.CreateBudgetRevisionParams{ServiceOrderID: so.ID})
	require.ErrorIs(t, err, repository.ErrConflict)

	require.NoError(t, flow.StartDiagnosis(context.Background(), so.ID, nil))

	b2, err := flow.CreateBudgetRevision(context.Background(), repository.CreateBudgetRevisionParams{
		ServiceOrderID: so.ID,
		Budget:         order.Budget{TotalAmountCents: 200},
	})
	require.NoError(t, err)
	require.NotNil(t, b2)
	assert.GreaterOrEqual(t, b2.Version, 2)

	err = flow.SendLatestBudget(context.Background(), "missing", nil)
	require.ErrorIs(t, err, repository.ErrNotFound)

	// missing budget on existing SO
	delete(s.budgetsByServiceOrder, so.ID)
	err = flow.SendLatestBudget(context.Background(), so.ID, nil)
	require.ErrorIs(t, err, repository.ErrNotFound)

	// restore budgets and set latest as DRAFT again
	s.budgetsByServiceOrder[so.ID] = []*order.Budget{{ID: b.ID, ServiceOrderID: so.ID, Version: 1, Status: order.BudgetStatusDraft}}
	// approve requires SENT (before sending)
	err = flow.ApproveLatestBudgetByCode(context.Background(), so.Code, "111", nil)
	require.Error(t, err)

	require.NoError(t, flow.SendLatestBudget(context.Background(), so.ID, nil))

	// set latest to SENT and test insufficient stock
	now := time.Now().UTC()
	latest := s.budgetsByServiceOrder[so.ID][0]
	latest.Status = order.BudgetStatusSent
	latest.SentAt = &now
	s.budgetPartsByBudget[latest.ID] = []order.BudgetPartItem{{PartID: "p1", Quantity: 2}}
	err = flow.ApproveLatestBudgetByCode(context.Background(), so.Code, "111", nil)
	require.Error(t, err)

	// success approval: consume stock and transition to IN_PROGRESS
	latest.Status = order.BudgetStatusSent
	latest.ApprovedAt = nil
	latest.ApprovedByName = nil
	s.budgetPartsByBudget[latest.ID] = []order.BudgetPartItem{{PartID: "p1", Quantity: 1}}
	err = flow.ApproveLatestBudgetByCode(context.Background(), so.Code, "111", nil)
	require.NoError(t, err)
	assert.Equal(t, 0, s.partsByID["p1"].StockQuantity)
	assert.Equal(t, order.StatusInProgress, s.serviceOrdersByID[so.ID].Status)
	require.NotNil(t, s.serviceOrdersByID[so.ID].ExecutionStart)

	// finish & deliver
	require.NoError(t, flow.Finish(context.Background(), so.ID, nil))
	require.NoError(t, flow.Deliver(context.Background(), so.ID, nil))
	require.NotNil(t, s.serviceOrdersByID[so.ID].FinishedAt)
	require.NotNil(t, s.serviceOrdersByID[so.ID].DeliveredAt)

	// client view
	view, err := flow.GetClientViewByCode(context.Background(), so.Code, "111")
	require.NoError(t, err)
	assert.Equal(t, so.Code, view.Code)
	assert.NotEmpty(t, view.OpenedAt)

	_, err = flow.GetClientViewByCode(context.Background(), "missing", "111")
	require.ErrorIs(t, err, repository.ErrNotFound)

	_, err = flow.GetClientViewByCode(context.Background(), so.Code, "wrong")
	require.ErrorIs(t, err, repository.ErrNotFound)

	// reject path requires SENT
	err = flow.RejectLatestBudgetByCode(context.Background(), so.Code, "111", "nope")
	require.Error(t, err)

	// rejection happy path (separate SO)
	p2 := &part.Part{ID: "p2", SKU: "sku-2", Name: "Velas", UnitPriceCents: 3000, StockQuantity: 10, Active: true}
	require.NoError(t, partsRepo.Create(context.Background(), p2))

	soR, bR, err := flow.CreateDraft(context.Background(), repository.CreateServiceOrderDraftParams{
		ServiceOrder: order.ServiceOrder{ClientID: "c1", VehicleID: "v1"},
		Budget:       order.Budget{TotalAmountCents: 100},
		BudgetParts:  []order.BudgetPartItem{{PartID: "p2", Quantity: 1}},
	})
	require.NoError(t, err)
	require.NoError(t, flow.SendLatestBudget(context.Background(), soR.ID, nil))

	err = flow.RejectLatestBudgetByCode(context.Background(), soR.Code, "111", "Cliente pediu ajuste")
	require.NoError(t, err)

	require.Equal(t, order.StatusInDiagnosis, s.serviceOrdersByID[soR.ID].Status)
	require.Equal(t, order.BudgetStatusRejected, s.budgetsByID[bR.ID].Status)
	require.NotNil(t, s.budgetsByID[bR.ID].RejectedAt)
	require.NotNil(t, s.budgetsByID[bR.ID].RejectionReason)
}

func TestServiceOrderFlowRepository_TransitionRules_AndHelpers(t *testing.T) {
	t.Parallel()

	require.True(t, allowedTransition(order.StatusReceived, order.StatusInDiagnosis))
	require.False(t, allowedTransition(order.StatusReceived, order.StatusDelivered))
	require.False(t, allowedTransition("UNKNOWN", order.StatusReceived))

	require.Nil(t, latestBudget(nil))
	require.Nil(t, latestBudget([]*order.Budget{}))
	b1 := &order.Budget{ID: "b1", Version: 1}
	b2 := &order.Budget{ID: "b2", Version: 3}
	b3 := &order.Budget{ID: "b3", Version: 2}
	assert.Equal(t, b2, latestBudget([]*order.Budget{b1, b2, b3}))

	assert.Equal(t, " x ", nonEmptyID(" x "))
	gotID := nonEmptyID("")
	assert.NotEmpty(t, gotID)

	items := []int{1, 2, 3, 4, 5}
	out := paginate(items, 0, -1, func(i int) int { return i })
	assert.Len(t, out, 5)
	out = paginate(items, 2, 3, func(i int) int { return i })
	assert.Equal(t, []int{4, 5}, out)
	out = paginate(items, 2, 999, func(i int) int { return i })
	assert.Empty(t, out)
}
