package config_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	config2 "github.com/soat-architecture/tech-challenge-project/internal/config"
	"github.com/stretchr/testify/require"
)

func TestNewRateLimitMiddleware_Disabled_AllowsRequests(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mw, err := config2.NewRateLimitMiddleware(config2.RateLimitConfig{Enabled: false})
	require.NoError(t, err)

	r := gin.New()
	r.Use(mw)
	r.GET("/ping", func(c *gin.Context) { c.Status(http.StatusOK) })

	rec1 := httptest.NewRecorder()
	req1 := httptest.NewRequest(http.MethodGet, "/ping", nil)
	req1.RemoteAddr = "1.2.3.4:1234"
	r.ServeHTTP(rec1, req1)
	require.Equal(t, http.StatusOK, rec1.Code)

	rec2 := httptest.NewRecorder()
	req2 := httptest.NewRequest(http.MethodGet, "/ping", nil)
	req2.RemoteAddr = "1.2.3.4:1234"
	r.ServeHTTP(rec2, req2)
	require.Equal(t, http.StatusOK, rec2.Code)
}

func TestNewRateLimitMiddleware_InvalidConfig(t *testing.T) {
	_, err := config2.NewRateLimitMiddleware(config2.RateLimitConfig{Enabled: true, RPS: 0, Burst: 1})
	require.ErrorIs(t, err, config2.ErrInvalidRateLimitConfig)

	_, err = config2.NewRateLimitMiddleware(config2.RateLimitConfig{Enabled: true, RPS: 1, Burst: 0})
	require.ErrorIs(t, err, config2.ErrInvalidRateLimitConfig)
}

func TestRateLimitMiddleware_BlocksAfterBurst(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Extremely low refill rate to keep the test deterministic.
	mw, err := config2.NewRateLimitMiddleware(config2.RateLimitConfig{
		Enabled: true,
		RPS:     0.000001,
		Burst:   1,
		TTL:     1 * time.Minute,
	})
	require.NoError(t, err)

	r := gin.New()
	r.Use(mw)
	r.GET("/ping", func(c *gin.Context) { c.Status(http.StatusOK) })

	rec1 := httptest.NewRecorder()
	req1 := httptest.NewRequest(http.MethodGet, "/ping", nil)
	req1.RemoteAddr = "10.0.0.1:1111"
	r.ServeHTTP(rec1, req1)
	require.Equal(t, http.StatusOK, rec1.Code)

	rec2 := httptest.NewRecorder()
	req2 := httptest.NewRequest(http.MethodGet, "/ping", nil)
	req2.RemoteAddr = "10.0.0.1:2222"
	r.ServeHTTP(rec2, req2)
	require.Equal(t, http.StatusTooManyRequests, rec2.Code)
}

func TestRateLimitMiddleware_IsPerClientIP(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mw, err := config2.NewRateLimitMiddleware(config2.RateLimitConfig{
		Enabled: true,
		RPS:     0.000001,
		Burst:   1,
		TTL:     1 * time.Minute,
	})
	require.NoError(t, err)

	r := gin.New()
	r.Use(mw)
	r.GET("/ping", func(c *gin.Context) { c.Status(http.StatusOK) })

	// IP1 uses its only token.
	rec1 := httptest.NewRecorder()
	req1 := httptest.NewRequest(http.MethodGet, "/ping", nil)
	req1.RemoteAddr = "10.0.0.1:1111"
	r.ServeHTTP(rec1, req1)
	require.Equal(t, http.StatusOK, rec1.Code)

	// IP2 has its own bucket.
	rec2 := httptest.NewRecorder()
	req2 := httptest.NewRequest(http.MethodGet, "/ping", nil)
	req2.RemoteAddr = "10.0.0.2:1111"
	r.ServeHTTP(rec2, req2)
	require.Equal(t, http.StatusOK, rec2.Code)

	// IP1 is blocked now.
	rec3 := httptest.NewRecorder()
	req3 := httptest.NewRequest(http.MethodGet, "/ping", nil)
	req3.RemoteAddr = "10.0.0.1:3333"
	r.ServeHTTP(rec3, req3)
	require.Equal(t, http.StatusTooManyRequests, rec3.Code)
}
