package entities

import (
	"time"

	"github.com/google/uuid"
)

type Notification struct {
	ID           uuid.UUID  `db:"id"`
	DeliveryType string     `db:"delivery_type"`
	Recipient    string     `db:"recipient"`
	Content      string     `db:"content"`
	Status       string     `db:"status"`
	Retries      uint8      `db:"retries"`
	CreatedAt    time.Time  `db:"created_at"`
	SentAt       *time.Time `db:"sent_at"`
}
