package repositories

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"

	"notification_system/config"
	"notification_system/internal/entities"
	"notification_system/pkg/database"
)

type NotificationPostgresRepository struct {
	db *database.PostgresDatabase
}

func NewNotificationPostgresRepository(db *database.PostgresDatabase) NotificationRepository {
	return &NotificationPostgresRepository{db: db}
}

func (r *NotificationPostgresRepository) GetNotificationByID(ctx context.Context, id uuid.UUID) (*entities.Notification, error) {
	notifications, err := r.GetNotificationsByIDs(ctx, []uuid.UUID{id})
	if err != nil {
		return nil, err
	}
	if len(notifications) == 0 {
		return nil, ErrNotFound
	}
	return notifications[0], nil
}

func (r *NotificationPostgresRepository) GetNewNotifications(ctx context.Context, limit uint) ([]*entities.Notification, error) {
	if limit > config.Cfg.MaxBatchSize {
		return nil, ErrMaxBatchSizeExceeded
	}

	query := `
		select *
		from notifications
		where status = $1
		order by created_at
		limit $2
	`
	notifications := make([]*entities.Notification, 0, limit)
	rows, err := r.db.Pool.Query(ctx, query, entities.StatusPending, limit)
	if err != nil {
		return nil, fmt.Errorf("NotificationPostgresRepository.GetNotifications query error: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var notification entities.Notification
		err := rows.Scan(
			&notification.ID,
			&notification.DeliveryType,
			&notification.Recipient,
			&notification.Content,
			&notification.Status,
			&notification.Retries,
			&notification.CreatedAt,
			&notification.SentAt,
		)
		if err != nil {
			return nil, fmt.Errorf("NotificationPostgresRepository.GetNotifications scan error: %w", err)
		}
		notifications = append(notifications, &notification)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("NotificationPostgresRepository.GetNotifications rows iteration error: %w", err)
	}
	return notifications, nil
}

func (r *NotificationPostgresRepository) GetNotificationsByIDs(ctx context.Context, ids []uuid.UUID) ([]*entities.Notification, error) {
	if len(ids) == 0 {
		return []*entities.Notification{}, nil
	}
	if len(ids) > int(config.Cfg.MaxBatchSize) {
		return nil, ErrMaxBatchSizeExceeded
	}
	placeholders := make([]string, len(ids))
	args := make([]any, len(ids))
	for i, id := range ids {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
		args[i] = id
	}

	query := fmt.Sprintf(`
		select *
		from notifications
		where id in (%s)`,
		strings.Join(placeholders, ","),
	)
	rows, err := r.db.Pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("NotificationPostgresRepository.GetNotificationsByIDs query error: %w", err)
	}
	defer rows.Close()

	var notifications []*entities.Notification
	for rows.Next() {
		var notification entities.Notification
		err := rows.Scan(
			&notification.ID,
			&notification.DeliveryType,
			&notification.Recipient,
			&notification.Content,
			&notification.Status,
			&notification.Retries,
			&notification.CreatedAt,
			&notification.SentAt,
		)
		if err != nil {
			return nil, fmt.Errorf("NotificationPostgresRepository.GetNotificationsByIDs scan error: %w", err)
		}
		notifications = append(notifications, &notification)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("NotificationPostgresRepository.GetNotificationsByIDs rows error: %w", err)
	}
	return notifications, nil
}

func (r *NotificationPostgresRepository) CreateNotifications(ctx context.Context, notifications []*entities.Notification) error {
	if len(notifications) == 0 {
		return nil
	}
	if len(notifications) > int(config.Cfg.MaxBatchSize) {
		return ErrMaxBatchSizeExceeded
	}

	query := "insert into notifications (delivery_type, recipient, content) values "
	args := make([]any, 0, len(notifications)*3)
	values := make([]string, 0, len(notifications))
	for i, notification := range notifications {
		values = append(values, fmt.Sprintf("($%d, $%d, $%d)", i*3+1, i*3+2, i*3+3))
		args = append(args, notification.DeliveryType, notification.Recipient, notification.Content)
	}
	query += strings.Join(values, ",")
	query += " returning *"

	rows, err := r.db.Pool.Query(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("NotificationPostgresRepository.CreateNotifications query error: %w", err)
	}
	defer rows.Close()

	i := 0
	for rows.Next() {
		notification := notifications[i]
		err := rows.Scan(
			&notification.ID,
			&notification.DeliveryType,
			&notification.Recipient,
			&notification.Content,
			&notification.Status,
			&notification.Retries,
			&notification.CreatedAt,
			&notification.SentAt,
		)
		if err != nil {
			return fmt.Errorf("NotificationPostgresRepository.CreateNotifications scan error: %w", err)
		}
		i++
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("NotificationPostgresRepository.CreateNotifications rows error: %w", err)
	}
	return nil
}

func (r *NotificationPostgresRepository) UpdateNotificationsStatus(ctx context.Context, ids []uuid.UUID, status string) error {
	if len(ids) == 0 {
		return nil
	}
	query := fmt.Sprintf(`
		update notifications
		set status = $1,
			sent_at = case when $1 = '%s' then now() else sent_at end
		where id = any($2)
	`, entities.StatusDelivered)
	_, err := r.db.Pool.Exec(ctx, query, status, ids)
	if err != nil {
		return fmt.Errorf("NotificationPostgresRepository.UpdateNotificationsStatus error: %w", err)
	}
	return nil
}

func (r *NotificationPostgresRepository) UpdateNotificationRetries(ctx context.Context, id uuid.UUID, retries uint8) error {
	query := `
		update notifications
		set retries = $1
		where id = $2
	`
	_, err := r.db.Pool.Exec(ctx, query, retries, id)
	if err != nil {
		return fmt.Errorf("NotificationPostgresRepository.UpdateNotificationRetries error: %w", err)
	}
	return nil
}
