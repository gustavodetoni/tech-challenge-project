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
	"github.com/soat-architecture/tech-challenge-project/internal/domain/client"
	"github.com/soat-architecture/tech-challenge-project/internal/interfaces/http/controllers"
	repomocks "github.com/soat-architecture/tech-challenge-project/internal/tests/interfaces/repository/mocks"
)

func TestAdminClientsController_List_Success(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)
	repo := new(repomocks.ClientRepository)
	svc := admin.NewClientAdminUseCase(repo)
	h := controllers.NewAdminClientsController(svc)

	repo.On("List", mock.Anything, 50, 0).Return([]client.Client{{
		ID:             "c1",
		DocumentType:   client.DocumentTypeCPF,
		DocumentNumber: "46420082412",
		Name:           "Maria",
	}}, nil).Once()

	r := gin.New()
	r.GET("/admin/clients", h.List)

	req := httptest.NewRequest(http.MethodGet, "/admin/clients", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	assert.JSONEq(t, `[{"id":"c1","document_type":"CPF","document_number":"46420082412","name":"Maria"}]`, w.Body.String())
	repo.AssertExpectations(t)
}

func TestAdminClientsController_List_InternalError(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)
	repo := new(repomocks.ClientRepository)
	svc := admin.NewClientAdminUseCase(repo)
	h := controllers.NewAdminClientsController(svc)

	repo.On("List", mock.Anything, 50, 0).Return(([]client.Client)(nil), assert.AnError).Once()

	r := gin.New()
	r.GET("/admin/clients", h.List)

	req := httptest.NewRequest(http.MethodGet, "/admin/clients", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	repo.AssertExpectations(t)
}

func TestAdminClientsController_Create_BadRequest_InvalidBody(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)
	repo := new(repomocks.ClientRepository)
	svc := admin.NewClientAdminUseCase(repo)
	h := controllers.NewAdminClientsController(svc)

	r := gin.New()
	r.POST("/admin/clients", h.Create)

	req := httptest.NewRequest(http.MethodPost, "/admin/clients", bytes.NewBufferString(`{`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAdminClientsController_Create_BadRequest_InvalidDocument(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)
	repo := new(repomocks.ClientRepository)
	svc := admin.NewClientAdminUseCase(repo)
	h := controllers.NewAdminClientsController(svc)

	r := gin.New()
	r.POST("/admin/clients", h.Create)

	req := httptest.NewRequest(http.MethodPost, "/admin/clients", bytes.NewBufferString(`{"document_type":"CPF","document_number":"123","name":"Maria"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAdminClientsController_Create_Success(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)
	repo := new(repomocks.ClientRepository)
	svc := admin.NewClientAdminUseCase(repo)
	h := controllers.NewAdminClientsController(svc)

	repo.On("Create", mock.Anything, mock.AnythingOfType("*client.Client")).Return(nil).Once()

	r := gin.New()
	r.POST("/admin/clients", h.Create)

	req := httptest.NewRequest(http.MethodPost, "/admin/clients", bytes.NewBufferString(`{"document_type":"CPF","document_number":"46420082412","name":"Maria"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusCreated, w.Code)

	var out map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &out))
	require.NotEmpty(t, out["id"])
	assert.Equal(t, "CPF", out["document_type"])
	assert.Equal(t, "46420082412", out["document_number"])
	assert.Equal(t, "Maria", out["name"])
	repo.AssertExpectations(t)
}

func TestAdminClientsController_Get_NotFound(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)
	repo := new(repomocks.ClientRepository)
	svc := admin.NewClientAdminUseCase(repo)
	h := controllers.NewAdminClientsController(svc)

	repo.On("FindByID", mock.Anything, "c1").Return((*client.Client)(nil), repository.ErrNotFound).Once()

	r := gin.New()
	r.GET("/admin/clients/:id", h.Get)

	req := httptest.NewRequest(http.MethodGet, "/admin/clients/c1", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	repo.AssertExpectations(t)
}

func TestAdminClientsController_Update_BadRequest_InvalidBody(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)
	repo := new(repomocks.ClientRepository)
	svc := admin.NewClientAdminUseCase(repo)
	h := controllers.NewAdminClientsController(svc)

	r := gin.New()
	r.PUT("/admin/clients/:id", h.Update)

	req := httptest.NewRequest(http.MethodPut, "/admin/clients/c1", bytes.NewBufferString(`{`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAdminClientsController_Update_Success(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)
	repo := new(repomocks.ClientRepository)
	svc := admin.NewClientAdminUseCase(repo)
	h := controllers.NewAdminClientsController(svc)

	repo.On("Update", mock.Anything, mock.AnythingOfType("*client.Client")).Return(nil).Once()
	repo.On("FindByID", mock.Anything, "c1").Return(&client.Client{
		ID:             "c1",
		DocumentType:   client.DocumentTypeCPF,
		DocumentNumber: "46420082412",
		Name:           "Maria Updated",
	}, nil).Once()

	r := gin.New()
	r.PUT("/admin/clients/:id", h.Update)

	req := httptest.NewRequest(http.MethodPut, "/admin/clients/c1", bytes.NewBufferString(`{"document_type":"CPF","document_number":"46420082412","name":"Maria Updated"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"id":"c1"`)
	repo.AssertExpectations(t)
}

func TestAdminClientsController_Delete_Success(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)
	repo := new(repomocks.ClientRepository)
	svc := admin.NewClientAdminUseCase(repo)
	h := controllers.NewAdminClientsController(svc)

	repo.On("Delete", mock.Anything, "c1").Return(nil).Once()

	r := gin.New()
	r.DELETE("/admin/clients/:id", h.Delete)

	req := httptest.NewRequest(http.MethodDelete, "/admin/clients/c1", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusNoContent, w.Code)
	repo.AssertExpectations(t)
}

func TestAdminClientsController_Delete_NotFound(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)
	repo := new(repomocks.ClientRepository)
	svc := admin.NewClientAdminUseCase(repo)
	h := controllers.NewAdminClientsController(svc)

	repo.On("Delete", mock.Anything, "c1").Return(repository.ErrNotFound).Once()

	r := gin.New()
	r.DELETE("/admin/clients/:id", h.Delete)

	req := httptest.NewRequest(http.MethodDelete, "/admin/clients/c1", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	repo.AssertExpectations(t)
}
