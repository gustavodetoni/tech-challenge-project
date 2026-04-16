package controllers_test

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/soat-architecture/tech-challenge-project/internal/application/serviceorder"
	"github.com/soat-architecture/tech-challenge-project/internal/domain/client"
	"github.com/soat-architecture/tech-challenge-project/internal/domain/order"
	"github.com/soat-architecture/tech-challenge-project/internal/interfaces/http/controllers"
	"github.com/soat-architecture/tech-challenge-project/internal/interfaces/repository"
	repomocks "github.com/soat-architecture/tech-challenge-project/internal/tests/interfaces/repository/mocks"
)

func TestClientServiceOrderController_Get_BadRequest_NoDocumentNumber(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)
	flowRepo := new(repomocks.ServiceOrderFlowRepository)
	svc := serviceorder.NewService(new(repomocks.ClientRepository), new(repomocks.VehicleRepository), new(repomocks.ServiceRepository), new(repomocks.PartRepository), flowRepo)
	h := controllers.NewClientServiceOrderController(svc)

	r := gin.New()
	r.GET("/client/service-orders/:code", h.Get)

	req := httptest.NewRequest(http.MethodGet, "/client/service-orders/C-1", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestClientServiceOrderController_Get_Success(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)
	flowRepo := new(repomocks.ServiceOrderFlowRepository)
	svc := serviceorder.NewService(new(repomocks.ClientRepository), new(repomocks.VehicleRepository), new(repomocks.ServiceRepository), new(repomocks.PartRepository), flowRepo)
	h := controllers.NewClientServiceOrderController(svc)

	flowRepo.On("GetClientViewByCode", mock.Anything, "C-1", "46420082412").Return(&repository.ClientServiceOrderView{
		Code:             "C-1",
		Status:           order.StatusReceived,
		OpenedAt:         "2026-04-01T10:00:00Z",
		BudgetStatus:     order.BudgetStatusDraft,
		BudgetTotalCents: 100,
	}, nil).Once()

	r := gin.New()
	r.GET("/client/service-orders/:code", h.Get)

	req := httptest.NewRequest(http.MethodGet, "/client/service-orders/C-1?document_number=46420082412", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"code":"C-1"`)
	flowRepo.AssertExpectations(t)
}

func TestClientServiceOrderController_Get_NotFound(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)
	flowRepo := new(repomocks.ServiceOrderFlowRepository)
	svc := serviceorder.NewService(new(repomocks.ClientRepository), new(repomocks.VehicleRepository), new(repomocks.ServiceRepository), new(repomocks.PartRepository), flowRepo)
	h := controllers.NewClientServiceOrderController(svc)

	flowRepo.On("GetClientViewByCode", mock.Anything, "C-1", "46420082412").Return((*repository.ClientServiceOrderView)(nil), repository.ErrNotFound).Once()

	r := gin.New()
	r.GET("/client/service-orders/:code", h.Get)

	req := httptest.NewRequest(http.MethodGet, "/client/service-orders/C-1?document_number=46420082412", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	flowRepo.AssertExpectations(t)
}

func TestClientServiceOrderController_Get_InternalError(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)
	flowRepo := new(repomocks.ServiceOrderFlowRepository)
	svc := serviceorder.NewService(new(repomocks.ClientRepository), new(repomocks.VehicleRepository), new(repomocks.ServiceRepository), new(repomocks.PartRepository), flowRepo)
	h := controllers.NewClientServiceOrderController(svc)

	flowRepo.On("GetClientViewByCode", mock.Anything, "C-1", "46420082412").Return((*repository.ClientServiceOrderView)(nil), assert.AnError).Once()

	r := gin.New()
	r.GET("/client/service-orders/:code", h.Get)

	req := httptest.NewRequest(http.MethodGet, "/client/service-orders/C-1?document_number=46420082412", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	flowRepo.AssertExpectations(t)
}

func TestClientServiceOrderController_ApproveBudget_NotFound(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)
	flowRepo := new(repomocks.ServiceOrderFlowRepository)
	clientRepo := new(repomocks.ClientRepository)
	svc := serviceorder.NewService(clientRepo, new(repomocks.VehicleRepository), new(repomocks.ServiceRepository), new(repomocks.PartRepository), flowRepo)
	h := controllers.NewClientServiceOrderController(svc)

	clientRepo.On("FindByDocument", mock.Anything, "46420082412").Return(&client.Client{
		ID:             "cl-1",
		DocumentNumber: "46420082412",
		Name:           "",
	}, nil).Once()
	flowRepo.On("ApproveLatestBudgetByCode", mock.Anything, "C-1", "46420082412", (*string)(nil)).Return(repository.ErrNotFound).Once()

	r := gin.New()
	r.POST("/client/service-orders/:code/budget/approve", h.ApproveBudget)

	req := httptest.NewRequest(http.MethodPost, "/client/service-orders/C-1/budget/approve?document_number=46420082412", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	clientRepo.AssertExpectations(t)
	flowRepo.AssertExpectations(t)
}

func TestClientServiceOrderController_ApproveBudget_BadRequest_NoDocumentNumber(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)
	flowRepo := new(repomocks.ServiceOrderFlowRepository)
	svc := serviceorder.NewService(new(repomocks.ClientRepository), new(repomocks.VehicleRepository), new(repomocks.ServiceRepository), new(repomocks.PartRepository), flowRepo)
	h := controllers.NewClientServiceOrderController(svc)

	r := gin.New()
	r.POST("/client/service-orders/:code/budget/approve", h.ApproveBudget)

	req := httptest.NewRequest(http.MethodPost, "/client/service-orders/C-1/budget/approve", bytes.NewBufferString(`{}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestClientServiceOrderController_ApproveBudget_BadRequest_GenericError(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)
	flowRepo := new(repomocks.ServiceOrderFlowRepository)
	clientRepo := new(repomocks.ClientRepository)
	svc := serviceorder.NewService(clientRepo, new(repomocks.VehicleRepository), new(repomocks.ServiceRepository), new(repomocks.PartRepository), flowRepo)
	h := controllers.NewClientServiceOrderController(svc)

	clientRepo.On("FindByDocument", mock.Anything, "46420082412").Return(&client.Client{
		ID:             "cl-1",
		DocumentNumber: "46420082412",
		Name:           "",
	}, nil).Once()
	flowRepo.On("ApproveLatestBudgetByCode", mock.Anything, "C-1", "46420082412", (*string)(nil)).Return(assert.AnError).Once()

	r := gin.New()
	r.POST("/client/service-orders/:code/budget/approve", h.ApproveBudget)

	req := httptest.NewRequest(http.MethodPost, "/client/service-orders/C-1/budget/approve?document_number=46420082412", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	clientRepo.AssertExpectations(t)
	flowRepo.AssertExpectations(t)
}

func TestClientServiceOrderController_ApproveBudget_Success(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)
	flowRepo := new(repomocks.ServiceOrderFlowRepository)
	clientRepo := new(repomocks.ClientRepository)
	svc := serviceorder.NewService(clientRepo, new(repomocks.VehicleRepository), new(repomocks.ServiceRepository), new(repomocks.PartRepository), flowRepo)
	h := controllers.NewClientServiceOrderController(svc)

	clientRepo.On("FindByDocument", mock.Anything, "46420082412").Return(&client.Client{
		ID:             "cl-1",
		DocumentNumber: "46420082412",
		Name:           "",
	}, nil).Once()
	flowRepo.On("ApproveLatestBudgetByCode", mock.Anything, "C-1", "46420082412", (*string)(nil)).Return(nil).Once()

	r := gin.New()
	r.POST("/client/service-orders/:code/budget/approve", h.ApproveBudget)

	req := httptest.NewRequest(http.MethodPost, "/client/service-orders/C-1/budget/approve?document_number=46420082412", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)
	clientRepo.AssertExpectations(t)
	flowRepo.AssertExpectations(t)
}

func TestClientServiceOrderController_ApproveBudget_Success_UsesClientName(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)
	flowRepo := new(repomocks.ServiceOrderFlowRepository)
	clientRepo := new(repomocks.ClientRepository)
	svc := serviceorder.NewService(clientRepo, new(repomocks.VehicleRepository), new(repomocks.ServiceRepository), new(repomocks.PartRepository), flowRepo)
	h := controllers.NewClientServiceOrderController(svc)

	clientRepo.On("FindByDocument", mock.Anything, "46420082412").Return(&client.Client{
		ID:             "cl-1",
		DocumentNumber: "46420082412",
		Name:           "Maria",
	}, nil).Once()
	flowRepo.On("ApproveLatestBudgetByCode", mock.Anything, "C-1", "46420082412", mock.MatchedBy(func(v *string) bool {
		return v != nil && *v == "Maria"
	})).Return(nil).Once()

	r := gin.New()
	r.POST("/client/service-orders/:code/budget/approve", h.ApproveBudget)

	req := httptest.NewRequest(http.MethodPost, "/client/service-orders/C-1/budget/approve?document_number=46420082412", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)
	clientRepo.AssertExpectations(t)
	flowRepo.AssertExpectations(t)
}

