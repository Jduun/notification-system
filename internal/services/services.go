package services

import (
	"context"
	"time"

	"github.com/google/uuid"

	"notification_system/internal/dto"
)

type NotificationService interface {
	GetNotificationByID(ctx context.Context, id uuid.UUID) (*dto.NotificationResponse, error)
	GetNotifications(ctx context.Context, cursorCreatedAt time.Time, cursorId uuid.UUID, limit int) ([]*dto.NotificationResponse, error)
	GetNotificationsByIDs(ctx context.Context, ids []uuid.UUID) ([]*dto.NotificationResponse, error)
	SendNotification(ctx context.Context, notification *dto.NotificationCreate) (uuid.UUID, error)
	SendNotifications(ctx context.Context, notifications []*dto.NotificationCreate) ([]uuid.UUID, error)
}
