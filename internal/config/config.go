package config

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

type Config struct {
	Port        int
	DatabaseURL string

	JWTSecret        string
	JWTIssuer        string
	JWTAudience      string
	JWTExpiryMinutes int

	CORS      CORSConfig
	RateLimit RateLimitConfig
}

func Load() (Config, error) {
	_ = godotenv.Load()
	viper.SetDefault("PORT", "8080")
	viper.SetDefault("JWT_ISSUER", "tech-challenge-project")
	viper.SetDefault("JWT_AUDIENCE", "")
	viper.SetDefault("JWT_EXPIRY_MINUTES", "60")
	viper.SetDefault("CORS_ENABLED", true)
	viper.SetDefault("CORS_ALLOW_CREDENTIALS", false)
	viper.SetDefault("CORS_ALLOWED_ORIGINS", "")
	viper.SetDefault("RATE_LIMIT_ENABLED", true)
	viper.SetDefault("RATE_LIMIT_RPS", 10.0)
	viper.SetDefault("RATE_LIMIT_BURST", 20)
	viper.SetDefault("RATE_LIMIT_TTL_SECONDS", int((10 * time.Minute).Seconds()))
	viper.SetDefault("RATE_LIMIT_CLEANUP_INTERVAL_SECONDS", int((1 * time.Minute).Seconds()))
	viper.AutomaticEnv()

	portStr := viper.GetString("PORT")
	port, err := strconv.Atoi(portStr)
	if err != nil || port < 1 || port > 65535 {
		return Config{}, fmt.Errorf("invalid PORT %q", portStr)
	}

	jwtExpiryStr := viper.GetString("JWT_EXPIRY_MINUTES")
	jwtExpiryMinutes, err := strconv.Atoi(jwtExpiryStr)
	if err != nil || jwtExpiryMinutes < 1 {
		return Config{}, fmt.Errorf("invalid JWT_EXPIRY_MINUTES %q", jwtExpiryStr)
	}

	corsEnabled := viper.GetBool("CORS_ENABLED")
	corsAllowCredentials := viper.GetBool("CORS_ALLOW_CREDENTIALS")
	corsAllowedOriginsStr := viper.GetString("CORS_ALLOWED_ORIGINS")
	corsAllowedOrigins := parseCommaList(corsAllowedOriginsStr)
	corsCfg := DefaultCORSConfig()
	corsCfg.Enabled = corsEnabled
	corsCfg.AllowCredentials = corsAllowCredentials
	if len(corsAllowedOrigins) > 0 {
		corsCfg.AllowAllOrigins = false
		corsCfg.AllowOrigins = corsAllowedOrigins
	}

	rateLimitEnabled := viper.GetBool("RATE_LIMIT_ENABLED")
	rateLimitRPS := viper.GetFloat64("RATE_LIMIT_RPS")
	rateLimitBurst := viper.GetInt("RATE_LIMIT_BURST")
	rateLimitTTLSeconds := viper.GetInt("RATE_LIMIT_TTL_SECONDS")
	rateLimitCleanupIntervalSeconds := viper.GetInt("RATE_LIMIT_CLEANUP_INTERVAL_SECONDS")
	if rateLimitEnabled {
		if rateLimitRPS <= 0 {
			return Config{}, fmt.Errorf("invalid RATE_LIMIT_RPS %v", rateLimitRPS)
		}
		if rateLimitBurst < 1 {
			return Config{}, fmt.Errorf("invalid RATE_LIMIT_BURST %d", rateLimitBurst)
		}
		if rateLimitTTLSeconds < 1 {
			return Config{}, fmt.Errorf("invalid RATE_LIMIT_TTL_SECONDS %d", rateLimitTTLSeconds)
		}
		if rateLimitCleanupIntervalSeconds < 1 {
			return Config{}, fmt.Errorf("invalid RATE_LIMIT_CLEANUP_INTERVAL_SECONDS %d", rateLimitCleanupIntervalSeconds)
		}
	}

	return Config{
		Port:             port,
		DatabaseURL:      viper.GetString("DATABASE_URL"),
		JWTSecret:        viper.GetString("JWT_SECRET"),
		JWTIssuer:        viper.GetString("JWT_ISSUER"),
		JWTAudience:      viper.GetString("JWT_AUDIENCE"),
		JWTExpiryMinutes: jwtExpiryMinutes,
		CORS:             corsCfg,
		RateLimit: RateLimitConfig{
			Enabled:         rateLimitEnabled,
			RPS:             rateLimitRPS,
			Burst:           rateLimitBurst,
			TTL:             time.Duration(rateLimitTTLSeconds) * time.Second,
			CleanupInterval: time.Duration(rateLimitCleanupIntervalSeconds) * time.Second,
		},
	}, nil
}

func parseCommaList(value string) []string {
	var out []string
	for _, p := range strings.Split(value, ",") {
		if p = strings.TrimSpace(p); p != "" {
			out = append(out, p)
		}
	}
	return out
}
