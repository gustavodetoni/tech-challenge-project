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
	repository "github.com/soat-architecture/tech-challenge-project/internal/application/port"
	"github.com/soat-architecture/tech-challenge-project/internal/domain/service"
	"github.com/soat-architecture/tech-challenge-project/internal/interfaces/http/controllers"
	repomocks "github.com/soat-architecture/tech-challenge-project/internal/tests/interfaces/repository/mocks"
)

func TestAdminServicesController_List_Success(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)
	repo := new(repomocks.ServiceRepository)
	svc := admin.NewServiceCatalogUseCase(repo)
	h := controllers.NewAdminServicesController(svc)

	repo.On("List", mock.Anything, 50, 0).Return([]service.Service{{
		ID:               "s1",
		Name:             "Alinhamento",
		BasePriceCents:   15000,
		EstimatedMinutes: 45,
		Active:           true,
	}}, nil).Once()

	r := gin.New()
	r.GET("/admin/services", h.List)

	req := httptest.NewRequest(http.MethodGet, "/admin/services", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"id":"s1"`)
	repo.AssertExpectations(t)
}

func TestAdminServicesController_List_InternalError(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)
	repo := new(repomocks.ServiceRepository)
	svc := admin.NewServiceCatalogUseCase(repo)
	h := controllers.NewAdminServicesController(svc)

	repo.On("List", mock.Anything, 50, 0).Return(([]service.Service)(nil), assert.AnError).Once()

	r := gin.New()
	r.GET("/admin/services", h.List)

	req := httptest.NewRequest(http.MethodGet, "/admin/services", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	repo.AssertExpectations(t)
}

func TestAdminServicesController_Create_BadRequest_InvalidBody(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)
	repo := new(repomocks.ServiceRepository)
	svc := admin.NewServiceCatalogUseCase(repo)
	h := controllers.NewAdminServicesController(svc)

	r := gin.New()
	r.POST("/admin/services", h.Create)

	req := httptest.NewRequest(http.MethodPost, "/admin/services", bytes.NewBufferString(`{`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAdminServicesController_Create_Success(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)
	repo := new(repomocks.ServiceRepository)
	svc := admin.NewServiceCatalogUseCase(repo)
	h := controllers.NewAdminServicesController(svc)

	repo.On("Create", mock.Anything, mock.AnythingOfType("*service.Service")).Return(nil).Once()

	r := gin.New()
	r.POST("/admin/services", h.Create)

	req := httptest.NewRequest(http.MethodPost, "/admin/services", bytes.NewBufferString(`{"name":"Alinhamento","base_price_cents":15000,"estimated_minutes":45,"active":true}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusCreated, w.Code)
	var out map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &out))
	require.NotEmpty(t, out["id"])
	assert.Equal(t, "Alinhamento", out["name"])
	repo.AssertExpectations(t)
}

func TestAdminServicesController_Create_BadRequest_InvalidInput(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)
	repo := new(repomocks.ServiceRepository)
	svc := admin.NewServiceCatalogUseCase(repo)
	h := controllers.NewAdminServicesController(svc)

	r := gin.New()
	r.POST("/admin/services", h.Create)

	req := httptest.NewRequest(http.MethodPost, "/admin/services", bytes.NewBufferString(`{"name":"Alinhamento","base_price_cents":-1,"estimated_minutes":45,"active":true}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAdminServicesController_Get_NotFound(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)
	repo := new(repomocks.ServiceRepository)
	svc := admin.NewServiceCatalogUseCase(repo)
	h := controllers.NewAdminServicesController(svc)

	repo.On("FindByID", mock.Anything, "s1").Return((*service.Service)(nil), repository.ErrNotFound).Once()

	r := gin.New()
	r.GET("/admin/services/:id", h.Get)

	req := httptest.NewRequest(http.MethodGet, "/admin/services/s1", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	repo.AssertExpectations(t)
}

func TestAdminServicesController_Update_BadRequest_InvalidBody(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)
	repo := new(repomocks.ServiceRepository)
	svc := admin.NewServiceCatalogUseCase(repo)
	h := controllers.NewAdminServicesController(svc)

	r := gin.New()
	r.PUT("/admin/services/:id", h.Update)

	req := httptest.NewRequest(http.MethodPut, "/admin/services/s1", bytes.NewBufferString(`{`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAdminServicesController_Update_NotFound(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)
	repo := new(repomocks.ServiceRepository)
	svc := admin.NewServiceCatalogUseCase(repo)
	h := controllers.NewAdminServicesController(svc)

	repo.On("Update", mock.Anything, mock.AnythingOfType("*service.Service")).Return(repository.ErrNotFound).Once()

	r := gin.New()
	r.PUT("/admin/services/:id", h.Update)

	req := httptest.NewRequest(http.MethodPut, "/admin/services/s1", bytes.NewBufferString(`{"name":"Alinhamento","base_price_cents":15000,"estimated_minutes":45,"active":true}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	repo.AssertExpectations(t)
}

func TestAdminServicesController_Update_Success(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)
	repo := new(repomocks.ServiceRepository)
	svc := admin.NewServiceCatalogUseCase(repo)
	h := controllers.NewAdminServicesController(svc)

	repo.On("Update", mock.Anything, mock.AnythingOfType("*service.Service")).Return(nil).Once()
	repo.On("FindByID", mock.Anything, "s1").Return(&service.Service{
		ID:               "s1",
		Name:             "Alinhamento Updated",
		BasePriceCents:   16000,
		EstimatedMinutes: 50,
		Active:           true,
	}, nil).Once()

	r := gin.New()
	r.PUT("/admin/services/:id", h.Update)

	req := httptest.NewRequest(http.MethodPut, "/admin/services/s1", bytes.NewBufferString(`{"name":"Alinhamento Updated","base_price_cents":16000,"estimated_minutes":50,"active":true}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"id":"s1"`)
	repo.AssertExpectations(t)
}

func TestAdminServicesController_Delete_Success(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)
	repo := new(repomocks.ServiceRepository)
	svc := admin.NewServiceCatalogUseCase(repo)
	h := controllers.NewAdminServicesController(svc)

	repo.On("Delete", mock.Anything, "s1").Return(nil).Once()

	r := gin.New()
	r.DELETE("/admin/services/:id", h.Delete)

	req := httptest.NewRequest(http.MethodDelete, "/admin/services/s1", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusNoContent, w.Code)
	repo.AssertExpectations(t)
}

func TestAdminServicesController_Delete_NotFound(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)
	repo := new(repomocks.ServiceRepository)
	svc := admin.NewServiceCatalogUseCase(repo)
	h := controllers.NewAdminServicesController(svc)

	repo.On("Delete", mock.Anything, "s1").Return(repository.ErrNotFound).Once()

	r := gin.New()
	r.DELETE("/admin/services/:id", h.Delete)

	req := httptest.NewRequest(http.MethodDelete, "/admin/services/s1", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	repo.AssertExpectations(t)
}