func TestClientServiceOrderController_RejectBudget_BadRequest_InvalidBody(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)
	flowRepo := new(repomocks.ServiceOrderFlowRepository)
	svc := serviceorder.NewService(new(repomocks.ClientRepository), new(repomocks.VehicleRepository), new(repomocks.ServiceRepository), new(repomocks.PartRepository), flowRepo)
	h := controllers.NewClientServiceOrderController(svc)

	r := gin.New()
	r.POST("/client/service-orders/:code/budget/reject", h.RejectBudget)

	req := httptest.NewRequest(http.MethodPost, "/client/service-orders/C-1/budget/reject?document_number=46420082412", bytes.NewBufferString(`{}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestClientServiceOrderController_RejectBudget_NotFound(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)
	flowRepo := new(repomocks.ServiceOrderFlowRepository)
	svc := serviceorder.NewService(new(repomocks.ClientRepository), new(repomocks.VehicleRepository), new(repomocks.ServiceRepository), new(repomocks.PartRepository), flowRepo)
	h := controllers.NewClientServiceOrderController(svc)

	flowRepo.On("RejectLatestBudgetByCode", mock.Anything, "C-1", "46420082412", "too expensive").Return(repository.ErrNotFound).Once()

	r := gin.New()
	r.POST("/client/service-orders/:code/budget/reject", h.RejectBudget)

	req := httptest.NewRequest(http.MethodPost, "/client/service-orders/C-1/budget/reject?document_number=46420082412", bytes.NewBufferString(`{"reason":"too expensive"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	flowRepo.AssertExpectations(t)
}

func TestClientServiceOrderController_RejectBudget_BadRequest_NoDocumentNumber(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)
	flowRepo := new(repomocks.ServiceOrderFlowRepository)
	svc := serviceorder.NewService(new(repomocks.ClientRepository), new(repomocks.VehicleRepository), new(repomocks.ServiceRepository), new(repomocks.PartRepository), flowRepo)
	h := controllers.NewClientServiceOrderController(svc)

	r := gin.New()
	r.POST("/client/service-orders/:code/budget/reject", h.RejectBudget)

	req := httptest.NewRequest(http.MethodPost, "/client/service-orders/C-1/budget/reject", bytes.NewBufferString(`{"reason":"too expensive"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestClientServiceOrderController_RejectBudget_BadRequest_GenericError(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)
	flowRepo := new(repomocks.ServiceOrderFlowRepository)
	svc := serviceorder.NewService(new(repomocks.ClientRepository), new(repomocks.VehicleRepository), new(repomocks.ServiceRepository), new(repomocks.PartRepository), flowRepo)
	h := controllers.NewClientServiceOrderController(svc)

	flowRepo.On("RejectLatestBudgetByCode", mock.Anything, "C-1", "46420082412", "too expensive").Return(assert.AnError).Once()

	r := gin.New()
	r.POST("/client/service-orders/:code/budget/reject", h.RejectBudget)

	req := httptest.NewRequest(http.MethodPost, "/client/service-orders/C-1/budget/reject?document_number=46420082412", bytes.NewBufferString(`{"reason":"too expensive"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	flowRepo.AssertExpectations(t)
}

func TestClientServiceOrderController_RejectBudget_Success(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)
	flowRepo := new(repomocks.ServiceOrderFlowRepository)
	svc := serviceorder.NewService(new(repomocks.ClientRepository), new(repomocks.VehicleRepository), new(repomocks.ServiceRepository), new(repomocks.PartRepository), flowRepo)
	h := controllers.NewClientServiceOrderController(svc)

	flowRepo.On("RejectLatestBudgetByCode", mock.Anything, "C-1", "46420082412", "too expensive").Return(nil).Once()

	r := gin.New()
	r.POST("/client/service-orders/:code/budget/reject", h.RejectBudget)

	req := httptest.NewRequest(http.MethodPost, "/client/service-orders/C-1/budget/reject?document_number=46420082412", bytes.NewBufferString(`{"reason":"too expensive"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)
	flowRepo.AssertExpectations(t)
}
