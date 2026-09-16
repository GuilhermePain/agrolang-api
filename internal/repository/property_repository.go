package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/GuilhermePain/agrolang-api/internal/model"
)

type PropertyRepository struct {
	pool *pgxpool.Pool
}

func NewPropertyRepository(pool *pgxpool.Pool) *PropertyRepository {
	return &PropertyRepository{pool: pool}
}

func (r *PropertyRepository) Create(ctx context.Context, p model.Property) (model.Property, error) {
	const query = `
		INSERT INTO properties (producer_id, location, crop, soil_type, crop_stage)
		VALUES ($1, ST_SetSRID(ST_MakePoint($2, $3), 4326), $4, $5, $6)
		RETURNING id, producer_id, ST_X(location::geometry), ST_Y(location::geometry),
			crop, soil_type, crop_stage, created_at, updated_at`

	var created model.Property
	err := r.pool.QueryRow(ctx, query, p.ProducerID, p.Longitude, p.Latitude, p.Crop, p.SoilType, p.CropStage).Scan(
		&created.ID, &created.ProducerID, &created.Longitude, &created.Latitude,
		&created.Crop, &created.SoilType, &created.CropStage, &created.CreatedAt, &created.UpdatedAt,
	)
	if err != nil {
		return model.Property{}, fmt.Errorf("repository: create property: %w", err)
	}
	return created, nil
}

func (r *PropertyRepository) GetByID(ctx context.Context, id string) (model.Property, error) {
	const query = `
		SELECT id, producer_id, ST_X(location::geometry), ST_Y(location::geometry),
			crop, soil_type, crop_stage, created_at, updated_at
		FROM properties
		WHERE id = $1`

	var p model.Property
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&p.ID, &p.ProducerID, &p.Longitude, &p.Latitude,
		&p.Crop, &p.SoilType, &p.CropStage, &p.CreatedAt, &p.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return model.Property{}, ErrNotFound
	}
	if err != nil {
		return model.Property{}, fmt.Errorf("repository: get property: %w", err)
	}
	return p, nil
}

func (r *PropertyRepository) ListAll(ctx context.Context) ([]model.Property, error) {
	const query = `
		SELECT id, producer_id, ST_X(location::geometry), ST_Y(location::geometry),
			crop, soil_type, crop_stage, created_at, updated_at
		FROM properties
		ORDER BY created_at`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("repository: list properties: %w", err)
	}
	defer rows.Close()

	var all []model.Property
	for rows.Next() {
		var p model.Property
		if err := rows.Scan(
			&p.ID, &p.ProducerID, &p.Longitude, &p.Latitude,
			&p.Crop, &p.SoilType, &p.CropStage, &p.CreatedAt, &p.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("repository: scan property: %w", err)
		}
		all = append(all, p)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("repository: iterate properties: %w", err)
	}
	return all, nil
}
