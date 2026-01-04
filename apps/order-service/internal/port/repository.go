package port

import (
	"context"
	"order-platform/apps/order-service/internal/adapter/storage/postgres"
	"order-platform/apps/order-service/internal/domain"

	"github.com/google/uuid"
)

type OrderRepository interface {
	Save(ctx context.Context, order *domain.Order) error
	GetUnpublishedEvents(ctx context.Context, limit int) ([]postgres.OutboxEvent, error)
	MarkEventAsPublished(ctx context.Context, eventID uuid.UUID) error
	UpdateStatus(ctx context.Context, id uuid.UUID, status domain.OrderStatus) error
}
