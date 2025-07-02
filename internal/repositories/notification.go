package repositories

import (
	"context"
	"fmt"
	"strings"
	"time"

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

func (r *NotificationPostgresRepository) GetNotifications(ctx context.Context, cursorCreatedAt time.Time, cursorId uuid.UUID, limit int) ([]*entities.Notification, error) {
	if limit > config.Cfg.MaxBatchSize {
		return nil, ErrMaxBatchSizeExceeded
	}
	var (
		query string
		args  []any
	)
	if cursorId == uuid.Nil && cursorCreatedAt.IsZero() {
		query = `
			select *
			from notifications
			order by created_at, id
			limit $1`
		args = []any{limit}
	} else {
		query = `
			select *
			from notifications
			where (created_at, id) > ($1, $2)
			order by created_at, id
			limit $3`
		args = []any{cursorCreatedAt, cursorId, limit}
	}
	notifications := make([]*entities.Notification, 0, limit)
	rows, err := r.db.Pool.Query(ctx, query, args...)
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
	if len(ids) > config.Cfg.MaxBatchSize {
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

func (r *NotificationPostgresRepository) CreateNotification(ctx context.Context, notification *entities.Notification) error {
	err := r.CreateNotifications(ctx, []*entities.Notification{notification})
	return err
}

func (r *NotificationPostgresRepository) CreateNotifications(ctx context.Context, notifications []*entities.Notification) error {
	if len(notifications) == 0 {
		return nil
	}
	if len(notifications) > config.Cfg.MaxBatchSize {
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
