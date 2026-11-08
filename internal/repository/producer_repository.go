package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/GuilhermePain/agrolang-api/internal/model"
)

var ErrNotFound = errors.New("repository: not found")

type ProducerRepository struct {
	pool *pgxpool.Pool
}

func NewProducerRepository(pool *pgxpool.Pool) *ProducerRepository {
	return &ProducerRepository{pool: pool}
}

func (r *ProducerRepository) Create(ctx context.Context, p model.Producer) (model.Producer, error) {
	const query = `
		INSERT INTO producers (name, whatsapp_phone, city, state)
		VALUES ($1, $2, $3, $4)
		RETURNING id, name, whatsapp_phone, city, state, created_at, updated_at`

	var created model.Producer
	err := r.pool.QueryRow(ctx, query, p.Name, p.WhatsAppPhone, p.City, p.State).Scan(
		&created.ID, &created.Name, &created.WhatsAppPhone, &created.City, &created.State,
		&created.CreatedAt, &created.UpdatedAt,
	)
	if err != nil {
		return model.Producer{}, fmt.Errorf("repository: create producer: %w", err)
	}
	return created, nil
}

func (r *ProducerRepository) ListAll(ctx context.Context) ([]model.Producer, error) {
	const query = `
		SELECT id, name, whatsapp_phone, city, state, created_at, updated_at
		FROM producers
		ORDER BY created_at`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("repository: list producers: %w", err)
	}
	defer rows.Close()

	var all []model.Producer
	for rows.Next() {
		var p model.Producer
		if err := rows.Scan(
			&p.ID, &p.Name, &p.WhatsAppPhone, &p.City, &p.State, &p.CreatedAt, &p.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("repository: scan producer: %w", err)
		}
		all = append(all, p)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("repository: iterate producers: %w", err)
	}
	return all, nil
}

func (r *ProducerRepository) GetByID(ctx context.Context, id string) (model.Producer, error) {
	const query = `
		SELECT p.id, p.name, p.whatsapp_phone, p.city, p.state, p.created_at, p.updated_at,
			COALESCE(u.concerns, '{}')
		FROM producers p
		LEFT JOIN users u ON u.id = p.user_id
		WHERE p.id = $1`

	var p model.Producer
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&p.ID, &p.Name, &p.WhatsAppPhone, &p.City, &p.State, &p.CreatedAt, &p.UpdatedAt,
		&p.Concerns,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return model.Producer{}, ErrNotFound
	}
	if err != nil {
		return model.Producer{}, fmt.Errorf("repository: get producer: %w", err)
	}
	return p, nil
}

func (r *ProducerRepository) CreateForUser(ctx context.Context, userID string, p model.Producer) (model.Producer, error) {
	const query = `
		INSERT INTO producers (user_id, name, whatsapp_phone, city, state)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, name, whatsapp_phone, city, state, created_at, updated_at`

	var created model.Producer
	err := r.pool.QueryRow(ctx, query, userID, p.Name, p.WhatsAppPhone, p.City, p.State).Scan(
		&created.ID, &created.Name, &created.WhatsAppPhone, &created.City, &created.State,
		&created.CreatedAt, &created.UpdatedAt,
	)
	if err != nil {
		return model.Producer{}, fmt.Errorf("repository: create producer for user: %w", err)
	}
	return created, nil
}

func (r *ProducerRepository) ListByUser(ctx context.Context, userID string) ([]model.Producer, error) {
	const query = `
		SELECT id, name, whatsapp_phone, city, state, created_at, updated_at
		FROM producers
		WHERE user_id = $1
		ORDER BY created_at`

	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("repository: list producers by user: %w", err)
	}
	defer rows.Close()

	all := make([]model.Producer, 0)
	for rows.Next() {
		var p model.Producer
		if err := rows.Scan(
			&p.ID, &p.Name, &p.WhatsAppPhone, &p.City, &p.State, &p.CreatedAt, &p.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("repository: scan producer: %w", err)
		}
		all = append(all, p)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("repository: iterate producers: %w", err)
	}
	return all, nil
}

func (r *ProducerRepository) GetByIDForUser(ctx context.Context, id, userID string) (model.Producer, error) {
	const query = `
		SELECT id, name, whatsapp_phone, city, state, created_at, updated_at
		FROM producers
		WHERE id = $1 AND user_id = $2`

	var p model.Producer
	err := r.pool.QueryRow(ctx, query, id, userID).Scan(
		&p.ID, &p.Name, &p.WhatsAppPhone, &p.City, &p.State, &p.CreatedAt, &p.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return model.Producer{}, ErrNotFound
	}
	if err != nil {
		return model.Producer{}, fmt.Errorf("repository: get producer for user: %w", err)
	}
	return p, nil
}

// UpdateForUser updates the producer only if it belongs to the user;
// otherwise (missing or someone else's) it returns ErrNotFound.
func (r *ProducerRepository) UpdateForUser(ctx context.Context, userID string, p model.Producer) (model.Producer, error) {
	const query = `
		UPDATE producers
		SET name = $3, whatsapp_phone = $4, city = $5, state = $6, updated_at = now()
		WHERE id = $1 AND user_id = $2
		RETURNING id, name, whatsapp_phone, city, state, created_at, updated_at`

	var updated model.Producer
	err := r.pool.QueryRow(ctx, query, p.ID, userID, p.Name, p.WhatsAppPhone, p.City, p.State).Scan(
		&updated.ID, &updated.Name, &updated.WhatsAppPhone, &updated.City, &updated.State,
		&updated.CreatedAt, &updated.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return model.Producer{}, ErrNotFound
	}
	if err != nil {
		return model.Producer{}, fmt.Errorf("repository: update producer for user: %w", err)
	}
	return updated, nil
}

// DeleteForUser deletes the producer (and, by cascade, its properties) only
// if it belongs to the user; otherwise it returns ErrNotFound.
func (r *ProducerRepository) DeleteForUser(ctx context.Context, id, userID string) error {
	const query = `DELETE FROM producers WHERE id = $1 AND user_id = $2`

	res, err := r.pool.Exec(ctx, query, id, userID)
	if err != nil {
		return fmt.Errorf("repository: delete producer for user: %w", err)
	}
	if res.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}
