package domain

// PaymentProcessedEvent - структура события, которое мы ждем от Payment Service
type PaymentProcessedEvent struct {
	OrderID   string `json:"order_id"`
	PaymentID string `json:"payment_id"`
	Status    string `json:"status"` // "PAID"
}
