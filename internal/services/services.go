package services

import (
	"context"

	"github.com/google/uuid"

	"notification_system/internal/dto"
)

type NotificationService interface {
	GetNotificationByID(ctx context.Context, id uuid.UUID) (*dto.NotificationResponse, error)
	GetNewNotifications(ctx context.Context, limit int) ([]*dto.NotificationResponse, error)
	GetNotificationsByIDs(ctx context.Context, ids []uuid.UUID) ([]*dto.NotificationResponse, error)
	CreateNotifications(ctx context.Context, notifications []*dto.NotificationCreate) ([]uuid.UUID, error)
}
