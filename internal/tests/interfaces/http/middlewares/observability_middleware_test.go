package middlewares_test

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/soat-architecture/tech-challenge-project/internal/interfaces/http/middlewares"
	"github.com/soat-architecture/tech-challenge-project/internal/interfaces/http/shared"
)

func TestRequestLogger_GeneratesCorrelationIDAndWritesJSONLog(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)
	var logs bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&logs, nil))

	r := gin.New()
	r.Use(middlewares.RequestLogger(logger))
	r.GET("/health", func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	correlationID := w.Header().Get(middlewares.HeaderCorrelationID)
	require.NotEmpty(t, correlationID)

	entry := decodeFirstLogLine(t, logs.String())
	assert.Equal(t, "http_request", entry["msg"])
	assert.Equal(t, "INFO", entry["level"])
	assert.Equal(t, correlationID, entry["correlation_id"])
	assert.Equal(t, http.MethodGet, entry["method"])
	assert.Equal(t, "/health", entry["path"])
	assert.Equal(t, "/health", entry["route"])
	assert.Equal(t, float64(http.StatusNoContent), entry["status"])
	assert.Contains(t, entry, "latency_ms")
}

func TestRequestLogger_PropagatesIncomingCorrelationID(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)
	var logs bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&logs, nil))

	r := gin.New()
	r.Use(middlewares.RequestLogger(logger))
	r.GET("/client/:id", func(c *gin.Context) {
		assert.Equal(t, "request-123", middlewares.CorrelationID(c))
		assert.Equal(t, "request-123", middlewares.CorrelationIDFromContext(c.Request.Context()))
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/client/abc", nil)
	req.Header.Set(middlewares.HeaderCorrelationID, "request-123")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, "request-123", w.Header().Get(middlewares.HeaderCorrelationID))

	entry := decodeFirstLogLine(t, logs.String())
	assert.Equal(t, "request-123", entry["correlation_id"])
	assert.Equal(t, "/client/:id", entry["route"])
}

func TestRequestLogger_LogsErrorsAndErrorResponseIncludesCorrelationID(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)
	var logs bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&logs, nil))

	r := gin.New()
	r.Use(middlewares.RequestLogger(logger))
	r.GET("/error", func(c *gin.Context) {
		shared.WriteHTTPError(c, http.StatusBadRequest, "invalid request")
	})

	req := httptest.NewRequest(http.MethodGet, "/error", nil)
	req.Header.Set(middlewares.HeaderCorrelationID, "request-456")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusBadRequest, w.Code)
	assert.Equal(t, "request-456", w.Header().Get(middlewares.HeaderCorrelationID))
	assert.JSONEq(t, `{"error":"invalid request","correlation_id":"request-456"}`, w.Body.String())

	entry := decodeFirstLogLine(t, logs.String())
	assert.Equal(t, "WARN", entry["level"])
	assert.Equal(t, float64(http.StatusBadRequest), entry["status"])
	assert.Equal(t, "request-456", entry["correlation_id"])
}

func TestRequestRecovery_LogsPanicAndReturnsCorrelationID(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)
	var logs bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&logs, nil))

	r := gin.New()
	r.Use(middlewares.RequestLogger(logger))
	r.Use(middlewares.RequestRecovery(logger))
	r.GET("/panic", func(c *gin.Context) {
		panic("boom")
	})

	req := httptest.NewRequest(http.MethodGet, "/panic", nil)
	req.Header.Set(middlewares.HeaderCorrelationID, "request-789")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Equal(t, "request-789", w.Header().Get(middlewares.HeaderCorrelationID))
	assert.JSONEq(t, `{"error":"internal error","correlation_id":"request-789"}`, w.Body.String())

	entries := decodeLogLines(t, logs.String())
	require.Len(t, entries, 2)
	assert.Equal(t, "http_panic", entries[0]["msg"])
	assert.Equal(t, "ERROR", entries[0]["level"])
	assert.Equal(t, "request-789", entries[0]["correlation_id"])
	assert.Equal(t, "http_request", entries[1]["msg"])
	assert.Equal(t, "ERROR", entries[1]["level"])
	assert.Equal(t, float64(http.StatusInternalServerError), entries[1]["status"])
}

func decodeFirstLogLine(t *testing.T, logs string) map[string]any {
	t.Helper()

	entries := decodeLogLines(t, logs)
	require.NotEmpty(t, entries)
	return entries[0]
}

func decodeLogLines(t *testing.T, logs string) []map[string]any {
	t.Helper()

	lines := strings.Split(strings.TrimSpace(logs), "\n")
	entries := make([]map[string]any, 0, len(lines))
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		var entry map[string]any
		require.NoError(t, json.Unmarshal([]byte(line), &entry))
		entries = append(entries, entry)
	}
	return entries
}
