package service

import (
	"context"
	"fmt"
	"log/slog"

	"order-platform/apps/payment-service/internal/domain"
	"order-platform/apps/payment-service/internal/port"
)

type PaymentService struct {
	repo   port.PaymentRepository
	logger *slog.Logger
}

func NewPaymentService(repo port.PaymentRepository, logger *slog.Logger) *PaymentService {
	return &PaymentService{
		repo:   repo,
		logger: logger,
	}
}

func (s *PaymentService) ProcessPayment(ctx context.Context, event domain.OrderCreatedEvent) error {
	s.logger.Info("Processing payment", slog.String("order_id", event.OrderID))

	payment := &domain.Payment{
		OrderID: event.OrderID,
		UserID:  event.UserID,
		Amount:  event.TotalAmount,
		Status:  "SUCCESS", // Упрощено, здесь была бы логика эквайринга
	}

	err := s.repo.Create(ctx, payment)
	if err != nil {
		if s.repo.IsDuplicateError(err) {
			s.logger.Warn("Payment already exists (idempotency)", slog.String("order_id", event.OrderID))
			return nil // Считаем успешным, чтобы сместить оффсет Kafka
		}
		return fmt.Errorf("failed to save payment: %w", err)
	}

	s.logger.Info("Payment processed successfully", slog.String("payment_id", payment.ID))
	return nil
}
