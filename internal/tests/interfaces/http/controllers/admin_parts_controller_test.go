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
	"github.com/soat-architecture/tech-challenge-project/internal/domain/part"
	"github.com/soat-architecture/tech-challenge-project/internal/interfaces/http/controllers"
	"github.com/soat-architecture/tech-challenge-project/internal/interfaces/repository"
	repomocks "github.com/soat-architecture/tech-challenge-project/internal/tests/interfaces/repository/mocks"
)

func TestAdminPartsController_List_Success(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)
	repo := new(repomocks.PartRepository)
	svc := admin.NewPartService(repo)
	h := controllers.NewAdminPartsController(svc)

	repo.On("List", mock.Anything, 50, 0).Return([]part.Part{{
		ID:             "p1",
		SKU:            "SKU-1",
		Name:           "Filtro",
		UnitPriceCents: 5000,
		StockQuantity:  10,
		Active:         true,
	}}, nil).Once()

	r := gin.New()
	r.GET("/admin/parts", h.List)

	req := httptest.NewRequest(http.MethodGet, "/admin/parts", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"id":"p1"`)
	repo.AssertExpectations(t)
}

func TestAdminPartsController_List_InternalError(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)
	repo := new(repomocks.PartRepository)
	svc := admin.NewPartService(repo)
	h := controllers.NewAdminPartsController(svc)

	repo.On("List", mock.Anything, 50, 0).Return(([]part.Part)(nil), assert.AnError).Once()

	r := gin.New()
	r.GET("/admin/parts", h.List)

	req := httptest.NewRequest(http.MethodGet, "/admin/parts", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	repo.AssertExpectations(t)
}

func TestAdminPartsController_Create_BadRequest_InvalidBody(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)
	repo := new(repomocks.PartRepository)
	svc := admin.NewPartService(repo)
	h := controllers.NewAdminPartsController(svc)

	r := gin.New()
	r.POST("/admin/parts", h.Create)

	req := httptest.NewRequest(http.MethodPost, "/admin/parts", bytes.NewBufferString(`{`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAdminPartsController_Create_Success(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)
	repo := new(repomocks.PartRepository)
	svc := admin.NewPartService(repo)
	h := controllers.NewAdminPartsController(svc)

	repo.On("Create", mock.Anything, mock.AnythingOfType("*part.Part")).Return(nil).Once()

	r := gin.New()
	r.POST("/admin/parts", h.Create)

	req := httptest.NewRequest(http.MethodPost, "/admin/parts", bytes.NewBufferString(`{"sku":"SKU-1","name":"Filtro","unit_price_cents":5000,"stock_quantity":10,"active":true}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusCreated, w.Code)
	var out map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &out))
	require.NotEmpty(t, out["id"])
	assert.Equal(t, "SKU-1", out["sku"])
	repo.AssertExpectations(t)
}

func TestAdminPartsController_Create_BadRequest_InvalidInput(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)
	repo := new(repomocks.PartRepository)
	svc := admin.NewPartService(repo)
	h := controllers.NewAdminPartsController(svc)

	r := gin.New()
	r.POST("/admin/parts", h.Create)

	req := httptest.NewRequest(http.MethodPost, "/admin/parts", bytes.NewBufferString(`{"sku":"SKU-1","name":"Filtro","unit_price_cents":-1,"stock_quantity":10,"active":true}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAdminPartsController_Get_NotFound(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)
	repo := new(repomocks.PartRepository)
	svc := admin.NewPartService(repo)
	h := controllers.NewAdminPartsController(svc)

	repo.On("FindByID", mock.Anything, "p1").Return((*part.Part)(nil), repository.ErrNotFound).Once()

	r := gin.New()
	r.GET("/admin/parts/:id", h.Get)

	req := httptest.NewRequest(http.MethodGet, "/admin/parts/p1", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	repo.AssertExpectations(t)
}

func TestAdminPartsController_Update_Success(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)
	repo := new(repomocks.PartRepository)
	svc := admin.NewPartService(repo)
	h := controllers.NewAdminPartsController(svc)

	repo.On("Update", mock.Anything, mock.AnythingOfType("*part.Part")).Return(nil).Once()
	repo.On("FindByID", mock.Anything, "p1").Return(&part.Part{
		ID:             "p1",
		SKU:            "SKU-1",
		Name:           "Filtro Updated",
		UnitPriceCents: 6000,
		StockQuantity:  10,
		Active:         true,
	}, nil).Once()

	r := gin.New()
	r.PUT("/admin/parts/:id", h.Update)

	req := httptest.NewRequest(http.MethodPut, "/admin/parts/p1", bytes.NewBufferString(`{"sku":"SKU-1","name":"Filtro Updated","unit_price_cents":6000,"stock_quantity":10,"active":true}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"id":"p1"`)
	repo.AssertExpectations(t)
}

func TestAdminPartsController_Update_NotFound(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)
	repo := new(repomocks.PartRepository)
	svc := admin.NewPartService(repo)
	h := controllers.NewAdminPartsController(svc)

	repo.On("Update", mock.Anything, mock.AnythingOfType("*part.Part")).Return(repository.ErrNotFound).Once()

	r := gin.New()
	r.PUT("/admin/parts/:id", h.Update)

	req := httptest.NewRequest(http.MethodPut, "/admin/parts/p1", bytes.NewBufferString(`{"sku":"SKU-1","name":"Filtro Updated","unit_price_cents":6000,"stock_quantity":10,"active":true}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	repo.AssertExpectations(t)
}

func TestAdminPartsController_Delete_Success(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)
	repo := new(repomocks.PartRepository)
	svc := admin.NewPartService(repo)
	h := controllers.NewAdminPartsController(svc)

	repo.On("Delete", mock.Anything, "p1").Return(nil).Once()

	r := gin.New()
	r.DELETE("/admin/parts/:id", h.Delete)

	req := httptest.NewRequest(http.MethodDelete, "/admin/parts/p1", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusNoContent, w.Code)
	repo.AssertExpectations(t)
}

