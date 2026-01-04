package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"order-platform/apps/payment-service/internal/domain"
	"order-platform/apps/payment-service/internal/port"

	"github.com/lib/pq"
)

type PaymentRepository struct {
	db *sql.DB
}

// Проверка интерфейса
var _ port.PaymentRepository = (*PaymentRepository)(nil)

func NewPaymentRepository(db *sql.DB) *PaymentRepository {
	return &PaymentRepository{db: db}
}

// Create сохраняет платеж и событие Outbox в одной транзакции
func (r *PaymentRepository) Create(ctx context.Context, payment *domain.Payment) error {
	// 1. Начинаем транзакцию
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback() // Откат, если не будет Commit

	// 2. Сохраняем сам Платеж
	queryPayment := `
		INSERT INTO payments (order_id, user_id, amount, status, created_at)
		VALUES ($1, $2, $3, $4, NOW())
		RETURNING id, created_at
	`
	err = tx.QueryRowContext(ctx, queryPayment,
		payment.OrderID,
		payment.UserID,
		payment.Amount,
		payment.Status,
	).Scan(&payment.ID, &payment.CreatedAt)

	if err != nil {
		return err
	}

	// 3. Подготавливаем событие для Outbox
	// Используем inline-структуру или DTO, как тебе удобнее.
	// Тут важно, что мы отправляем статус PAID.
	eventPayload := map[string]interface{}{
		"order_id":   payment.OrderID,
		"status":     "PAID",
		"payment_id": payment.ID,
	}

	payloadBytes, err := json.Marshal(eventPayload)
	if err != nil {
		return fmt.Errorf("failed to marshal outbox event: %w", err)
	}

	// 4. Сохраняем в Outbox (в той же транзакции)
	queryOutbox := `
		INSERT INTO payment_outbox (topic, payload)
		VALUES ($1, $2)
	`
	// Топик указываем жестко или берем из конфига (пока жестко payment-events)
	_, err = tx.ExecContext(ctx, queryOutbox, "payment-events", payloadBytes)
	if err != nil {
		return err
	}

	// 5. Фиксируем транзакцию
	return tx.Commit()
}

func (r *PaymentRepository) IsDuplicateError(err error) bool {
	if pqErr, ok := err.(*pq.Error); ok {
		return pqErr.Code == "23505"
	}
	return false
}
