package port

import (
	"context"
	"order-platform/apps/payment-service/internal/domain"
)

type PaymentRepository interface {
	Create(ctx context.Context, payment *domain.Payment) error
	IsDuplicateError(err error) bool
}
