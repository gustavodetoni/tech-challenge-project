package config

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	config2 "github.com/soat-architecture/tech-challenge-project/internal/config"
	"github.com/stretchr/testify/require"
)

func TestNewCORSMiddleware_Disabled_NoHeaders(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mw, err := config2.NewCORSMiddleware(config2.CORSConfig{Enabled: false})
	require.NoError(t, err)

	r := gin.New()
	r.Use(mw)
	r.GET("/ping", func(c *gin.Context) { c.Status(http.StatusOK) })

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	req.Header.Set("Origin", "http://example.com")
	r.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)
	require.Empty(t, rec.Header().Get("Access-Control-Allow-Origin"))
}

func TestNewCORSMiddleware_InvalidConfig(t *testing.T) {
	_, err := config2.NewCORSMiddleware(config2.CORSConfig{Enabled: true, AllowAllOrigins: false, AllowOrigins: nil})
	require.ErrorIs(t, err, config2.ErrInvalidCORSConfig)

	_, err = config2.NewCORSMiddleware(config2.CORSConfig{Enabled: true, AllowAllOrigins: true, MaxAge: -1})
	require.ErrorIs(t, err, config2.ErrInvalidCORSConfig)
}

func TestCORSMiddleware_AllowsAnyOrigin_Defaults(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := config2.DefaultCORSConfig()
	mw, err := config2.NewCORSMiddleware(cfg)
	require.NoError(t, err)

	r := gin.New()
	r.Use(mw)
	r.GET("/ping", func(c *gin.Context) { c.Status(http.StatusOK) })

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	req.Header.Set("Origin", "http://evil.example")
	r.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	require.Equal(t, "*", rec.Header().Get("Access-Control-Allow-Origin"))
	require.NotEmpty(t, rec.Header().Get("Access-Control-Allow-Methods"))
	require.NotEmpty(t, rec.Header().Get("Access-Control-Allow-Headers"))
	require.NotEmpty(t, rec.Header().Get("Access-Control-Max-Age"))
}

func TestCORSMiddleware_Preflight_ReturnsNoContent(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mw, err := config2.NewCORSMiddleware(config2.CORSConfig{
		Enabled:         true,
		AllowAllOrigins: true,
		AllowMethods:    []string{http.MethodGet, http.MethodOptions},
		AllowHeaders:    []string{"Authorization"},
		MaxAge:          1 * time.Minute,
	})
	require.NoError(t, err)

	r := gin.New()
	r.Use(mw)
	r.GET("/ping", func(c *gin.Context) { c.Status(http.StatusOK) })

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodOptions, "/ping", nil)
	req.Header.Set("Origin", "http://example.com")
	req.Header.Set("Access-Control-Request-Method", "GET")
	r.ServeHTTP(rec, req)

	require.Equal(t, http.StatusNoContent, rec.Code)
	require.NotEmpty(t, rec.Header().Get("Access-Control-Allow-Origin"))
}

func TestCORSMiddleware_AllowCredentials_EchoesOrigin(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mw, err := config2.NewCORSMiddleware(config2.CORSConfig{
		Enabled:          true,
		AllowAllOrigins:  true,
		AllowCredentials: true,
		AllowMethods:     []string{http.MethodGet},
		AllowHeaders:     []string{"Authorization"},
		MaxAge:           0,
	})
	require.NoError(t, err)

	r := gin.New()
	r.Use(mw)
	r.GET("/ping", func(c *gin.Context) { c.Status(http.StatusOK) })

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	req.Header.Set("Origin", "http://example.com")
	r.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	require.Equal(t, "http://example.com", rec.Header().Get("Access-Control-Allow-Origin"))
	require.Equal(t, "true", rec.Header().Get("Access-Control-Allow-Credentials"))
	require.Equal(t, "Origin", rec.Header().Get("Vary"))
}

func TestCORSMiddleware_RejectsDisallowedOrigin(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mw, err := config2.NewCORSMiddleware(config2.CORSConfig{
		Enabled:         true,
		AllowAllOrigins: false,
		AllowOrigins:    []string{"http://good.example"},
		AllowMethods:    []string{http.MethodGet},
	})
	require.NoError(t, err)

	r := gin.New()
	r.Use(mw)
	r.GET("/ping", func(c *gin.Context) { c.Status(http.StatusOK) })

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	req.Header.Set("Origin", "http://evil.example")
	r.ServeHTTP(rec, req)

	require.Equal(t, http.StatusForbidden, rec.Code)
}
