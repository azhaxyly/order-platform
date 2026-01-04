package domain

import (
	"time"
)

// Payment - основная сущность платежа
type Payment struct {
	ID        string    `json:"id"`
	OrderID   string    `json:"order_id"` // Unique Key
	UserID    string    `json:"user_id"`
	Amount    int64     `json:"amount"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}
