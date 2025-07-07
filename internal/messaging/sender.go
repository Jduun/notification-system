package messaging

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/google/uuid"

	"notification_system/internal/entities"
	"notification_system/internal/repositories"
	"notification_system/pkg/database"
)

type Sender struct {
	producer         *kafka.Producer
	topic            string
	notificationRepo repositories.NotificationRepository
}

func NewSender(topic string, db *database.PostgresDatabase) *Sender {
	const op = "messaging.sender.NewSender"
	log := slog.With(slog.String("op", op))

	producer, err := kafka.NewProducer(&kafka.ConfigMap{
		"bootstrap.servers": fmt.Sprintf("kafka:9092"),
		"acks":              "all",
	})
	if err != nil {
		log.Error("error connecting to kafka", slog.Any("error", err))
		panic("failed to connect kafka")
	}
	notificationRepo := repositories.NewNotificationPostgresRepository(db)
	return &Sender{
		producer:         producer,
		topic:            topic,
		notificationRepo: notificationRepo,
	}
}

func (s *Sender) SendMessages(messages [][]byte) error {
	const op = "messaging.sender.SendMessages"
	log := slog.With(slog.String("op", op))

	for _, message := range messages {
		err := s.producer.Produce(&kafka.Message{
			TopicPartition: kafka.TopicPartition{Topic: &s.topic, Partition: kafka.PartitionAny},
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

func (s *Sender) StartProcessNotifications(ctx context.Context, handlePeriod time.Duration) {
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
				notificationBytes, err := json.Marshal(*notification)
				if err != nil {
					log.Error("failed to convert notification", slog.Any("error", err))
					continue
				}
				notificationsBytes[i] = notificationBytes
			}
			err = s.SendMessages(notificationsBytes)
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

func (s *Sender) Close() {
	s.producer.Close()
}
