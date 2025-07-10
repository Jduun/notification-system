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
	"notification_system/internal/notifiers"
	"notification_system/internal/repositories"
	"notification_system/pkg/database"
)

type NotificationReceiver struct {
	consumer         *kafka.Consumer
	notificationRepo repositories.NotificationRepository
	cfg              *config.Config
}

func NewNotificationReceiver(cfg *config.Config, db *database.PostgresDatabase) *NotificationReceiver {
	const op = "messaging.sender.NewNotificationReceiver"
	log := slog.With(slog.String("op", op))

	consumer, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers": "kafka:9092",
		"group.id":          cfg.ConsumerGroupID,
		"auto.offset.reset": "earliest",
	})
	if err != nil {
		log.Error("error connecting to kafka", slog.Any("error", err))
		panic("failed to connect kafka")
	}
	err = consumer.SubscribeTopics([]string{cfg.NotificationTopicName}, nil)
	if err != nil {
		log.Error("error subscribing to topic", slog.Any("error", err))
		panic("failed to subscribe to topic")
	}
	notificationRepo := repositories.NewNotificationPostgresRepository(db)
	return &NotificationReceiver{
		consumer:         consumer,
		notificationRepo: notificationRepo,
		cfg:              cfg,
	}
}

func (r *NotificationReceiver) StartProcessNotifications(ctx context.Context) {
	const op = "messaging.receiver.StartProcessNotifications"
	log := slog.With(slog.String("op", op))

	gmailNotifier := notifiers.GmailNotifier{
		From: r.cfg.Gmail,
	}

	go func() {
		for {
			select {
			case <-ctx.Done():
				log.Info("stopping receiver notification processing")
				return
			default:
				msg, err := r.consumer.ReadMessage(time.Second)
				if err == nil {
					log.Debug("Got message from kafka",
						slog.String("topic-partition", msg.TopicPartition.String()),
						slog.String("message", string(msg.Value)))
					var notification entities.Notification
					notificationBytes := msg.Value
					err = json.Unmarshal(notificationBytes, &notification)
					if err != nil {
						log.Error("error unmarshalling notification", slog.Any("error", err))
					}
					switch notification.DeliveryType {
					case entities.DeliveryTypeEmail:
						err = gmailNotifier.Notify(notification.Recipient, notification.Content)
						// there may be more notifiers here
					}
					if err == nil {
						log.Info("send notification", slog.Any("notification", notification))
						err = r.notificationRepo.UpdateNotificationsStatus(ctx, []uuid.UUID{notification.ID}, entities.StatusDelivered)
						if err != nil {
							log.Error("cannot update notification status", slog.Any("notification", notification))
						}
					} else {
						log.Error("error sending notification", slog.Any("error", err))
						newStatus := entities.StatusPending
						if notification.Retries > r.cfg.MaxRetries {
							newStatus = entities.StatusFailed
						}
						err = r.notificationRepo.UpdateNotificationsStatus(ctx, []uuid.UUID{notification.ID}, newStatus)
						if err != nil {
							log.Error("cannot update notification status", slog.Any("notification", notification))
						}
					}
					err := r.notificationRepo.UpdateNotificationRetries(ctx, notification.ID, notification.Retries+1)
					if err != nil {
						log.Error("cannot update notification retries", slog.Any("notification", notification))
					}
				} else if !err.(kafka.Error).IsTimeout() {
					log.Error("Consumer error", slog.Any("error", err))
				}
			}
		}
	}()
}

func (r *NotificationReceiver) Close() error {
	err := r.consumer.Close()
	return err
}
