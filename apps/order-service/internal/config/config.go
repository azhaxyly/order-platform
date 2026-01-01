package config

import (
	"os"
)

type Config struct {
	AppEnv   string
	GRPCPort string
	DBUrl    string
}

func Load() *Config {
	return &Config{
		AppEnv:   getEnv("APP_ENV", "development"),
		GRPCPort: getEnv("GRPC_PORT", "50051"),
		DBUrl:    getEnv("DATABASE_URL", "postgres://order_user:order_pass@127.0.0.1:5435/order_service?sslmode=disable"),
	}
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}
