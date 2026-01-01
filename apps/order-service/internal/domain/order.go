package domain

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrOrderNotFound = errors.New("order not found")
	ErrInvalidOrder  = errors.New("invalid order data")
)

type OrderStatus string

const (
	OrderStatusPending   OrderStatus = "PENDING"
	OrderStatusPaid      OrderStatus = "PAID"
	OrderStatusCancelled OrderStatus = "CANCELLED"
)

type Order struct {
	ID          uuid.UUID
	UserID      uuid.UUID
	Status      OrderStatus
	TotalAmount int64
	Items       []OrderItem
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type OrderItem struct {
	ProductID uuid.UUID
	Quantity  int32
	Price     int64
}

func NewOrder(userID uuid.UUID, items []OrderItem) (*Order, error) {
	if len(items) == 0 {
		return nil, ErrInvalidOrder
	}

	total := int64(0)
	for _, item := range items {
		total += item.Price * int64(item.Quantity)
	}

	return &Order{
		ID:          uuid.New(),
		UserID:      userID,
		Status:      OrderStatusPending,
		TotalAmount: total,
		Items:       items,
		CreatedAt:   time.Now().UTC(),
		UpdatedAt:   time.Now().UTC(),
	}, nil
}
