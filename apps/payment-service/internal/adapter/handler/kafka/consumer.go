package kafka

import (
	"encoding/json"
	"log/slog"

	"order-platform/apps/payment-service/internal/domain"
	"order-platform/apps/payment-service/internal/port"

	"github.com/IBM/sarama"
)

type ConsumerGroupHandler struct {
	service port.PaymentService
	logger  *slog.Logger
}

func NewConsumerGroupHandler(service port.PaymentService, logger *slog.Logger) *ConsumerGroupHandler {
	return &ConsumerGroupHandler{
		service: service,
		logger:  logger,
	}
}

func (h *ConsumerGroupHandler) Setup(_ sarama.ConsumerGroupSession) error   { return nil }
func (h *ConsumerGroupHandler) Cleanup(_ sarama.ConsumerGroupSession) error { return nil }

func (h *ConsumerGroupHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for msg := range claim.Messages() {
		var event domain.OrderCreatedEvent
		if err := json.Unmarshal(msg.Value, &event); err != nil {
			h.logger.Error("Invalid JSON", slog.String("error", err.Error()))
			// Помечаем как обработанное, так как мы никогда не сможем его прочитать
			session.MarkMessage(msg, "")
			continue
		}

		err := h.service.ProcessPayment(session.Context(), event)
		if err != nil {
			h.logger.Error("Process payment failed", slog.String("error", err.Error()))
			// НЕ помечаем сообщение, Kafka отправит его снова (retry logic)
			continue
		}

		session.MarkMessage(msg, "")
	}
	return nil
}
