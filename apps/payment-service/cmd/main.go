package main

import (
	"context"
	"database/sql"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/IBM/sarama"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"

	// Импорты
	kafkaAdapter "order-platform/apps/payment-service/internal/adapter/handler/kafka"
	postgresAdapter "order-platform/apps/payment-service/internal/adapter/storage/postgres"
	"order-platform/apps/payment-service/internal/config"
	"order-platform/apps/payment-service/internal/service"

	// Общий пакет
	pkgKafka "order-platform/pkg/kafka"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	if err := godotenv.Load("apps/payment-service/.env"); err != nil {
		log.Println("Info: .env file not found, using default config values")
	}
	cfg := config.Load() // Твой новый config.Load, который сам ищет .env

	// 1. DB Connection
	db, err := sql.Open("postgres", cfg.DBUrl)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	// 2. Init Kafka Producer (из pkg/kafka)
	producer, err := pkgKafka.NewProducer(cfg.KafkaBrokers)
	if err != nil {
		logger.Error("Failed to create producer", slog.String("error", err.Error()))
		os.Exit(1)
	}
	defer producer.Close()

	// 3. Init Services
	repo := postgresAdapter.NewPaymentRepository(db)

	// Передаем продюсера в процессор
	outboxProcessor := service.NewOutboxProcessor(db, producer, logger)
	paymentService := service.NewPaymentService(repo, logger)
	consumerHandler := kafkaAdapter.NewConsumerGroupHandler(paymentService, logger)

	// 4. Init Kafka Consumer Client
	saramaConfig := sarama.NewConfig()
	saramaConfig.Version = sarama.V3_0_0_0
	saramaConfig.Consumer.Return.Errors = true
	saramaConfig.Consumer.Offsets.Initial = sarama.OffsetOldest

	consumerGroup, err := sarama.NewConsumerGroup(cfg.KafkaBrokers, cfg.KafkaGroupID, saramaConfig)
	if err != nil {
		panic(err)
	}

	// 5. Run Background Tasks
	ctx, cancel := context.WithCancel(context.Background())

	// Запуск Outbox Processor (отправка исходящих)
	go outboxProcessor.Start(ctx)

	// Запуск Consumer (чтение входящих)
	go func() {
		for {
			if err := consumerGroup.Consume(ctx, []string{cfg.KafkaTopic}, consumerHandler); err != nil {
				logger.Error("Kafka consume error", slog.String("error", err.Error()))
			}
			if ctx.Err() != nil {
				return
			}
		}
	}()

	logger.Info("Payment Service Started", slog.String("port", cfg.GRPCPort))

	// Graceful Shutdown
	sigterm := make(chan os.Signal, 1)
	signal.Notify(sigterm, syscall.SIGINT, syscall.SIGTERM)
	<-sigterm

	cancel()
	consumerGroup.Close()
	logger.Info("Service Stopped")
}
