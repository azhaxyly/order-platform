package service

import (
	"context"
	"fmt"

	"order-platform/apps/order-service/internal/domain"
	"order-platform/apps/order-service/internal/port"

	"github.com/google/uuid"
)

type OrderService struct {
	repo port.OrderRepository
}

func NewOrderService(repo port.OrderRepository) *OrderService {
	return &OrderService{repo: repo}
}

func (s *OrderService) CreateOrder(ctx context.Context, userID uuid.UUID, items []domain.OrderItem) (*domain.Order, error) {

	order, err := domain.NewOrder(userID, items)
	if err != nil {
		return nil, fmt.Errorf("failed to create order domain: %w", err)
	}

	if err := s.repo.Save(ctx, order); err != nil {
		return nil, fmt.Errorf("failed to save order: %w", err)
	}

	return order, nil
}
