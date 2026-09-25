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
		SELECT id, name, whatsapp_phone, city, state, created_at, updated_at
		FROM producers
		WHERE id = $1`

	var p model.Producer
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&p.ID, &p.Name, &p.WhatsAppPhone, &p.City, &p.State, &p.CreatedAt, &p.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return model.Producer{}, ErrNotFound
	}
	if err != nil {
		return model.Producer{}, fmt.Errorf("repository: get producer: %w", err)
	}
	return p, nil
}
