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

	"github.com/soat-architecture/tech-challenge-project/internal/application/admin"
	"github.com/soat-architecture/tech-challenge-project/internal/domain/vehicle"
	"github.com/soat-architecture/tech-challenge-project/internal/interfaces/http/controllers"
	"github.com/soat-architecture/tech-challenge-project/internal/interfaces/repository"
	repomocks "github.com/soat-architecture/tech-challenge-project/internal/tests/interfaces/repository/mocks"
)

func TestAdminVehiclesController_ListByClient_Success(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)
	repo := new(repomocks.VehicleRepository)
	svc := admin.NewVehicleService(repo)
	h := controllers.NewAdminVehiclesController(svc)

	repo.On("ListByClientID", mock.Anything, "c1", 50, 0).Return([]vehicle.Vehicle{{
		ID:        "v1",
		ClientID:  "c1",
		Plate:     "ABC1D23",
		Brand:     "Fiat",
		Model:     "Uno",
		ModelYear: 2015,
	}}, nil).Once()

	r := gin.New()
	r.GET("/admin/clients/:id/vehicles", h.ListByClient)

	req := httptest.NewRequest(http.MethodGet, "/admin/clients/c1/vehicles", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"id":"v1"`)
	repo.AssertExpectations(t)
}

func TestAdminVehiclesController_ListByClient_InternalError(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)
	repo := new(repomocks.VehicleRepository)
	svc := admin.NewVehicleService(repo)
	h := controllers.NewAdminVehiclesController(svc)

	repo.On("ListByClientID", mock.Anything, "c1", 50, 0).Return(([]vehicle.Vehicle)(nil), assert.AnError).Once()

	r := gin.New()
	r.GET("/admin/clients/:id/vehicles", h.ListByClient)

	req := httptest.NewRequest(http.MethodGet, "/admin/clients/c1/vehicles", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	repo.AssertExpectations(t)
}

func TestAdminVehiclesController_CreateForClient_BadRequest_InvalidBody(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)
	repo := new(repomocks.VehicleRepository)
	svc := admin.NewVehicleService(repo)
	h := controllers.NewAdminVehiclesController(svc)

	r := gin.New()
	r.POST("/admin/clients/:id/vehicles", h.CreateForClient)

	req := httptest.NewRequest(http.MethodPost, "/admin/clients/c1/vehicles", bytes.NewBufferString(`{`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAdminVehiclesController_CreateForClient_BadRequest_InvalidPlate(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)
	repo := new(repomocks.VehicleRepository)
	svc := admin.NewVehicleService(repo)
	h := controllers.NewAdminVehiclesController(svc)

	r := gin.New()
	r.POST("/admin/clients/:id/vehicles", h.CreateForClient)

	req := httptest.NewRequest(http.MethodPost, "/admin/clients/c1/vehicles", bytes.NewBufferString(`{"plate":"NOPE","brand":"Fiat","model":"Uno","model_year":2015}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAdminVehiclesController_CreateForClient_Success(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)
	repo := new(repomocks.VehicleRepository)
	svc := admin.NewVehicleService(repo)
	h := controllers.NewAdminVehiclesController(svc)

	repo.On("Create", mock.Anything, mock.AnythingOfType("*vehicle.Vehicle")).Return(nil).Once()

	r := gin.New()
	r.POST("/admin/clients/:id/vehicles", h.CreateForClient)

	req := httptest.NewRequest(http.MethodPost, "/admin/clients/c1/vehicles", bytes.NewBufferString(`{"plate":"abc1d23","brand":"Fiat","model":"Uno","model_year":2015}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusCreated, w.Code)

	var out map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &out))
	require.NotEmpty(t, out["id"])
	assert.Equal(t, "c1", out["client_id"])
	assert.Equal(t, "ABC1D23", out["plate"])
	repo.AssertExpectations(t)
}

func TestAdminVehiclesController_Get_NotFound(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)
	repo := new(repomocks.VehicleRepository)
	svc := admin.NewVehicleService(repo)
	h := controllers.NewAdminVehiclesController(svc)

	repo.On("FindByID", mock.Anything, "v1").Return((*vehicle.Vehicle)(nil), repository.ErrNotFound).Once()

	r := gin.New()
	r.GET("/admin/vehicles/:id", h.Get)

	req := httptest.NewRequest(http.MethodGet, "/admin/vehicles/v1", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	repo.AssertExpectations(t)
}

func TestAdminVehiclesController_Update_Success(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)
	repo := new(repomocks.VehicleRepository)
	svc := admin.NewVehicleService(repo)
	h := controllers.NewAdminVehiclesController(svc)

	repo.On("FindByID", mock.Anything, "v1").Return(&vehicle.Vehicle{
		ID:        "v1",
		ClientID:  "c1",
		Plate:     "ABC1D23",
		Brand:     "Fiat",
		Model:     "Uno",
		ModelYear: 2015,
	}, nil).Once()
	repo.On("Update", mock.Anything, mock.AnythingOfType("*vehicle.Vehicle")).Return(nil).Once()
	repo.On("FindByID", mock.Anything, "v1").Return(&vehicle.Vehicle{
		ID:        "v1",
		ClientID:  "c1",
		Plate:     "ABC1D23",
		Brand:     "Fiat",
		Model:     "Uno Mille",
		ModelYear: 2015,
	}, nil).Once()

	r := gin.New()
	r.PUT("/admin/vehicles/:id", h.Update)

	req := httptest.NewRequest(http.MethodPut, "/admin/vehicles/v1", bytes.NewBufferString(`{"plate":"ABC1D23","brand":"Fiat","model":"Uno Mille","model_year":2015}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"id":"v1"`)
	repo.AssertExpectations(t)
}

func TestAdminVehiclesController_Update_NotFound_WhenCurrentMissing(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)
	repo := new(repomocks.VehicleRepository)
	svc := admin.NewVehicleService(repo)
	h := controllers.NewAdminVehiclesController(svc)

	repo.On("FindByID", mock.Anything, "v1").Return((*vehicle.Vehicle)(nil), repository.ErrNotFound).Once()

	r := gin.New()
	r.PUT("/admin/vehicles/:id", h.Update)

	req := httptest.NewRequest(http.MethodPut, "/admin/vehicles/v1", bytes.NewBufferString(`{"plate":"ABC1D23","brand":"Fiat","model":"Uno Mille","model_year":2015}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	repo.AssertExpectations(t)
}

func TestAdminVehiclesController_Update_BadRequest_InvalidBody(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)
	repo := new(repomocks.VehicleRepository)
	svc := admin.NewVehicleService(repo)
	h := controllers.NewAdminVehiclesController(svc)

	r := gin.New()
	r.PUT("/admin/vehicles/:id", h.Update)

	req := httptest.NewRequest(http.MethodPut, "/admin/vehicles/v1", bytes.NewBufferString(`{`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAdminVehiclesController_Delete_Success(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)
	repo := new(repomocks.VehicleRepository)
	svc := admin.NewVehicleService(repo)
	h := controllers.NewAdminVehiclesController(svc)

	repo.On("Delete", mock.Anything, "v1").Return(nil).Once()

	r := gin.New()
	r.DELETE("/admin/vehicles/:id", h.Delete)

	req := httptest.NewRequest(http.MethodDelete, "/admin/vehicles/v1", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusNoContent, w.Code)
	repo.AssertExpectations(t)
}

func TestAdminVehiclesController_Delete_NotFound(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)
	repo := new(repomocks.VehicleRepository)
	svc := admin.NewVehicleService(repo)
	h := controllers.NewAdminVehiclesController(svc)

	repo.On("Delete", mock.Anything, "v1").Return(repository.ErrNotFound).Once()

	r := gin.New()
	r.DELETE("/admin/vehicles/:id", h.Delete)

	req := httptest.NewRequest(http.MethodDelete, "/admin/vehicles/v1", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	repo.AssertExpectations(t)
}
