package repositories

import (
	"context"
	"time"

	"github.com/google/uuid"

	"notification_system/internal/entities"
)

type NotificationRepository interface {
	GetNotificationByID(ctx context.Context, id uuid.UUID) (*entities.Notification, error)
	GetNotifications(ctx context.Context, cursorCreatedAt time.Time, cursorId uuid.UUID, limit int) ([]*entities.Notification, error)
	GetNotificationsByIDs(ctx context.Context, ids []uuid.UUID) ([]*entities.Notification, error)
	CreateNotification(ctx context.Context, notification *entities.Notification) error
	CreateNotifications(ctx context.Context, notifications []*entities.Notification) error
}
