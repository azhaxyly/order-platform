package service

import (
	"context"
	"database/sql"
	"log/slog"
	"time"

	// Импортируем общий пакет
	pkgKafka "order-platform/pkg/kafka"
)

type OutboxProcessor struct {
	db       *sql.DB
	producer *pkgKafka.Producer // Используем структуру из pkg
	logger   *slog.Logger
}

func NewOutboxProcessor(db *sql.DB, producer *pkgKafka.Producer, logger *slog.Logger) *OutboxProcessor {
	return &OutboxProcessor{
		db:       db,
		producer: producer,
		logger:   logger,
	}
}

func (p *OutboxProcessor) Start(ctx context.Context) {
	// Поллинг каждые 5 секунд
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			p.processOutbox(ctx)
		}
	}
}

func (p *OutboxProcessor) processOutbox(ctx context.Context) {
	rows, err := p.db.QueryContext(ctx, `
		SELECT id, topic, payload 
		FROM payment_outbox 
		WHERE status = 'PENDING' 
		ORDER BY created_at ASC 
		LIMIT 10
	`)
	if err != nil {
		p.logger.Error("Failed to fetch outbox", slog.String("error", err.Error()))
		return
	}
	defer rows.Close()

	for rows.Next() {
		var id string
		var topic string
		var payload []byte

		if err := rows.Scan(&id, &topic, &payload); err != nil {
			continue
		}

		// Отправка через pkg/kafka
		err := p.producer.SendMessage(topic, id, payload)
		if err != nil {
			p.logger.Error("Failed to send kafka msg", slog.String("id", id), slog.String("error", err.Error()))
			continue
		}

		// Удаление после отправки
		_, err = p.db.ExecContext(ctx, "DELETE FROM payment_outbox WHERE id = $1", id)
		if err != nil {
			p.logger.Error("Failed to delete outbox msg", slog.String("id", id), slog.String("error", err.Error()))
		} else {
			p.logger.Info("Outbox message sent", slog.String("id", id))
		}
	}
}
