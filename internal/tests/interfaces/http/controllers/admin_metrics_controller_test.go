package controllers_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/soat-architecture/tech-challenge-project/internal/application/admin"
	repository "github.com/soat-architecture/tech-challenge-project/internal/application/port"
	"github.com/soat-architecture/tech-challenge-project/internal/interfaces/http/controllers"
	repomocks "github.com/soat-architecture/tech-challenge-project/internal/tests/interfaces/repository/mocks"
)

func TestAdminMetricsController_AverageExecutionTime_BadRequest_InvalidFrom(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)
	repo := new(repomocks.ServiceOrderRepository)
	svc := admin.NewServiceOrderAdminUseCase(repo)
	h := controllers.NewAdminMetricsController(svc)

	r := gin.New()
	r.GET("/admin/metrics/avg-execution-time", h.AverageExecutionTime)

	req := httptest.NewRequest(http.MethodGet, "/admin/metrics/avg-execution-time?from=invalid", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAdminMetricsController_AverageExecutionTime_BadRequest_InvalidTo(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)
	repo := new(repomocks.ServiceOrderRepository)
	svc := admin.NewServiceOrderAdminUseCase(repo)
	h := controllers.NewAdminMetricsController(svc)

	r := gin.New()
	r.GET("/admin/metrics/avg-execution-time", h.AverageExecutionTime)

	req := httptest.NewRequest(http.MethodGet, "/admin/metrics/avg-execution-time?to=invalid", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAdminMetricsController_AverageExecutionTime_Success(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)
	repo := new(repomocks.ServiceOrderRepository)
	svc := admin.NewServiceOrderAdminUseCase(repo)
	h := controllers.NewAdminMetricsController(svc)

	repo.On("AverageExecutionMinutes", mock.Anything, (*time.Time)(nil), (*time.Time)(nil)).Return(12.5, nil).Once()

	r := gin.New()
	r.GET("/admin/metrics/avg-execution-time", h.AverageExecutionTime)

	req := httptest.NewRequest(http.MethodGet, "/admin/metrics/avg-execution-time", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	assert.JSONEq(t, `{"average_minutes":12.5}`, w.Body.String())
	repo.AssertExpectations(t)
}

func TestAdminMetricsController_AverageExecutionTime_ServiceError(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)
	repo := new(repomocks.ServiceOrderRepository)
	svc := admin.NewServiceOrderAdminUseCase(repo)
	h := controllers.NewAdminMetricsController(svc)

	repo.On("AverageExecutionMinutes", mock.Anything, (*time.Time)(nil), (*time.Time)(nil)).
		Return(0.0, assert.AnError).
		Once()

	r := gin.New()
	r.GET("/admin/metrics/avg-execution-time", h.AverageExecutionTime)

	req := httptest.NewRequest(http.MethodGet, "/admin/metrics/avg-execution-time", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusInternalServerError, w.Code)
	repo.AssertExpectations(t)
}

func TestAdminMetricsController_AverageServiceExecutionTime_Success(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)
	repo := new(repomocks.ServiceOrderRepository)
	svc := admin.NewServiceOrderAdminUseCase(repo)
	h := controllers.NewAdminMetricsController(svc)

	repo.On("AverageServiceExecutionMinutes", mock.Anything, (*string)(nil), (*time.Time)(nil), (*time.Time)(nil)).
		Return([]repository.ServiceExecutionAverage{{
			ServiceID:      func() *string { s := "s1"; return &s }(),
			Description:    "Alinhamento",
			AverageMinutes: 30.5,
			SampleCount:    2,
		}}, nil).
		Once()

	r := gin.New()
	r.GET("/admin/metrics/avg-service-execution-time", h.AverageServiceExecutionTime)

	req := httptest.NewRequest(http.MethodGet, "/admin/metrics/avg-service-execution-time", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"items"`)
	assert.Contains(t, w.Body.String(), `"description":"Alinhamento"`)
	repo.AssertExpectations(t)
}

func TestAdminMetricsController_AverageServiceExecutionTime_ServiceError(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)
	repo := new(repomocks.ServiceOrderRepository)
	svc := admin.NewServiceOrderAdminUseCase(repo)
	h := controllers.NewAdminMetricsController(svc)

	repo.On("AverageServiceExecutionMinutes", mock.Anything, (*string)(nil), (*time.Time)(nil), (*time.Time)(nil)).
		Return(nil, assert.AnError).
		Once()

	r := gin.New()
	r.GET("/admin/metrics/avg-service-execution-time", h.AverageServiceExecutionTime)

	req := httptest.NewRequest(http.MethodGet, "/admin/metrics/avg-service-execution-time", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusInternalServerError, w.Code)
	repo.AssertExpectations(t)
}

func TestAdminMetricsController_AverageServiceExecutionTime_BadRequest_InvalidFrom(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)
	repo := new(repomocks.ServiceOrderRepository)
	svc := admin.NewServiceOrderAdminUseCase(repo)
	h := controllers.NewAdminMetricsController(svc)

	r := gin.New()
	r.GET("/admin/metrics/avg-service-execution-time", h.AverageServiceExecutionTime)

	req := httptest.NewRequest(http.MethodGet, "/admin/metrics/avg-service-execution-time?from=invalid", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAdminMetricsController_AverageServiceExecutionTime_FilterByServiceID(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)
	repo := new(repomocks.ServiceOrderRepository)
	svc := admin.NewServiceOrderAdminUseCase(repo)
	h := controllers.NewAdminMetricsController(svc)

	serviceID := "s1"
	repo.On("AverageServiceExecutionMinutes", mock.Anything, &serviceID, (*time.Time)(nil), (*time.Time)(nil)).
		Return([]repository.ServiceExecutionAverage{}, nil).
		Once()

	r := gin.New()
	r.GET("/admin/metrics/avg-service-execution-time", h.AverageServiceExecutionTime)

	req := httptest.NewRequest(http.MethodGet, "/admin/metrics/avg-service-execution-time?service_id=s1", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	repo.AssertExpectations(t)
}
