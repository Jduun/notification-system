package messaging

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/google/uuid"

	"notification_system/config"
	"notification_system/internal/entities"
	"notification_system/internal/repositories"
	"notification_system/pkg/database"
)

type Sender struct {
	producer         *kafka.Producer
	topic            string
	notificationRepo repositories.NotificationRepository
	cfg              *config.Config
}

func NewSender(topic string, cfg *config.Config, db *database.PostgresDatabase) *Sender {
	producer, err := kafka.NewProducer(&kafka.ConfigMap{
		"bootstrap.servers": fmt.Sprintf("localhost:%d", cfg.KafkaClientPort),
		"acks":              "all",
	})
	if err != nil {
		slog.Error("error connecting to kafka", slog.Any("error", err))
		panic("failed to connect kafka")
	}
	notificationRepo := repositories.NewNotificationPostgresRepository(db)
	return &Sender{
		producer:         producer,
		topic:            topic,
		notificationRepo: notificationRepo,
		cfg:              cfg,
	}
}

func (s *Sender) SendMessages(messages [][]byte) error {
	for _, message := range messages {
		err := s.producer.Produce(&kafka.Message{
			TopicPartition: kafka.TopicPartition{Topic: &s.topic, Partition: kafka.PartitionAny},
			Value:          message,
		}, nil)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *Sender) StartProcessNotifications(ctx context.Context, handlePeriod time.Duration) {
	const op = "messaging.sender.StartProcessNotifications"
	log := slog.With(slog.String("op", op))

	ticker := time.NewTicker(handlePeriod)

	go func() {
		for {
			select {
			case <-ctx.Done():
				log.Info("stopping event processing")
				return
			case <-ticker.C:
			}
			const limit = 50
			notifications, err := s.notificationRepo.GetNewNotifications(ctx, limit)
			if err != nil {
				log.Error("failed to get new notifications", slog.Any("error", err))
				continue
			}
			notificationsBytes := make([][]byte, len(notifications))
			ids := make([]uuid.UUID, len(notifications))
			for i, notification := range notifications {
				ids[i] = notification.ID
				notificationBytes, err := json.Marshal(notification)
				if err != nil {
					log.Error("failed to convert notification", slog.Any("error", err))
					continue
				}
				notificationsBytes[i] = notificationBytes
			}
			err = s.SendMessages(notificationsBytes)
			if err != nil {
				log.Error("failed to enqueue kafka message", slog.Any("error", err))
			}
			err = s.notificationRepo.UpdateNotificationsStatus(ctx, ids, entities.StatusInQueue)
			if err != nil {
				log.Error("failed to update notification statuses", slog.Any("error", err))
			}
		}
	}()
}

func (s *Sender) Close() {
	s.producer.Close()
}
