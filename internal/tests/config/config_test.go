package config_test

import (
	"testing"

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
