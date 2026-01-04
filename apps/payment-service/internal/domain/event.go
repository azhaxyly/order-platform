package domain

// OrderCreatedEvent отображает JSON структуру из топика order-events
type OrderCreatedEvent struct {
	OrderID     string `json:"order_id"`
	UserID      string `json:"user_id"`
	Status      string `json:"status"`
	TotalAmount int64  `json:"total_amount"` // В копейках/центах, чтобы избежать float
}

// PaymentProcessedEvent - событие, которое мы отправляем, когда оплата прошла
type PaymentProcessedEvent struct {
	OrderID string `json:"order_id"`
	Status  string `json:"status"` // "PAID", "FAILED"
}
