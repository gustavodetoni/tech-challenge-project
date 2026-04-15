package controllers_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/soat-architecture/tech-challenge-project/internal/application/admin"
	"github.com/soat-architecture/tech-challenge-project/internal/domain/order"
	"github.com/soat-architecture/tech-challenge-project/internal/interfaces/http/controllers"
	"github.com/soat-architecture/tech-challenge-project/internal/interfaces/repository"
	repomocks "github.com/soat-architecture/tech-challenge-project/internal/tests/interfaces/repository/mocks"
)

func TestAdminServiceOrdersController_List_Success(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)
	repo := new(repomocks.ServiceOrderRepository)
	svc := admin.NewServiceOrderService(repo)
	h := controllers.NewAdminServiceOrdersController(svc)

	openedAt := time.Date(2026, 4, 1, 10, 0, 0, 0, time.UTC)
	repo.On("List", mock.Anything, 50, 0, mock.MatchedBy(func(v *order.Status) bool { return v != nil && *v == order.StatusReceived })).
		Return([]order.ServiceOrderSummary{{
			ID:         "so1",
			Code:       "C-1",
			Status:     order.StatusReceived,
			OpenedAt:   openedAt,
			ClientID:   "c1",
			ClientName: "Maria",
			VehicleID:  "v1",
			Plate:      "ABC1D23",
		}}, nil).Once()

	r := gin.New()
	r.GET("/admin/service-orders", h.List)

	req := httptest.NewRequest(http.MethodGet, "/admin/service-orders?status=RECEIVED", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"id":"so1"`)
	repo.AssertExpectations(t)
}

func TestAdminServiceOrdersController_List_Success_NoStatusFilter(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)
	repo := new(repomocks.ServiceOrderRepository)
	svc := admin.NewServiceOrderService(repo)
	h := controllers.NewAdminServiceOrdersController(svc)

	repo.On("List", mock.Anything, 50, 0, (*order.Status)(nil)).Return([]order.ServiceOrderSummary{}, nil).Once()

	r := gin.New()
	r.GET("/admin/service-orders", h.List)

	req := httptest.NewRequest(http.MethodGet, "/admin/service-orders", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	repo.AssertExpectations(t)
}

func TestAdminServiceOrdersController_Get_NotFound(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)
	repo := new(repomocks.ServiceOrderRepository)
	svc := admin.NewServiceOrderService(repo)
	h := controllers.NewAdminServiceOrdersController(svc)

	repo.On("GetDetailByID", mock.Anything, "so1").Return((*order.ServiceOrderDetail)(nil), repository.ErrNotFound).Once()

	r := gin.New()
	r.GET("/admin/service-orders/:id", h.Get)

	req := httptest.NewRequest(http.MethodGet, "/admin/service-orders/so1", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	repo.AssertExpectations(t)
}

func TestAdminServiceOrdersController_Get_Success_WithBudgetAndHistory(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)
	repo := new(repomocks.ServiceOrderRepository)
	svc := admin.NewServiceOrderService(repo)
	h := controllers.NewAdminServiceOrdersController(svc)

	openedAt := time.Date(2026, 4, 1, 10, 0, 0, 0, time.UTC)
	sentAt := time.Date(2026, 4, 2, 10, 0, 0, 0, time.UTC)
	approvedAt := time.Date(2026, 4, 3, 10, 0, 0, 0, time.UTC)

	repo.On("GetDetailByID", mock.Anything, "so1").Return(&order.ServiceOrderDetail{
		ServiceOrder: order.ServiceOrder{
			ID:        "so1",
			Code:      "C-1",
			Status:    order.StatusWaitingApproval,
			ClientID:  "c1",
			VehicleID: "v1",
			OpenedAt:  openedAt,
		},
		LatestBudget: &order.Budget{
			ID:               "b1",
			ServiceOrderID:   "so1",
			Version:          2,
			Status:           order.BudgetStatusSent,
			TotalAmountCents: 110,
			SentAt:           &sentAt,
			ApprovedAt:       &approvedAt,
		},
		BudgetServices: []order.BudgetServiceItem{{ServiceID: "s1", Description: "Svc", Quantity: 1, UnitPriceCents: 100, TotalPriceCents: 100}},
		BudgetParts:    []order.BudgetPartItem{{PartID: "p1", Description: "Part", Quantity: 1, UnitPriceCents: 10, TotalPriceCents: 10}},
		StatusHistory: []order.StatusHistoryEntry{{
			FromStatus:      nil,
			ToStatus:        order.StatusWaitingApproval,
			ChangedAt:       approvedAt,
			ChangedByUserID: nil,
			Reason:          nil,
		}},
	}, nil).Once()

	r := gin.New()
	r.GET("/admin/service-orders/:id", h.Get)

	req := httptest.NewRequest(http.MethodGet, "/admin/service-orders/so1", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"latest_budget"`)
	repo.AssertExpectations(t)
}

func TestAdminServiceOrdersController_Get_Success_Minimal(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)
	repo := new(repomocks.ServiceOrderRepository)
	svc := admin.NewServiceOrderService(repo)
	h := controllers.NewAdminServiceOrdersController(svc)

	openedAt := time.Date(2026, 4, 1, 10, 0, 0, 0, time.UTC)
	repo.On("GetDetailByID", mock.Anything, "so1").Return(&order.ServiceOrderDetail{
		ServiceOrder: order.ServiceOrder{
			ID:        "so1",
			Code:      "C-1",
			Status:    order.StatusReceived,
			ClientID:  "c1",
			VehicleID: "v1",
			OpenedAt:  openedAt,
		},
		BudgetServices: []order.BudgetServiceItem{},
		BudgetParts:    []order.BudgetPartItem{},
		StatusHistory:  []order.StatusHistoryEntry{},
	}, nil).Once()

	r := gin.New()
	r.GET("/admin/service-orders/:id", h.Get)

	req := httptest.NewRequest(http.MethodGet, "/admin/service-orders/so1", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	var out map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &out))
	assert.Equal(t, "so1", out["id"])
	assert.Equal(t, "C-1", out["code"])
	repo.AssertExpectations(t)
}
