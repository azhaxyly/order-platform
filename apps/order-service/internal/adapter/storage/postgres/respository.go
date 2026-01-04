package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"order-platform/apps/order-service/internal/domain"

	"github.com/google/uuid"
)

type OrderRepository struct {
	db *sql.DB
}

type OutboxEvent struct {
	ID        uuid.UUID
	EventType string
	Payload   []byte
}

func NewOrderRepository(db *sql.DB) *OrderRepository {
	return &OrderRepository{db: db}
}

func (r *OrderRepository) Save(ctx context.Context, order *domain.Order) error {

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	defer tx.Rollback()

	queryOrder := `
		INSERT INTO orders (id, user_id, status, total_amount, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	_, err = tx.ExecContext(ctx, queryOrder,
		order.ID,
		order.UserID,
		order.Status,
		order.TotalAmount,
		order.CreatedAt,
		order.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to insert order: %w", err)
	}

	eventPayload := map[string]interface{}{
		"order_id":     order.ID,
		"user_id":      order.UserID,
		"total_amount": order.TotalAmount,
		"status":       order.Status,
		"items":        order.Items,
	}

	payloadBytes, err := json.Marshal(eventPayload)
	if err != nil {
		return fmt.Errorf("failed to marshal event payload: %w", err)
	}

	queryOutbox := `
		INSERT INTO outbox (id, aggregate_id, event_type, payload, created_at)
		VALUES ($1, $2, $3, $4, $5)
	`
	_, err = tx.ExecContext(ctx, queryOutbox,
		uuid.New(),
		order.ID,
		"OrderCreated",
		payloadBytes,
		time.Now().UTC(),
	)
	if err != nil {
		return fmt.Errorf("failed to insert outbox event: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (r *OrderRepository) GetByID(ctx context.Context, id string) (*domain.Order, error) {
	return nil, nil
}

func (r *OrderRepository) GetUnpublishedEvents(ctx context.Context, limit int) ([]OutboxEvent, error) {
	query := `
		SELECT id, event_type, payload
		FROM outbox
		WHERE published_at IS NULL
		ORDER BY created_at ASC
		LIMIT $1
	`
	rows, err := r.db.QueryContext(ctx, query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []OutboxEvent
	for rows.Next() {
		var e OutboxEvent
		if err := rows.Scan(&e.ID, &e.EventType, &e.Payload); err != nil {
			return nil, err
		}
		events = append(events, e)
	}
	return events, nil
}

func (r *OrderRepository) MarkEventAsPublished(ctx context.Context, eventID uuid.UUID) error {
	query := `
		UPDATE outbox
		SET published_at = $1
		WHERE id = $2
	`
	_, err := r.db.ExecContext(ctx, query, time.Now().UTC(), eventID)
	return err
}

func (r *OrderRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status domain.OrderStatus) error {
	query := `
		UPDATE orders 
		SET status = $1, updated_at = $2
		WHERE id = $3
	`
	// Используем time.Now().UTC()
	_, err := r.db.ExecContext(ctx, query, status, time.Now().UTC(), id)
	if err != nil {
		return fmt.Errorf("failed to update order status: %w", err)
	}
	return nil
}
