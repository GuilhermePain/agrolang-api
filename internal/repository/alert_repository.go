package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/GuilhermePain/agrolang-api/internal/model"
)

type AlertRepository struct {
	pool *pgxpool.Pool
}

func NewAlertRepository(pool *pgxpool.Pool) *AlertRepository {
	return &AlertRepository{pool: pool}
}

func (r *AlertRepository) Create(ctx context.Context, a model.Alert) (model.Alert, error) {
	const query = `
		INSERT INTO alerts (property_id, level, alert_type, period_start, period_end)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, property_id, level, alert_type, period_start, period_end, triggered_at, resolved_at`

	var created model.Alert
	err := r.pool.QueryRow(ctx, query, a.PropertyID, a.Level, a.AlertType, a.PeriodStart, a.PeriodEnd).Scan(
		&created.ID, &created.PropertyID, &created.Level, &created.AlertType,
		&created.PeriodStart, &created.PeriodEnd, &created.TriggeredAt, &created.ResolvedAt,
	)
	if err != nil {
		return model.Alert{}, fmt.Errorf("repository: create alert: %w", err)
	}
	return created, nil
}

func (r *AlertRepository) ListByProperty(ctx context.Context, propertyID string) ([]model.Alert, error) {
	const query = `
		SELECT id, property_id, level, alert_type, period_start, period_end, triggered_at, resolved_at
		FROM alerts
		WHERE property_id = $1
		ORDER BY triggered_at DESC`

	return r.queryAlerts(ctx, query, propertyID)
}

func (r *AlertRepository) ListRecent(ctx context.Context, limit int) ([]model.Alert, error) {
	const query = `
		SELECT id, property_id, level, alert_type, period_start, period_end, triggered_at, resolved_at
		FROM alerts
		ORDER BY triggered_at DESC
		LIMIT $1`

	return r.queryAlerts(ctx, query, limit)
}

// HasActiveAlert reports whether there is already an unresolved alert of the
// given type for the property.
func (r *AlertRepository) HasActiveAlert(ctx context.Context, propertyID, alertType string) (bool, error) {
	const query = `
		SELECT EXISTS (
			SELECT 1 FROM alerts
			WHERE property_id = $1
			  AND alert_type  = $2
			  AND resolved_at IS NULL
		)`

	var exists bool
	if err := r.pool.QueryRow(ctx, query, propertyID, alertType).Scan(&exists); err != nil {
		return false, fmt.Errorf("repository: has active alert: %w", err)
	}
	return exists, nil
}

// Resolve marks a single alert as resolved.
func (r *AlertRepository) Resolve(ctx context.Context, id string) error {
	const query = `
		UPDATE alerts
		SET resolved_at = now()
		WHERE id = $1 AND resolved_at IS NULL`

	res, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("repository: resolve alert: %w", err)
	}
	if res.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// ResolveInactiveForProperty resolves unresolved alerts of types that are no longer
// detected as active in the latest scan cycle.
func (r *AlertRepository) ResolveInactiveForProperty(ctx context.Context, propertyID string, activeAlertTypes []string) error {
	var query string
	var args []any

	if len(activeAlertTypes) == 0 {
		query = `UPDATE alerts SET resolved_at = now() WHERE property_id = $1 AND resolved_at IS NULL`
		args = []any{propertyID}
	} else {
		query = `UPDATE alerts SET resolved_at = now() WHERE property_id = $1 AND resolved_at IS NULL AND alert_type != ALL($2)`
		args = []any{propertyID, activeAlertTypes}
	}

	if _, err := r.pool.Exec(ctx, query, args...); err != nil {
		return fmt.Errorf("repository: auto resolve inactive alerts: %w", err)
	}
	return nil
}

func (r *AlertRepository) queryAlerts(ctx context.Context, query string, arg any) ([]model.Alert, error) {
	rows, err := r.pool.Query(ctx, query, arg)
	if err != nil {
		return nil, fmt.Errorf("repository: query alerts: %w", err)
	}
	defer rows.Close()

	all := make([]model.Alert, 0)
	for rows.Next() {
		var a model.Alert
		if err := rows.Scan(
			&a.ID, &a.PropertyID, &a.Level, &a.AlertType,
			&a.PeriodStart, &a.PeriodEnd, &a.TriggeredAt, &a.ResolvedAt,
		); err != nil {
			return nil, fmt.Errorf("repository: scan alert: %w", err)
		}
		all = append(all, a)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("repository: iterate alerts: %w", err)
	}
	return all, nil
}
