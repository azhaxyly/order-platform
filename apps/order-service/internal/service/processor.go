package service

import (
	"context"
	"log"
	"time"

	"order-platform/apps/order-service/internal/adapter/storage/postgres"
	"order-platform/pkg/kafka"
)

type OutboxProcessor struct {
	repo     *postgres.OrderRepository
	producer *kafka.Producer
	topic    string
}

func NewOutboxProcessor(repo *postgres.OrderRepository, producer *kafka.Producer) *OutboxProcessor {
	return &OutboxProcessor{
		repo:     repo,
		producer: producer,
		topic:    "order-events",
	}
}

func (p *OutboxProcessor) Start(ctx context.Context) {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			p.processEvents(ctx)
		}
	}
}

func (p *OutboxProcessor) processEvents(ctx context.Context) {

	events, err := p.repo.GetUnpublishedEvents(ctx, 10)
	if err != nil {
		log.Printf("Error getting unpublished events: %v", err)
		return
	}

	if len(events) == 0 {
		return
	}

	log.Printf("Found %d unpublished events. Processing...", len(events))

	for _, event := range events {

		err := p.producer.SendMessage(p.topic, event.ID.String(), event.Payload)
		if err != nil {
			log.Printf("Failed to publish event %s: %v", event.ID, err)
			continue
		}

		if err := p.repo.MarkEventAsPublished(ctx, event.ID); err != nil {
			log.Printf("Failed to mark event %s as published: %v", event.ID, err)
		}
	}
}
