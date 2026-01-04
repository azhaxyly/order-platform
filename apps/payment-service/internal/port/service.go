package port

import (
	"context"
	"order-platform/apps/payment-service/internal/domain"
)

type PaymentService interface {
	ProcessPayment(ctx context.Context, event domain.OrderCreatedEvent) error
}
