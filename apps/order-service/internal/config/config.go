package config

import (
	"os"
	"strings"
)

type Config struct {
	AppEnv       string
	GRPCPort     string
	DBUrl        string
	KafkaBrokers []string
	KafkaTopic   string // Топик, который мы слушаем (payment-events)
	KafkaGroupID string
}

func Load() *Config {
	return &Config{
		AppEnv:   getEnv("APP_ENV", "dev"),
		GRPCPort: getEnv("GRPC_PORT", "50051"),
		DBUrl:    getEnv("DATABASE_URL", "postgres://order_user:order_pass@127.0.0.1:5435/order_service?sslmode=disable"),

		// Новые настройки Kafka
		KafkaBrokers: strings.Split(getEnv("KAFKA_BROKERS", "127.0.0.1:9092"), ","),
		KafkaTopic:   getEnv("KAFKA_PAYMENT_TOPIC", "payment-events"), // Слушаем события платежей
		KafkaGroupID: getEnv("KAFKA_GROUP_ID", "order-service-group"),
	}
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}
