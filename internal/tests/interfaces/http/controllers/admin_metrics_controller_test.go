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
	"github.com/soat-architecture/tech-challenge-project/internal/interfaces/http/controllers"
	repomocks "github.com/soat-architecture/tech-challenge-project/internal/tests/interfaces/repository/mocks"
)

func TestAdminMetricsController_AverageExecutionTime_BadRequest_InvalidFrom(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)
	repo := new(repomocks.ServiceOrderRepository)
	svc := admin.NewServiceOrderService(repo)
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
	svc := admin.NewServiceOrderService(repo)
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
	svc := admin.NewServiceOrderService(repo)
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
