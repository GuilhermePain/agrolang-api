package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/GuilhermePain/agrolang-api/internal/model"
)

type NotificationRepository struct {
	pool *pgxpool.Pool
}

func NewNotificationRepository(pool *pgxpool.Pool) *NotificationRepository {
	return &NotificationRepository{pool: pool}
}

func (r *NotificationRepository) Create(ctx context.Context, n model.NotificationLog) (model.NotificationLog, error) {
	const query = `
		INSERT INTO notification_logs (alert_id, producer_id, property_id, phone, message, status, error_message)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, alert_id, producer_id, property_id, phone, message, status, sent_at, delivered_at, read_at, error_message`

	var created model.NotificationLog
	err := r.pool.QueryRow(ctx, query, n.AlertID, n.ProducerID, n.PropertyID, n.Phone, n.Message, n.Status, n.ErrorMessage).Scan(
		&created.ID, &created.AlertID, &created.ProducerID, &created.PropertyID,
		&created.Phone, &created.Message, &created.Status, &created.SentAt,
		&created.DeliveredAt, &created.ReadAt, &created.ErrorMessage,
	)
	if err != nil {
		return model.NotificationLog{}, fmt.Errorf("repository: create notification log: %w", err)
	}
	return created, nil
}

// ListRecentForUser returns the latest notification logs of the user's
// producers.
func (r *NotificationRepository) ListRecentForUser(ctx context.Context, userID string, limit int) ([]model.NotificationLog, error) {
	const query = `
		SELECT n.id, n.alert_id, n.producer_id, n.property_id, n.phone, n.message, n.status,
			n.sent_at, n.delivered_at, n.read_at, n.error_message
		FROM notification_logs n
		JOIN producers pr ON pr.id = n.producer_id
		WHERE pr.user_id = $1
		ORDER BY n.sent_at DESC
		LIMIT $2`

	return r.queryLogs(ctx, query, userID, limit)
}

// ListByProducerForUser lists a producer's notification logs only if the
// producer belongs to the user; otherwise it returns ErrNotFound.
func (r *NotificationRepository) ListByProducerForUser(ctx context.Context, producerID, userID string) ([]model.NotificationLog, error) {
	const ownedQuery = `SELECT EXISTS (SELECT 1 FROM producers WHERE id = $1 AND user_id = $2)`

	var owned bool
	if err := r.pool.QueryRow(ctx, ownedQuery, producerID, userID).Scan(&owned); err != nil {
		return nil, fmt.Errorf("repository: check producer ownership: %w", err)
	}
	if !owned {
		return nil, ErrNotFound
	}

	const query = `
		SELECT n.id, n.alert_id, n.producer_id, n.property_id, n.phone, n.message, n.status,
			n.sent_at, n.delivered_at, n.read_at, n.error_message
		FROM notification_logs n
		JOIN producers pr ON pr.id = n.producer_id
		WHERE n.producer_id = $1 AND pr.user_id = $2
		ORDER BY n.sent_at DESC`

	return r.queryLogs(ctx, query, producerID, userID)
}

func (r *NotificationRepository) UpdateStatusByPhone(ctx context.Context, phone string, status model.NotificationStatus, eventTime time.Time) error {
	var query string
	switch status {
	case model.NotificationStatusDelivered:
		query = `
			UPDATE notification_logs
			SET status = $2, delivered_at = $3
			WHERE id = (
				SELECT id FROM notification_logs
				WHERE phone = $1
				ORDER BY sent_at DESC
				LIMIT 1
			)`
	case model.NotificationStatusRead:
		query = `
			UPDATE notification_logs
			SET status = $2, read_at = $3
			WHERE id = (
				SELECT id FROM notification_logs
				WHERE phone = $1
				ORDER BY sent_at DESC
				LIMIT 1
			)`
	default:
		query = `
			UPDATE notification_logs
			SET status = $2
			WHERE id = (
				SELECT id FROM notification_logs
				WHERE phone = $1
				ORDER BY sent_at DESC
				LIMIT 1
			)`
	}

	if _, err := r.pool.Exec(ctx, query, phone, status, eventTime); err != nil {
		return fmt.Errorf("repository: update notification status: %w", err)
	}
	return nil
}

func (r *NotificationRepository) queryLogs(ctx context.Context, query string, args ...any) ([]model.NotificationLog, error) {
	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("repository: query notification logs: %w", err)
	}
	defer rows.Close()

	all := make([]model.NotificationLog, 0)
	for rows.Next() {
		var n model.NotificationLog
		if err := rows.Scan(
			&n.ID, &n.AlertID, &n.ProducerID, &n.PropertyID,
			&n.Phone, &n.Message, &n.Status, &n.SentAt,
			&n.DeliveredAt, &n.ReadAt, &n.ErrorMessage,
		); err != nil {
			return nil, fmt.Errorf("repository: scan notification log: %w", err)
		}
		all = append(all, n)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("repository: iterate notification logs: %w", err)
	}
	return all, nil
}
