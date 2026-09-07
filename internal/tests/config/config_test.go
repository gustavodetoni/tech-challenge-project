package config_test

import (
	"testing"
	"time"

	config2 "github.com/soat-architecture/tech-challenge-project/internal/config"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
)

func TestLoad_DefaultsAndEnv(t *testing.T) {
	viper.Reset()
	t.Cleanup(viper.Reset)

	t.Setenv("JWT_SECRET", "secret")
	t.Setenv("DATABASE_URL", "postgres://example")

	cfg, err := config2.Load()
	require.NoError(t, err)
	require.Equal(t, 8080, cfg.Port)
	require.Equal(t, "postgres://example", cfg.DatabaseURL)
	require.Equal(t, "secret", cfg.JWTSecret)
	require.Equal(t, "tech-challenge-project", cfg.JWTIssuer)
	require.Equal(t, "", cfg.JWTAudience)
	require.Equal(t, 60, cfg.JWTExpiryMinutes)
	require.Equal(t, "tech-challenge-auth-lambda", cfg.ClientJWTIssuer)
	require.Equal(t, "", cfg.ClientJWTAudience)
	require.True(t, cfg.CORS.Enabled)
	require.True(t, cfg.CORS.AllowAllOrigins)
	require.False(t, cfg.CORS.AllowCredentials)
	require.Equal(t, 12*time.Hour, cfg.CORS.MaxAge)
	require.Contains(t, cfg.CORS.AllowMethods, "GET")
	require.Contains(t, cfg.CORS.AllowHeaders, "Authorization")
	require.True(t, cfg.RateLimit.Enabled)
	require.Equal(t, 10.0, cfg.RateLimit.RPS)
	require.Equal(t, 20, cfg.RateLimit.Burst)
	require.Equal(t, 10*time.Minute, cfg.RateLimit.TTL)
	require.Equal(t, 1*time.Minute, cfg.RateLimit.CleanupInterval)
	require.Empty(t, cfg.Brevo.APIKey)
	require.Empty(t, cfg.Brevo.SenderEmail)
	require.Equal(t, "Oficina - Pos", cfg.Brevo.SenderName)
}

func TestLoad_InvalidPort(t *testing.T) {
	viper.Reset()
	t.Cleanup(viper.Reset)

	t.Setenv("PORT", "nope")

	_, err := config2.Load()
	require.Error(t, err)
}

func TestLoad_InvalidJWTExpiryMinutes(t *testing.T) {
	viper.Reset()
	t.Cleanup(viper.Reset)

	t.Setenv("JWT_EXPIRY_MINUTES", "0")

	_, err := config2.Load()
	require.Error(t, err)
}

func TestLoad_CORSAllowedOrigins_DisablesAllowAllOrigins(t *testing.T) {
	viper.Reset()
	t.Cleanup(viper.Reset)

	t.Setenv("JWT_SECRET", "secret")
	t.Setenv("DATABASE_URL", "postgres://example")
	t.Setenv("CORS_ALLOWED_ORIGINS", "https://front.example, https://front2.example")

	cfg, err := config2.Load()
	require.NoError(t, err)
	require.True(t, cfg.CORS.Enabled)
	require.False(t, cfg.CORS.AllowAllOrigins)
	require.Equal(t, []string{"https://front.example", "https://front2.example"}, cfg.CORS.AllowOrigins)
}

func TestLoad_BrevoConfig(t *testing.T) {
	viper.Reset()
	t.Cleanup(viper.Reset)

	t.Setenv("JWT_SECRET", "secret")
	t.Setenv("DATABASE_URL", "postgres://example")
	t.Setenv("BREVO_API_KEY", "key")
	t.Setenv("BREVO_SENDER_EMAIL", "sender@example.com")
	t.Setenv("BREVO_SENDER_NAME", "Oficina Teste")

	cfg, err := config2.Load()
	require.NoError(t, err)
	require.Equal(t, "key", cfg.Brevo.APIKey)
	require.Equal(t, "sender@example.com", cfg.Brevo.SenderEmail)
	require.Equal(t, "Oficina Teste", cfg.Brevo.SenderName)
}

func TestLoad_BrevoPartialConfig(t *testing.T) {
	viper.Reset()
	t.Cleanup(viper.Reset)

	t.Setenv("JWT_SECRET", "secret")
	t.Setenv("DATABASE_URL", "postgres://example")
	t.Setenv("BREVO_API_KEY", "key")

	_, err := config2.Load()
	require.Error(t, err)
}
