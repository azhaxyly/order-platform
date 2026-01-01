package main

import (
	"context"
	"fmt"
	"log"
	"net"

	pb "order-platform/api/order"
	grpcHandler "order-platform/apps/order-service/internal/adapter/handler/grpc"
	"order-platform/apps/order-service/internal/adapter/storage/postgres"
	"order-platform/apps/order-service/internal/config"
	"order-platform/apps/order-service/internal/service"
	"order-platform/pkg/kafka"
	pkgPostgres "order-platform/pkg/postgres"

	"github.com/joho/godotenv"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	if err := godotenv.Load(".env"); err != nil {
		log.Println("Info: .env file not found, using default config values")
	}
	cfg := config.Load()

	fmt.Printf("Connecting to DB at: ...%s\n", cfg.DBUrl[len(cfg.DBUrl)-20:])
	db, err := pkgPostgres.Connect(context.Background(), cfg.DBUrl)
	if err != nil {
		log.Fatalf("Failed to connect to DB: %v", err)
	}
	defer db.Close()

	kafkaBrokers := []string{"127.0.0.1:9092"}
	producer, err := kafka.NewProducer(kafkaBrokers)
	if err != nil {
		log.Fatalf("Failed to connect to Kafka: %v", err)
	}
	defer producer.Close()
	fmt.Println("Connected to Kafka")

	orderRepo := postgres.NewOrderRepository(db)
	orderService := service.NewOrderService(orderRepo)
	orderHandler := grpcHandler.NewOrderGrpcServer(orderService)

	processor := service.NewOutboxProcessor(orderRepo, producer)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		fmt.Println("Outbox Processor started")
		processor.Start(ctx)
	}()

	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", cfg.GRPCPort))
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterOrderServiceServer(grpcServer, orderHandler)
	reflection.Register(grpcServer)

	fmt.Printf("Order Service started on port %s\n", cfg.GRPCPort)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
