package messaging

import (
	"context"
	"encoding/json"
	"log/slog"
	"time"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/google/uuid"

	"notification_system/config"
	"notification_system/internal/entities"
	"notification_system/internal/repositories"
	"notification_system/pkg/database"
)

type NotificationSender struct {
	producer         *kafka.Producer
	notificationRepo repositories.NotificationRepository
	cfg              *config.Config
}

func NewNotificationSender(cfg *config.Config, db *database.PostgresDatabase) *NotificationSender {
	const op = "messaging.sender.NewNotificationSender"
	log := slog.With(slog.String("op", op))

	producer, err := kafka.NewProducer(&kafka.ConfigMap{
		"bootstrap.servers": "kafka:9092",
		"acks":              "all",
	})
	if err != nil {
		log.Error("error connecting to kafka", slog.Any("error", err))
		panic("failed to connect kafka")
	}
	notificationRepo := repositories.NewNotificationPostgresRepository(db)
	return &NotificationSender{
		producer:         producer,
		notificationRepo: notificationRepo,
		cfg:              cfg,
	}
}

func (s *NotificationSender) SendNotificationsToKafka(messages [][]byte) error {
	const op = "messaging.sender.SendNotificationsToKafka"
	log := slog.With(slog.String("op", op))

	for _, message := range messages {
		err := s.producer.Produce(&kafka.Message{
			TopicPartition: kafka.TopicPartition{Topic: &s.cfg.NotificationTopicName, Partition: kafka.PartitionAny},
			Value:          message,
		}, nil)
		if err != nil {
			return err
		}
	}
	count := len(messages)
	if count != 0 {
		log.Info("successfully send notifications to kafka", slog.Int("count", count))
	}
	return nil
}

func (s *NotificationSender) StartProcessNotifications(ctx context.Context, handlePeriod time.Duration) {
	const op = "messaging.sender.StartProcessNotifications"
	log := slog.With(slog.String("op", op))

	ticker := time.NewTicker(handlePeriod)

	go func() {
		for {
			select {
			case <-ctx.Done():
				log.Info("stopping sender notification processing")
				return
			case <-ticker.C:
			}
			limit := s.cfg.MaxBatchSize
			notifications, err := s.notificationRepo.GetNewNotifications(ctx, limit)
			if err != nil {
				log.Error("failed to get new notifications", slog.Any("error", err))
				continue
			}
			notificationsBytes := make([][]byte, len(notifications))
			ids := make([]uuid.UUID, len(notifications))
			for i, notification := range notifications {
				ids[i] = notification.ID
				notificationBytes, err := json.Marshal(*notification)
				if err != nil {
					log.Error("failed to convert notification", slog.Any("error", err))
					continue
				}
				notificationsBytes[i] = notificationBytes
			}
			err = s.SendNotificationsToKafka(notificationsBytes)
			if err != nil {
				log.Error("failed to enqueue kafka message", slog.Any("error", err))
			} else {
				err = s.notificationRepo.UpdateNotificationsStatus(ctx, ids, entities.StatusInQueue)
				if err != nil {
					log.Error("failed to update notification statuses", slog.Any("error", err))
				}
			}
		}
	}()
}

func (s *NotificationSender) Close() {
	s.producer.Close()
}
