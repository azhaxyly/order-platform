package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	pb "order-platform/api/order"
	grpcHandler "order-platform/apps/order-service/internal/adapter/handler/grpc"
	kafkaConsumer "order-platform/apps/order-service/internal/adapter/handler/kafka" // Алиас для нового пакета
	"order-platform/apps/order-service/internal/adapter/storage/postgres"
	"order-platform/apps/order-service/internal/config"
	"order-platform/apps/order-service/internal/service"
	"order-platform/pkg/kafka"
	pkgPostgres "order-platform/pkg/postgres"

	"github.com/IBM/sarama"
	"github.com/joho/godotenv"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	if err := godotenv.Load(".env"); err != nil {
		log.Println("Info: .env file not found")
	}
	cfg := config.Load()

	// 1. DB
	fmt.Printf("Connecting to DB at: ...%s\n", cfg.DBUrl[len(cfg.DBUrl)-20:])
	db, err := pkgPostgres.Connect(context.Background(), cfg.DBUrl)
	if err != nil {
		log.Fatalf("Failed to connect to DB: %v", err)
	}
	defer db.Close()

	// 2. Kafka Producer (для отправки заказов)
	producer, err := kafka.NewProducer(cfg.KafkaBrokers)
	if err != nil {
		log.Fatalf("Failed to connect to Kafka Producer: %v", err)
	}
	defer producer.Close()

	// 3. Kafka Consumer (НОВОЕ: для чтения оплат)
	saramaConfig := sarama.NewConfig()
	saramaConfig.Version = sarama.V3_0_0_0
	saramaConfig.Consumer.Return.Errors = true
	saramaConfig.Consumer.Offsets.Initial = sarama.OffsetOldest

	consumerGroup, err := sarama.NewConsumerGroup(cfg.KafkaBrokers, cfg.KafkaGroupID, saramaConfig)
	if err != nil {
		log.Fatalf("Failed to create Kafka Consumer Group: %v", err)
	}
	defer consumerGroup.Close()

	// 4. Wiring
	orderRepo := postgres.NewOrderRepository(db)
	orderService := service.NewOrderService(orderRepo)
	orderHandler := grpcHandler.NewOrderGrpcServer(orderService)
	outboxProcessor := service.NewOutboxProcessor(orderRepo, producer)

	// Хендлер для входящих сообщений об оплате
	paymentConsumerHandler := kafkaConsumer.NewConsumerGroupHandler(orderRepo)

	// 5. Run Background Tasks
	ctx, cancel := context.WithCancel(context.Background())
	// defer cancel() // Убираем отсюда, вызовем вручную при шатдауне

	// Outbox Processor
	go func() {
		fmt.Println("Outbox Processor started")
		outboxProcessor.Start(ctx)
	}()

	// Kafka Consumer Loop (НОВОЕ)
	go func() {
		fmt.Printf("Kafka Consumer started (listening %s)\n", cfg.KafkaTopic)
		for {
			if err := consumerGroup.Consume(ctx, []string{cfg.KafkaTopic}, paymentConsumerHandler); err != nil {
				log.Printf("Kafka consume error: %v", err)
			}
			if ctx.Err() != nil {
				return
			}
		}
	}()

	// 6. gRPC Server
	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", cfg.GRPCPort))
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterOrderServiceServer(grpcServer, orderHandler)
	reflection.Register(grpcServer)

	// 7. Graceful Shutdown
	// Запускаем сервер в горутине, чтобы main поток ждал сигнала
	go func() {
		fmt.Printf("Order Service started on port %s\n", cfg.GRPCPort)
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("Failed to serve: %v", err)
		}
	}()

	// Ждем сигнала завершения
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	fmt.Println("Shutting down Order Service...")
	grpcServer.GracefulStop()
	cancel() // Останавливаем консьюмеры и процессоры
	fmt.Println("Order Service stopped")
}
