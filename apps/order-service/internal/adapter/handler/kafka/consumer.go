package kafka

import (
	"encoding/json"
	"log"

	"github.com/IBM/sarama"
	"github.com/google/uuid"

	"order-platform/apps/order-service/internal/domain"
	"order-platform/apps/order-service/internal/port"
)

type ConsumerGroupHandler struct {
	repo port.OrderRepository
}

func NewConsumerGroupHandler(repo port.OrderRepository) *ConsumerGroupHandler {
	return &ConsumerGroupHandler{repo: repo}
}

func (h *ConsumerGroupHandler) Setup(_ sarama.ConsumerGroupSession) error   { return nil }
func (h *ConsumerGroupHandler) Cleanup(_ sarama.ConsumerGroupSession) error { return nil }

func (h *ConsumerGroupHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for msg := range claim.Messages() {
		log.Printf("Order Service received payment event: %s", string(msg.Value))

		var event domain.PaymentProcessedEvent
		if err := json.Unmarshal(msg.Value, &event); err != nil {
			log.Printf("Error unmarshalling event: %v", err)
			session.MarkMessage(msg, "") // Пропускаем битое сообщение
			continue
		}

		if event.Status == "PAID" {
			orderUUID, err := uuid.Parse(event.OrderID)
			if err != nil {
				log.Printf("Invalid UUID in event: %v", err)
				session.MarkMessage(msg, "")
				continue
			}

			// Обновляем статус в БД
			err = h.repo.UpdateStatus(session.Context(), orderUUID, domain.OrderStatusPaid)
			if err != nil {
				log.Printf("Failed to update order status: %v", err)
				// Не маркаем сообщение, чтобы Kafka прислала его снова (retry)
				continue
			}

			log.Printf("Order %s updated to PAID", event.OrderID)
		}

		session.MarkMessage(msg, "")
	}
	return nil
}
