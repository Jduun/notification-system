package repositories

import (
	"context"

	"github.com/google/uuid"

	"notification_system/internal/entities"
)

type NotificationRepository interface {
	GetNotificationByID(ctx context.Context, id uuid.UUID) (*entities.Notification, error)
	GetNewNotifications(ctx context.Context, limit int) ([]*entities.Notification, error)
	GetNotificationsByIDs(ctx context.Context, ids []uuid.UUID) ([]*entities.Notification, error)
	CreateNotifications(ctx context.Context, notifications []*entities.Notification) error
	UpdateNotificationsStatus(ctx context.Context, ids []uuid.UUID, status string) error
	UpdateNotificationRetries(ctx context.Context, id uuid.UUID, retries uint8) error
}
