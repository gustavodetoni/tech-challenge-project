package config

import (
	"fmt"
	"strconv"

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
}

func Load() (Config, error) {
	_ = godotenv.Load()
	viper.SetDefault("PORT", "8080")
	viper.SetDefault("JWT_ISSUER", "tech-challenge-project")
	viper.SetDefault("JWT_AUDIENCE", "")
	viper.SetDefault("JWT_EXPIRY_MINUTES", "60")
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

	return Config{
		Port:             port,
		DatabaseURL:      viper.GetString("DATABASE_URL"),
		JWTSecret:        viper.GetString("JWT_SECRET"),
		JWTIssuer:        viper.GetString("JWT_ISSUER"),
		JWTAudience:      viper.GetString("JWT_AUDIENCE"),
		JWTExpiryMinutes: jwtExpiryMinutes,
	}, nil
}
