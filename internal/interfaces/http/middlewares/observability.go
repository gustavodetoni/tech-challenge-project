package middlewares

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"
	"unicode"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const (
	HeaderCorrelationID       = "X-Correlation-ID"
	ContextCorrelationIDKey   = "request.correlation_id"
	maxCorrelationIDLength    = 128
	unknownRouteTemplateValue = "unknown"
)

type correlationIDContextKey struct{}

func RequestLogger(logger *slog.Logger) gin.HandlerFunc {
	logger = ensureLogger(logger)

	return func(c *gin.Context) {
		start := time.Now()
		correlationID := correlationIDFromHeader(c.GetHeader(HeaderCorrelationID))
		if correlationID == "" {
			correlationID = uuid.NewString()
		}

		c.Set(ContextCorrelationIDKey, correlationID)
		c.Writer.Header().Set(HeaderCorrelationID, correlationID)
		c.Request = c.Request.WithContext(context.WithValue(c.Request.Context(), correlationIDContextKey{}, correlationID))

		c.Next()

		latency := time.Since(start)
		status := c.Writer.Status()
		route := c.FullPath()
		if route == "" {
			route = unknownRouteTemplateValue
		}

		attrs := []slog.Attr{
			slog.String("correlation_id", correlationID),
			slog.String("method", c.Request.Method),
			slog.String("path", c.Request.URL.Path),
			slog.String("route", route),
			slog.Int("status", status),
			slog.Int64("latency_ms", latency.Milliseconds()),
			slog.String("client_ip", c.ClientIP()),
			slog.String("user_agent", c.Request.UserAgent()),
		}
		if len(c.Errors) > 0 {
			attrs = append(attrs, slog.String("error", c.Errors.String()))
		}

		level := slog.LevelInfo
		if status >= http.StatusInternalServerError {
			level = slog.LevelError
		} else if status >= http.StatusBadRequest {
			level = slog.LevelWarn
		}

		logger.LogAttrs(c.Request.Context(), level, "http_request", attrs...)
	}
}

func RequestRecovery(logger *slog.Logger) gin.HandlerFunc {
	logger = ensureLogger(logger)

	return gin.CustomRecoveryWithWriter(io.Discard, func(c *gin.Context, recovered any) {
		correlationID := CorrelationID(c)
		logger.ErrorContext(
			c.Request.Context(),
			"http_panic",
			slog.String("correlation_id", correlationID),
			slog.Any("panic", recovered),
			slog.String("method", c.Request.Method),
			slog.String("path", c.Request.URL.Path),
		)

		c.AbortWithStatusJSON(http.StatusInternalServerError, ErrorResponse(c, "internal error"))
	})
}

func ErrorResponse(c *gin.Context, message string) gin.H {
	resp := gin.H{"error": message}
	if correlationID := CorrelationID(c); correlationID != "" {
		resp["correlation_id"] = correlationID
	}
	return resp
}

func CorrelationID(c *gin.Context) string {
	if c == nil {
		return ""
	}
	if v, ok := c.Get(ContextCorrelationIDKey); ok {
		if id, ok := v.(string); ok {
			return id
		}
	}
	if c.Request == nil {
		return ""
	}
	return CorrelationIDFromContext(c.Request.Context())
}

func CorrelationIDFromContext(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	id, _ := ctx.Value(correlationIDContextKey{}).(string)
	return id
}

func ensureLogger(logger *slog.Logger) *slog.Logger {
	if logger != nil {
		return logger
	}
	return slog.New(slog.NewJSONHandler(os.Stdout, nil))
}

func correlationIDFromHeader(value string) string {
	id := strings.TrimSpace(value)
	if id == "" || len(id) > maxCorrelationIDLength {
		return ""
	}
	for _, r := range id {
		if unicode.IsControl(r) || unicode.IsSpace(r) {
			return ""
		}
	}
	return id
}
