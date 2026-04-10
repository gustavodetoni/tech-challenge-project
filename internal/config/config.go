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
}

func Load() (Config, error) {
	_ = godotenv.Load()
	viper.SetDefault("PORT", "8080")
	viper.AutomaticEnv()

	portStr := viper.GetString("PORT")
	port, err := strconv.Atoi(portStr)
	if err != nil || port < 1 || port > 65535 {
		return Config{}, fmt.Errorf("invalid PORT %q", portStr)
	}

	return Config{
		Port:        port,
		DatabaseURL: viper.GetString("DATABASE_URL"),
	}, nil
}
