package controllers_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	repository "github.com/soat-architecture/tech-challenge-project/internal/application/port"
	"github.com/soat-architecture/tech-challenge-project/internal/application/serviceorder"
	"github.com/soat-architecture/tech-challenge-project/internal/domain/client"
	"github.com/soat-architecture/tech-challenge-project/internal/domain/order"
	"github.com/soat-architecture/tech-challenge-project/internal/domain/service"
	"github.com/soat-architecture/tech-challenge-project/internal/domain/vehicle"
	"github.com/soat-architecture/tech-challenge-project/internal/interfaces/http/controllers"
	repomocks "github.com/soat-architecture/tech-challenge-project/internal/tests/interfaces/repository/mocks"
)

func TestAdminServiceOrderFlowController_CreateDraft_BadRequest_InvalidBody(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)
	svc := serviceorder.NewService(new(repomocks.ClientRepository), new(repomocks.VehicleRepository), new(repomocks.ServiceRepository), new(repomocks.PartRepository), new(repomocks.ServiceOrderFlowRepository))
	h := controllers.NewAdminServiceOrderFlowController(svc)

	r := gin.New()
	r.POST("/admin/service-orders", h.CreateDraft)

	req := httptest.NewRequest(http.MethodPost, "/admin/service-orders", bytes.NewBufferString(`{`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAdminServiceOrderFlowController_CreateDraft_BadRequest_InvalidDocument(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)
	svc := serviceorder.NewService(new(repomocks.ClientRepository), new(repomocks.VehicleRepository), new(repomocks.ServiceRepository), new(repomocks.PartRepository), new(repomocks.ServiceOrderFlowRepository))
	h := controllers.NewAdminServiceOrderFlowController(svc)

	r := gin.New()
	r.POST("/admin/service-orders", h.CreateDraft)

	req := httptest.NewRequest(http.MethodPost, "/admin/service-orders", bytes.NewBufferString(`{"client_document_type":"CPF","client_document_number":"123","vehicle_plate":"ABC1D23"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAdminServiceOrderFlowController_CreateDraft_BadRequest_ServiceNotFound(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)
	clientRepo := new(repomocks.ClientRepository)
	vehicleRepo := new(repomocks.VehicleRepository)
	serviceRepo := new(repomocks.ServiceRepository)
	partRepo := new(repomocks.PartRepository)
	flowRepo := new(repomocks.ServiceOrderFlowRepository)
	svc := serviceorder.NewService(clientRepo, vehicleRepo, serviceRepo, partRepo, flowRepo)
	h := controllers.NewAdminServiceOrderFlowController(svc)

	clientRepo.On("FindByDocument", mock.Anything, "46420082412").Return(&client.Client{ID: "c1"}, nil).Once()
	vehicleRepo.On("FindByPlate", mock.Anything, "ABC1D23").Return(&vehicle.Vehicle{ID: "v1", ClientID: "c1"}, nil).Once()
	serviceRepo.On("FindByIDs", mock.Anything, mock.MatchedBy(func(ids []string) bool { return len(ids) == 1 && ids[0] == "s1" })).Return([]service.Service{}, nil).Once()

	r := gin.New()
	r.POST("/admin/service-orders", h.CreateDraft)

	req := httptest.NewRequest(http.MethodPost, "/admin/service-orders", bytes.NewBufferString(`{"client_document_type":"CPF","client_document_number":"46420082412","vehicle_plate":"ABC1D23","services":[{"id":"s1","quantity":1}]}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	clientRepo.AssertExpectations(t)
	vehicleRepo.AssertExpectations(t)
	serviceRepo.AssertExpectations(t)
}

func TestAdminServiceOrderFlowController_CreateDraft_Conflict(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)
	clientRepo := new(repomocks.ClientRepository)
	vehicleRepo := new(repomocks.VehicleRepository)
	serviceRepo := new(repomocks.ServiceRepository)
	partRepo := new(repomocks.PartRepository)
	flowRepo := new(repomocks.ServiceOrderFlowRepository)
	svc := serviceorder.NewService(clientRepo, vehicleRepo, serviceRepo, partRepo, flowRepo)
	h := controllers.NewAdminServiceOrderFlowController(svc)

	clientRepo.On("FindByDocument", mock.Anything, "46420082412").Return(&client.Client{ID: "c1"}, nil).Once()
	vehicleRepo.On("FindByPlate", mock.Anything, "ABC1D23").Return(&vehicle.Vehicle{ID: "v1", ClientID: "c1"}, nil).Once()
	serviceRepo.On("FindByIDs", mock.Anything, mock.Anything).Return([]service.Service{{
		ID:             "s1",
		Name:           "Alinhamento",
		BasePriceCents: 100,
	}}, nil).Once()
	flowRepo.On("CreateDraft", mock.Anything, mock.Anything).Return((*order.ServiceOrder)(nil), (*order.Budget)(nil), repository.ErrConflict).Once()

	r := gin.New()
	r.POST("/admin/service-orders", h.CreateDraft)

	req := httptest.NewRequest(http.MethodPost, "/admin/service-orders", bytes.NewBufferString(`{"client_document_type":"CPF","client_document_number":"46420082412","vehicle_plate":"ABC1D23","services":[{"id":"s1","quantity":1}]}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusConflict, w.Code)
	flowRepo.AssertExpectations(t)
}

func TestAdminServiceOrderFlowController_CreateDraft_Success(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)
	clientRepo := new(repomocks.ClientRepository)
	vehicleRepo := new(repomocks.VehicleRepository)
	serviceRepo := new(repomocks.ServiceRepository)
	partRepo := new(repomocks.PartRepository)
	flowRepo := new(repomocks.ServiceOrderFlowRepository)
	svc := serviceorder.NewService(clientRepo, vehicleRepo, serviceRepo, partRepo, flowRepo)
	h := controllers.NewAdminServiceOrderFlowController(svc)

	clientRepo.On("FindByDocument", mock.Anything, "46420082412").Return(&client.Client{ID: "c1"}, nil).Once()
	vehicleRepo.On("FindByPlate", mock.Anything, "ABC1D23").Return(&vehicle.Vehicle{ID: "v1", ClientID: "c1"}, nil).Once()
	serviceRepo.On("FindByIDs", mock.Anything, mock.Anything).Return([]service.Service{{
		ID:             "s1",
		Name:           "Alinhamento",
		BasePriceCents: 100,
	}}, nil).Once()
	flowRepo.On("CreateDraft", mock.Anything, mock.Anything).Return(&order.ServiceOrder{ID: "so1", Code: "C-1"}, &order.Budget{ID: "b1", Status: order.BudgetStatusDraft, TotalAmountCents: 100}, nil).Once()

	r := gin.New()
	r.POST("/admin/service-orders", h.CreateDraft)

	req := httptest.NewRequest(http.MethodPost, "/admin/service-orders", bytes.NewBufferString(`{"client_document_type":"CPF","client_document_number":"46420082412","vehicle_plate":"ABC1D23","services":[{"id":"s1","quantity":1}]}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusCreated, w.Code)
	var out map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &out))
	assert.Equal(t, "so1", out["service_order_id"])
	assert.Equal(t, "C-1", out["code"])
	assert.Equal(t, "b1", out["budget_id"])
	flowRepo.AssertExpectations(t)
}

func TestAdminServiceOrderFlowController_CreateDraft_InternalError(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)
	clientRepo := new(repomocks.ClientRepository)
	vehicleRepo := new(repomocks.VehicleRepository)
	serviceRepo := new(repomocks.ServiceRepository)
	partRepo := new(repomocks.PartRepository)
	flowRepo := new(repomocks.ServiceOrderFlowRepository)
	svc := serviceorder.NewService(clientRepo, vehicleRepo, serviceRepo, partRepo, flowRepo)
	h := controllers.NewAdminServiceOrderFlowController(svc)

	clientRepo.On("FindByDocument", mock.Anything, "46420082412").Return(&client.Client{ID: "c1"}, nil).Once()
	vehicleRepo.On("FindByPlate", mock.Anything, "ABC1D23").Return(&vehicle.Vehicle{ID: "v1", ClientID: "c1"}, nil).Once()
	serviceRepo.On("FindByIDs", mock.Anything, mock.Anything).Return([]service.Service{{
		ID:             "s1",
		Name:           "Alinhamento",
		BasePriceCents: 100,
	}}, nil).Once()
	flowRepo.On("CreateDraft", mock.Anything, mock.Anything).Return((*order.ServiceOrder)(nil), (*order.Budget)(nil), assert.AnError).Once()

	r := gin.New()
	r.POST("/admin/service-orders", h.CreateDraft)

	req := httptest.NewRequest(http.MethodPost, "/admin/service-orders", bytes.NewBufferString(`{"client_document_type":"CPF","client_document_number":"46420082412","vehicle_plate":"ABC1D23","services":[{"id":"s1","quantity":1}]}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	flowRepo.AssertExpectations(t)
}

func TestAdminServiceOrderFlowController_ReviseBudget_BadRequest_InvalidBody(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)
	svc := serviceorder.NewService(new(repomocks.ClientRepository), new(repomocks.VehicleRepository), new(repomocks.ServiceRepository), new(repomocks.PartRepository), new(repomocks.ServiceOrderFlowRepository))
	h := controllers.NewAdminServiceOrderFlowController(svc)

	r := gin.New()
	r.POST("/admin/service-orders/:id/budget/revise", h.ReviseBudget)

	req := httptest.NewRequest(http.MethodPost, "/admin/service-orders/so1/budget/revise", bytes.NewBufferString(`{`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAdminServiceOrderFlowController_ReviseBudget_BadRequest_NoItems(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)
	svc := serviceorder.NewService(new(repomocks.ClientRepository), new(repomocks.VehicleRepository), new(repomocks.ServiceRepository), new(repomocks.PartRepository), new(repomocks.ServiceOrderFlowRepository))
	h := controllers.NewAdminServiceOrderFlowController(svc)

	r := gin.New()
	r.POST("/admin/service-orders/:id/budget/revise", h.ReviseBudget)

	req := httptest.NewRequest(http.MethodPost, "/admin/service-orders/so1/budget/revise", bytes.NewBufferString(`{"services":[],"parts":[]}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAdminServiceOrderFlowController_ReviseBudget_NotFound(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)
	clientRepo := new(repomocks.ClientRepository)
	vehicleRepo := new(repomocks.VehicleRepository)
	serviceRepo := new(repomocks.ServiceRepository)
	partRepo := new(repomocks.PartRepository)
	flowRepo := new(repomocks.ServiceOrderFlowRepository)
	svc := serviceorder.NewService(clientRepo, vehicleRepo, serviceRepo, partRepo, flowRepo)
	h := controllers.NewAdminServiceOrderFlowController(svc)

	serviceRepo.On("FindByIDs", mock.Anything, mock.Anything).Return([]service.Service{{
		ID:             "s1",
		Name:           "Alinhamento",
		BasePriceCents: 100,
	}}, nil).Once()
	flowRepo.On("CreateBudgetRevision", mock.Anything, mock.Anything).Return((*order.Budget)(nil), repository.ErrNotFound).Once()

	r := gin.New()
	r.POST("/admin/service-orders/:id/budget/revise", h.ReviseBudget)

	req := httptest.NewRequest(http.MethodPost, "/admin/service-orders/so1/budget/revise", bytes.NewBufferString(`{"services":[{"id":"s1","quantity":1}],"parts":[]}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	flowRepo.AssertExpectations(t)
}

func TestAdminServiceOrderFlowController_ReviseBudget_Success_WithClaims(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)
	clientRepo := new(repomocks.ClientRepository)
	vehicleRepo := new(repomocks.VehicleRepository)
	serviceRepo := new(repomocks.ServiceRepository)
	partRepo := new(repomocks.PartRepository)
	flowRepo := new(repomocks.ServiceOrderFlowRepository)
	svc := serviceorder.NewService(clientRepo, vehicleRepo, serviceRepo, partRepo, flowRepo)
	h := controllers.NewAdminServiceOrderFlowController(svc)

	serviceRepo.On("FindByIDs", mock.Anything, mock.Anything).Return([]service.Service{{
		ID:             "s1",
		Name:           "Alinhamento",
		BasePriceCents: 100,
	}}, nil).Once()
	flowRepo.On("CreateBudgetRevision", mock.Anything, mock.MatchedBy(func(p repository.CreateBudgetRevisionParams) bool {
		return p.ServiceOrderID == "so1" && p.ChangedByUserID != nil && *p.ChangedByUserID == "subject-1"
	})).Return(&order.Budget{ID: "b2", Status: order.BudgetStatusDraft, Version: 2, TotalAmountCents: 100}, nil).Once()

	r := gin.New()
	r.Use(withClaims("ADMIN"))
	r.POST("/admin/service-orders/:id/budget/revise", h.ReviseBudget)

	req := httptest.NewRequest(http.MethodPost, "/admin/service-orders/so1/budget/revise", bytes.NewBufferString(`{"services":[{"id":"s1","quantity":1}],"parts":[]}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusCreated, w.Code)
	assert.Contains(t, w.Body.String(), `"budget_id":"b2"`)
	flowRepo.AssertExpectations(t)
}

func TestAdminServiceOrderFlowController_ReviseBudget_BadRequest_GenericError(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)
	serviceRepo := new(repomocks.ServiceRepository)
	flowRepo := new(repomocks.ServiceOrderFlowRepository)
	svc := serviceorder.NewService(new(repomocks.ClientRepository), new(repomocks.VehicleRepository), serviceRepo, new(repomocks.PartRepository), flowRepo)
	h := controllers.NewAdminServiceOrderFlowController(svc)

	serviceRepo.On("FindByIDs", mock.Anything, mock.Anything).Return([]service.Service{{
		ID:             "s1",
		Name:           "Alinhamento",
		BasePriceCents: 100,
	}}, nil).Once()
	flowRepo.On("CreateBudgetRevision", mock.Anything, mock.Anything).Return((*order.Budget)(nil), assert.AnError).Once()

	r := gin.New()
	r.POST("/admin/service-orders/:id/budget/revise", h.ReviseBudget)

	req := httptest.NewRequest(http.MethodPost, "/admin/service-orders/so1/budget/revise", bytes.NewBufferString(`{"services":[{"id":"s1","quantity":1}],"parts":[]}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	flowRepo.AssertExpectations(t)
}

func TestAdminServiceOrderFlowController_StartDiagnosis_Success_WithClaims(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)
	flowRepo := new(repomocks.ServiceOrderFlowRepository)
	svc := serviceorder.NewService(new(repomocks.ClientRepository), new(repomocks.VehicleRepository), new(repomocks.ServiceRepository), new(repomocks.PartRepository), flowRepo)
	h := controllers.NewAdminServiceOrderFlowController(svc)

	flowRepo.On("StartDiagnosis", mock.Anything, "so1", mock.MatchedBy(func(v *string) bool { return v != nil && *v == "subject-1" })).Return(nil).Once()

	r := gin.New()
	r.Use(withClaims("ADMIN"))
	r.POST("/admin/service-orders/:id/diagnosis/start", h.StartDiagnosis)

	req := httptest.NewRequest(http.MethodPost, "/admin/service-orders/so1/diagnosis/start", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusNoContent, w.Code)
	flowRepo.AssertExpectations(t)
}

func TestAdminServiceOrderFlowController_StartDiagnosis_NotFound(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)
	flowRepo := new(repomocks.ServiceOrderFlowRepository)
	svc := serviceorder.NewService(new(repomocks.ClientRepository), new(repomocks.VehicleRepository), new(repomocks.ServiceRepository), new(repomocks.PartRepository), flowRepo)
	h := controllers.NewAdminServiceOrderFlowController(svc)

	flowRepo.On("StartDiagnosis", mock.Anything, "so1", mock.Anything).Return(repository.ErrNotFound).Once()

	r := gin.New()
	r.POST("/admin/service-orders/:id/diagnosis/start", h.StartDiagnosis)

	req := httptest.NewRequest(http.MethodPost, "/admin/service-orders/so1/diagnosis/start", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	flowRepo.AssertExpectations(t)
}

func TestAdminServiceOrderFlowController_SendBudget_Success_WithClaims(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)
	flowRepo := new(repomocks.ServiceOrderFlowRepository)
	svc := serviceorder.NewService(new(repomocks.ClientRepository), new(repomocks.VehicleRepository), new(repomocks.ServiceRepository), new(repomocks.PartRepository), flowRepo)
	h := controllers.NewAdminServiceOrderFlowController(svc)

	flowRepo.On("SendLatestBudget", mock.Anything, "so1", mock.MatchedBy(func(v *string) bool { return v != nil && *v == "subject-1" })).Return(nil).Once()

	r := gin.New()
	r.Use(withClaims("ADMIN"))
	r.POST("/admin/service-orders/:id/budget/send", h.SendBudget)

	req := httptest.NewRequest(http.MethodPost, "/admin/service-orders/so1/budget/send", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusNoContent, w.Code)
	flowRepo.AssertExpectations(t)
}

func TestAdminServiceOrderFlowController_SendBudget_NotFound(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)
	flowRepo := new(repomocks.ServiceOrderFlowRepository)
	svc := serviceorder.NewService(new(repomocks.ClientRepository), new(repomocks.VehicleRepository), new(repomocks.ServiceRepository), new(repomocks.PartRepository), flowRepo)
	h := controllers.NewAdminServiceOrderFlowController(svc)

	flowRepo.On("SendLatestBudget", mock.Anything, "so1", mock.Anything).Return(repository.ErrNotFound).Once()

	r := gin.New()
	r.POST("/admin/service-orders/:id/budget/send", h.SendBudget)

	req := httptest.NewRequest(http.MethodPost, "/admin/service-orders/so1/budget/send", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	flowRepo.AssertExpectations(t)
}

func TestAdminServiceOrderFlowController_Finish_Success_WithClaims(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)
	flowRepo := new(repomocks.ServiceOrderFlowRepository)
	svc := serviceorder.NewService(new(repomocks.ClientRepository), new(repomocks.VehicleRepository), new(repomocks.ServiceRepository), new(repomocks.PartRepository), flowRepo)
	h := controllers.NewAdminServiceOrderFlowController(svc)

	flowRepo.On("Finish", mock.Anything, "so1", mock.MatchedBy(func(v *string) bool { return v != nil && *v == "subject-1" })).Return(nil).Once()

	r := gin.New()
	r.Use(withClaims("ADMIN"))
	r.POST("/admin/service-orders/:id/finish", h.Finish)

	req := httptest.NewRequest(http.MethodPost, "/admin/service-orders/so1/finish", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusNoContent, w.Code)
	flowRepo.AssertExpectations(t)
}

func TestAdminServiceOrderFlowController_Finish_InternalError(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)
	flowRepo := new(repomocks.ServiceOrderFlowRepository)
	svc := serviceorder.NewService(new(repomocks.ClientRepository), new(repomocks.VehicleRepository), new(repomocks.ServiceRepository), new(repomocks.PartRepository), flowRepo)
	h := controllers.NewAdminServiceOrderFlowController(svc)

	flowRepo.On("Finish", mock.Anything, "so1", mock.Anything).Return(assert.AnError).Once()

	r := gin.New()
	r.POST("/admin/service-orders/:id/finish", h.Finish)

	req := httptest.NewRequest(http.MethodPost, "/admin/service-orders/so1/finish", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	flowRepo.AssertExpectations(t)
}

func TestAdminServiceOrderFlowController_Deliver_Success_WithClaims(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)
	flowRepo := new(repomocks.ServiceOrderFlowRepository)
	svc := serviceorder.NewService(new(repomocks.ClientRepository), new(repomocks.VehicleRepository), new(repomocks.ServiceRepository), new(repomocks.PartRepository), flowRepo)
	h := controllers.NewAdminServiceOrderFlowController(svc)

	flowRepo.On("Deliver", mock.Anything, "so1", mock.MatchedBy(func(v *string) bool { return v != nil && *v == "subject-1" })).Return(nil).Once()

	r := gin.New()
	r.Use(withClaims("ADMIN"))
	r.POST("/admin/service-orders/:id/deliver", h.Deliver)

	req := httptest.NewRequest(http.MethodPost, "/admin/service-orders/so1/deliver", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusNoContent, w.Code)
	flowRepo.AssertExpectations(t)
}

func TestAdminServiceOrderFlowController_Deliver_NotFound(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)
	flowRepo := new(repomocks.ServiceOrderFlowRepository)
	svc := serviceorder.NewService(new(repomocks.ClientRepository), new(repomocks.VehicleRepository), new(repomocks.ServiceRepository), new(repomocks.PartRepository), flowRepo)
	h := controllers.NewAdminServiceOrderFlowController(svc)

	flowRepo.On("Deliver", mock.Anything, "so1", mock.Anything).Return(repository.ErrNotFound).Once()

	r := gin.New()
	r.POST("/admin/service-orders/:id/deliver", h.Deliver)

	req := httptest.NewRequest(http.MethodPost, "/admin/service-orders/so1/deliver", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	flowRepo.AssertExpectations(t)
}
