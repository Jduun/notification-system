package dto

import (
	"time"

	"github.com/google/uuid"

	"notification_system/internal/entities"
)

type (
	NotificationCreate struct {
		DeliveryType string `json:"delivery_type"`
		Recipient    string `json:"recipient"`
		Content      string `json:"content"`
	}

	Notification struct {
		ID           uuid.UUID  `json:"id"`
		DeliveryType string     `json:"delivery_type"`
		Recipient    string     `json:"recipient"`
		Content      string     `json:"content"`
		Status       string     `json:"status"`
		Retries      uint8      `json:"retries"`
		CreatedAt    time.Time  `json:"created_at"`
		SentAt       *time.Time `json:"sent_at"`
	}
)

func NotificationEntityToDTO(notification *entities.Notification) *Notification {
	return &Notification{
		ID:           notification.ID,
		DeliveryType: notification.DeliveryType,
		Recipient:    notification.Recipient,
		Content:      notification.Content,
		Status:       notification.Status,
		Retries:      notification.Retries,
		CreatedAt:    notification.CreatedAt,
		SentAt:       notification.SentAt,
	}
}

func NotificationEntitiesToDTOs(notifications []*entities.Notification) []*Notification {
	notificationsResponse := make([]*Notification, len(notifications))
	for i, notification := range notifications {
		notificationsResponse[i] = NotificationEntityToDTO(notification)
	}
	return notificationsResponse
}
