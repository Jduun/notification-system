package services

import (
	"context"

	"github.com/google/uuid"

	"notification_system/internal/dto"
)

type NotificationService interface {
	GetNotificationByID(ctx context.Context, id uuid.UUID) (*dto.Notification, error)
	GetNewNotifications(ctx context.Context, limit uint) ([]*dto.Notification, error)
	GetNotificationsByIDs(ctx context.Context, ids []uuid.UUID) ([]*dto.Notification, error)
	CreateNotifications(ctx context.Context, notifications []*dto.NotificationCreate) ([]uuid.UUID, error)
}
