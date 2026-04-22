package config

import (
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

var ErrInvalidCORSConfig = errors.New("invalid cors config")

type CORSConfig struct {
	Enabled bool

	AllowAllOrigins bool

	AllowOrigins []string

	AllowMethods     []string
	AllowHeaders     []string
	ExposeHeaders    []string
	AllowCredentials bool
	MaxAge           time.Duration
}

func DefaultCORSConfig() CORSConfig {
	return CORSConfig{
		Enabled:         true,
		AllowAllOrigins: true,
		AllowOrigins:    nil,
		AllowMethods:    []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodOptions},
		AllowHeaders:    []string{"Authorization", "Content-Type"},
		ExposeHeaders:   nil,

		AllowCredentials: false,
		MaxAge:           12 * time.Hour,
	}
}

func NewCORSMiddleware(cfg CORSConfig) (gin.HandlerFunc, error) {
	if !cfg.Enabled {
		return func(c *gin.Context) { c.Next() }, nil
	}

	if cfg.MaxAge < 0 {
		return nil, ErrInvalidCORSConfig
	}
	if !cfg.AllowAllOrigins && len(cfg.AllowOrigins) == 0 {
		return nil, ErrInvalidCORSConfig
	}

	allowMethods := strings.Join(nonEmpty(cfg.AllowMethods), ", ")
	allowHeaders := strings.Join(nonEmpty(cfg.AllowHeaders), ", ")
	exposeHeaders := strings.Join(nonEmpty(cfg.ExposeHeaders), ", ")
	allowedOrigins := make(map[string]struct{}, len(cfg.AllowOrigins))
	for _, o := range cfg.AllowOrigins {
		o = strings.TrimSpace(o)
		if o == "" {
			continue
		}
		allowedOrigins[o] = struct{}{}
	}

	return func(c *gin.Context) {
		origin := strings.TrimSpace(c.GetHeader("Origin"))
		if origin == "" {
			c.Next()
			return
		}

		allowOrigin, ok := corsAllowOrigin(cfg, origin, allowedOrigins)
		if !ok {
			c.AbortWithStatus(http.StatusForbidden)
			return
		}

		if allowOrigin != "" {
			c.Header("Access-Control-Allow-Origin", allowOrigin)
			if allowOrigin != "*" {
				c.Header("Vary", "Origin")
			}
		}
		if cfg.AllowCredentials {
			c.Header("Access-Control-Allow-Credentials", "true")
		}
		if allowMethods != "" {
			c.Header("Access-Control-Allow-Methods", allowMethods)
		}
		if allowHeaders != "" {
			c.Header("Access-Control-Allow-Headers", allowHeaders)
		}
		if exposeHeaders != "" {
			c.Header("Access-Control-Expose-Headers", exposeHeaders)
		}
		if cfg.MaxAge > 0 {
			c.Header("Access-Control-Max-Age", maxAgeSeconds(cfg.MaxAge))
		}

		if c.Request.Method == http.MethodOptions && c.GetHeader("Access-Control-Request-Method") != "" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}, nil
}

func corsAllowOrigin(cfg CORSConfig, origin string, allowed map[string]struct{}) (string, bool) {
	if cfg.AllowAllOrigins {
		if cfg.AllowCredentials {
			return origin, true
		}
		return "*", true
	}

	if _, ok := allowed[origin]; ok {
		return origin, true
	}
	return "", false
}

func nonEmpty(in []string) []string {
	out := make([]string, 0, len(in))
	for _, s := range in {
		s = strings.TrimSpace(s)
		if s == "" {
			continue
		}
		out = append(out, s)
	}
	return out
}

func maxAgeSeconds(d time.Duration) string {
	sec := int64(d.Seconds())
	if sec < 0 {
		sec = 0
	}
	return strconv.FormatInt(sec, 10)
}
