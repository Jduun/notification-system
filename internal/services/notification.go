package services

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/google/uuid"

	"notification_system/internal/dto"
	"notification_system/internal/entities"
	"notification_system/internal/repositories"
	slogger "notification_system/pkg/logger"
)

type NotificationServiceImpl struct {
	notificationRepo repositories.NotificationRepository
}

func NewNotificationServiceImpl(notificationRepo repositories.NotificationRepository) NotificationService {
	return &NotificationServiceImpl{notificationRepo: notificationRepo}
}

func (s *NotificationServiceImpl) GetNotificationByID(ctx context.Context, id uuid.UUID) (*dto.NotificationResponse, error) {

	notification, err := s.notificationRepo.GetNotificationByID(ctx, id)
	if err != nil {
		if errors.Is(err, repositories.ErrNotFound) {
			return nil, ErrNotificationNotFound
		}
		return nil, ErrCannotGetNotificationByID
	}
	notificationResponse := dto.NotificationEntityToDTO(notification)
	return notificationResponse, nil
}

func (s *NotificationServiceImpl) GetNotifications(ctx context.Context, cursorCreatedAt time.Time, cursorId uuid.UUID, limit int) ([]*dto.NotificationResponse, error) {
	notifications, err := s.notificationRepo.GetNotifications(ctx, cursorCreatedAt, cursorId, limit)
	if err != nil {
		if errors.Is(err, repositories.ErrMaxBatchSizeExceeded) {
			return nil, ErrTooManyRequestedNotifications
		}
		return nil, ErrCannotGetNotifications
	}
	notificationsResponse := dto.NotificationEntitiesToDTOs(notifications)
	return notificationsResponse, nil
}

func (s *NotificationServiceImpl) GetNotificationsByIDs(ctx context.Context, ids []uuid.UUID) ([]*dto.NotificationResponse, error) {
	notifications, err := s.notificationRepo.GetNotificationsByIDs(ctx, ids)
	if err != nil {
		if errors.Is(err, repositories.ErrMaxBatchSizeExceeded) {
			return nil, ErrTooManyRequestedNotifications
		}
		return nil, ErrCannotGetNotificationsByIDs
	}
	notificationsResponse := dto.NotificationEntitiesToDTOs(notifications)
	return notificationsResponse, nil
}

func (s *NotificationServiceImpl) SendNotification(ctx context.Context, notification *dto.NotificationCreate) (uuid.UUID, error) {
	logger := slogger.GetLoggerFromContext(ctx)
	notificationEntity := entities.Notification{
		DeliveryType: notification.DeliveryType,
		Recipient:    notification.Recipient,
		Content:      notification.Content,
	}

	logger.Debug("sending single notification",
		slog.String("recipient", notification.Recipient),
		slog.String("delivery_type", notification.DeliveryType),
	)

	err := s.notificationRepo.CreateNotification(ctx, &notificationEntity)
	if err != nil {
		logger.Error("failed to send notification", slog.Any("error", err))
		return uuid.Nil, ErrCannotSendNotification
	}

	logger.Info("notification sent successfully",
		slog.String("id", notificationEntity.ID.String()),
	)

	return notificationEntity.ID, nil
}

func (s *NotificationServiceImpl) SendNotifications(ctx context.Context, notifications []*dto.NotificationCreate) ([]uuid.UUID, error) {
	logger := slogger.GetLoggerFromContext(ctx)

	logger.Debug("sending notifications",
		slog.Int("count", len(notifications)),
	)

	notificationEntities := make([]*entities.Notification, len(notifications))
	for i, notification := range notifications {
		notificationEntities[i] = &entities.Notification{
			DeliveryType: notification.DeliveryType,
			Recipient:    notification.Recipient,
			Content:      notification.Content,
		}
	}
	err := s.notificationRepo.CreateNotifications(ctx, notificationEntities)
	if err != nil {
		if errors.Is(err, repositories.ErrMaxBatchSizeExceeded) {
			logger.Warn("too many notifications in batch",
				slog.Int("count", len(notifications)),
			)
			return nil, ErrTooManyNotificationsToSend
		}
		logger.Error("failed to send notifications",
			slog.Any("error", err),
		)
		return nil, ErrCannotSendNotifications
	}
	ids := make([]uuid.UUID, len(notifications))
	for i, notification := range notificationEntities {
		ids[i] = notification.ID
	}

	logger.Info("notifications sent successfully",
		slog.Int("count", len(ids)),
	)

	return ids, nil
}
