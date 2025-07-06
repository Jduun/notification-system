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
	CreateNotification(ctx context.Context, notification *entities.Notification) error
	CreateNotifications(ctx context.Context, notifications []*entities.Notification) error
	UpdateNotificationsStatus(ctx context.Context, ids []uuid.UUID, status string) error
}
