package controllers_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	appVersion "github.com/soat-architecture/tech-challenge-project"
	"github.com/soat-architecture/tech-challenge-project/internal/interfaces/http/controllers"
)

func TestHealthController_Health_Success(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)
	h := controllers.NewHealthController()

	r := gin.New()
	r.GET("/health", h.Health)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	assert.JSONEq(t, fmt.Sprintf(`{"status":"ok","version":%q}`, appVersion.Version), w.Body.String())
}