func TestAdminPartsController_Delete_NotFound(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)
	repo := new(repomocks.PartRepository)
	svc := admin.NewPartService(repo)
	h := controllers.NewAdminPartsController(svc)

	repo.On("Delete", mock.Anything, "p1").Return(repository.ErrNotFound).Once()

	r := gin.New()
	r.DELETE("/admin/parts/:id", h.Delete)

	req := httptest.NewRequest(http.MethodDelete, "/admin/parts/p1", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	repo.AssertExpectations(t)
}

func TestAdminPartsController_AdjustStock_NotFound(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)
	repo := new(repomocks.PartRepository)
	svc := admin.NewPartService(repo)
	h := controllers.NewAdminPartsController(svc)

	repo.On(
		"AdjustStock",
		mock.Anything,
		"p1",
		part.StockMovementType("IN"),
		5,
		(*string)(nil),
		mock.MatchedBy(func(v *string) bool { return v != nil && *v == "subject-1" }),
	).Return((*part.Part)(nil), repository.ErrNotFound).Once()

	r := gin.New()
	r.Use(withClaims("ADMIN"))
	r.POST("/admin/parts/:id/stock-movements", h.AdjustStock)

	req := httptest.NewRequest(http.MethodPost, "/admin/parts/p1/stock-movements", bytes.NewBufferString(`{"movement_type":"IN","quantity":5}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	repo.AssertExpectations(t)
}

func TestAdminPartsController_AdjustStock_Success(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)
	repo := new(repomocks.PartRepository)
	svc := admin.NewPartService(repo)
	h := controllers.NewAdminPartsController(svc)

	repo.On(
		"AdjustStock",
		mock.Anything,
		"p1",
		part.StockMovementType("IN"),
		5,
		(*string)(nil),
		mock.MatchedBy(func(v *string) bool { return v != nil && *v == "subject-1" }),
	).Return(&part.Part{
		ID:             "p1",
		SKU:            "SKU-1",
		Name:           "Filtro",
		UnitPriceCents: 5000,
		StockQuantity:  15,
		Active:         true,
	}, nil).Once()

	r := gin.New()
	r.Use(withClaims("ADMIN"))
	r.POST("/admin/parts/:id/stock-movements", h.AdjustStock)

	req := httptest.NewRequest(http.MethodPost, "/admin/parts/p1/stock-movements", bytes.NewBufferString(`{"movement_type":"IN","quantity":5}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"id":"p1"`)
	repo.AssertExpectations(t)
}

func TestAdminPartsController_AdjustStock_Conflict(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)
	repo := new(repomocks.PartRepository)
	svc := admin.NewPartService(repo)
	h := controllers.NewAdminPartsController(svc)

	repo.On(
		"AdjustStock",
		mock.Anything,
		"p1",
		part.StockMovementType("OUT"),
		5,
		(*string)(nil),
		mock.MatchedBy(func(v *string) bool { return v != nil && *v == "subject-1" }),
	).Return((*part.Part)(nil), repository.ErrConflict).Once()

	r := gin.New()
	r.Use(withClaims("ADMIN"))
	r.POST("/admin/parts/:id/stock-movements", h.AdjustStock)

	req := httptest.NewRequest(http.MethodPost, "/admin/parts/p1/stock-movements", bytes.NewBufferString(`{"movement_type":"OUT","quantity":5}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusConflict, w.Code)
	repo.AssertExpectations(t)
}

func TestAdminPartsController_AdjustStock_BadRequest_GenericError(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)
	repo := new(repomocks.PartRepository)
	svc := admin.NewPartService(repo)
	h := controllers.NewAdminPartsController(svc)

	repo.On(
		"AdjustStock",
		mock.Anything,
		"p1",
		part.StockMovementType("OUT"),
		5,
		(*string)(nil),
		mock.Anything,
	).Return((*part.Part)(nil), assert.AnError).Once()

	r := gin.New()
	r.Use(withClaims("ADMIN"))
	r.POST("/admin/parts/:id/stock-movements", h.AdjustStock)

	req := httptest.NewRequest(http.MethodPost, "/admin/parts/p1/stock-movements", bytes.NewBufferString(`{"movement_type":"OUT","quantity":5}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusBadRequest, w.Code)
	repo.AssertExpectations(t)
}
