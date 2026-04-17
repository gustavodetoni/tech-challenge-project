package config

import (
	"errors"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

var ErrInvalidRateLimitConfig = errors.New("invalid rate limit config")

type RateLimitConfig struct {
	Enabled         bool
	RPS             float64
	Burst           int
	TTL             time.Duration
	CleanupInterval time.Duration
}

func DefaultRateLimitConfig() RateLimitConfig {
	return RateLimitConfig{
		Enabled:         true,
		RPS:             10,
		Burst:           20,
		TTL:             10 * time.Minute,
		CleanupInterval: 1 * time.Minute,
	}
}

func NewRateLimitMiddleware(cfg RateLimitConfig) (gin.HandlerFunc, error) {
	if !cfg.Enabled {
		return func(c *gin.Context) { c.Next() }, nil
	}
	if cfg.RPS <= 0 || cfg.Burst < 1 {
		return nil, ErrInvalidRateLimitConfig
	}
	if cfg.TTL <= 0 {
		cfg.TTL = DefaultRateLimitConfig().TTL
	}
	if cfg.CleanupInterval <= 0 {
		cfg.CleanupInterval = DefaultRateLimitConfig().CleanupInterval
	}

	rl := &rateLimiter{
		limit:           rate.Limit(cfg.RPS),
		burst:           cfg.Burst,
		ttl:             cfg.TTL,
		cleanupInterval: cfg.CleanupInterval,
		clients:         make(map[string]*clientLimiter),
		now:             time.Now,
		keyFunc: func(c *gin.Context) string {
			ip := c.ClientIP()
			if ip == "" {
				return "unknown"
			}
			return ip
		},
	}

	return func(c *gin.Context) {
		key := rl.keyFunc(c)
		limiter := rl.get(key)
		if !limiter.Allow() {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"error": "rate limit exceeded"})
			return
		}
		c.Next()
	}, nil
}

type clientLimiter struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

type rateLimiter struct {
	limit           rate.Limit
	burst           int
	ttl             time.Duration
	cleanupInterval time.Duration

	mu          sync.Mutex
	clients     map[string]*clientLimiter
	lastCleanup time.Time

	now     func() time.Time
	keyFunc func(*gin.Context) string
}

func (r *rateLimiter) get(key string) *rate.Limiter {
	now := r.now()

	r.mu.Lock()
	defer r.mu.Unlock()

	if r.lastCleanup.IsZero() {
		r.lastCleanup = now
	}
	if now.Sub(r.lastCleanup) >= r.cleanupInterval {
		for k, v := range r.clients {
			if now.Sub(v.lastSeen) > r.ttl {
				delete(r.clients, k)
			}
		}
		r.lastCleanup = now
	}

	if cl, ok := r.clients[key]; ok {
		cl.lastSeen = now
		return cl.limiter
	}

	limiter := rate.NewLimiter(r.limit, r.burst)
	r.clients[key] = &clientLimiter{limiter: limiter, lastSeen: now}
	return limiter
}
