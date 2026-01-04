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
	KafkaTopic   string
	KafkaGroupID string
}

func Load() *Config {
	return &Config{
		AppEnv: getEnv("APP_ENV", "dev"),
		// Порт 50052, чтобы не конфликтовать с Order Service (50051)
		GRPCPort: getEnv("GRPC_PORT", "50052"),
		// Обрати внимание: база данных payment_service
		DBUrl: getEnv("DATABASE_URL", "postgres://postgres:postgres@127.0.0.1:5435/payment_service?sslmode=disable"),

		// Kafka настройки
		KafkaBrokers: strings.Split(getEnv("KAFKA_BROKERS", "127.0.0.1:9092"), ","),
		KafkaTopic:   getEnv("KAFKA_TOPIC", "order-events"),
		KafkaGroupID: getEnv("KAFKA_GROUP_ID", "payment-service-group"),
	}
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}
