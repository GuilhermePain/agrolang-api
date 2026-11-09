package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/GuilhermePain/agrolang-api/internal/model"
)

var ErrEmailTaken = errors.New("repository: email already registered")

// SignupInput carries everything created at account signup. Producer and
// Properties are set for the producer profile and empty for technicians; there
// is one property per crop. The producer's name and phone are copied from User.
type SignupInput struct {
	User             model.User
	Producer         *model.Producer
	Properties       []model.Property
	SessionTokenHash string
	SessionExpiresAt time.Time
}

type SignupResult struct {
	User       model.User
	Producer   *model.Producer
	Properties []model.Property
}

type SignupRepository struct {
	pool *pgxpool.Pool
}

func NewSignupRepository(pool *pgxpool.Pool) *SignupRepository {
	return &SignupRepository{pool: pool}
}

// Signup creates the user, optional producer and property, and the first
// session in one transaction: either all of it exists or none of it.
func (r *SignupRepository) Signup(ctx context.Context, in SignupInput) (SignupResult, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return SignupResult{}, fmt.Errorf("repository: begin signup: %w", err)
	}
	defer tx.Rollback(ctx)

	concerns := in.User.Concerns
	if concerns == nil {
		concerns = []string{}
	}
	onboarding := in.User.Onboarding
	if onboarding == nil {
		onboarding = map[string]string{}
	}

	const insertUser = `
		INSERT INTO users (email, password_hash, profile, name, whatsapp_phone, organization, concerns, onboarding)
		VALUES (lower($1), $2, $3, $4, $5, $6, $7, $8)
		RETURNING ` + userReturning

	user, err := scanUser(tx.QueryRow(ctx, insertUser,
		in.User.Email, in.User.PasswordHash, in.User.Profile, in.User.Name, in.User.WhatsAppPhone,
		in.User.Organization, concerns, onboarding,
	))
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return SignupResult{}, ErrEmailTaken
		}
		return SignupResult{}, fmt.Errorf("repository: signup insert user: %w", err)
	}

	result := SignupResult{User: user}

	if in.Producer != nil {
		const insertProducer = `
			INSERT INTO producers (user_id, name, whatsapp_phone, city, state)
			VALUES ($1, $2, $3, $4, $5)
			RETURNING id, name, whatsapp_phone, city, state, created_at, updated_at`

		var producer model.Producer
		if err := tx.QueryRow(ctx, insertProducer, user.ID, user.Name, user.WhatsAppPhone, in.Producer.City, in.Producer.State).Scan(
			&producer.ID, &producer.Name, &producer.WhatsAppPhone, &producer.City, &producer.State,
			&producer.CreatedAt, &producer.UpdatedAt,
		); err != nil {
			return SignupResult{}, fmt.Errorf("repository: signup insert producer: %w", err)
		}
		result.Producer = &producer

		const insertProperty = `
			INSERT INTO properties (producer_id, location, crop, soil_type, crop_stage)
			VALUES ($1, ST_SetSRID(ST_MakePoint($2, $3), 4326), $4, $5, $6)
			RETURNING id, producer_id, ST_X(location::geometry), ST_Y(location::geometry),
				crop, soil_type, crop_stage, created_at, updated_at`

		for _, prop := range in.Properties {
			var property model.Property
			if err := tx.QueryRow(ctx, insertProperty,
				producer.ID, prop.Longitude, prop.Latitude, prop.Crop, prop.SoilType, prop.CropStage,
			).Scan(
				&property.ID, &property.ProducerID, &property.Longitude, &property.Latitude,
				&property.Crop, &property.SoilType, &property.CropStage, &property.CreatedAt, &property.UpdatedAt,
			); err != nil {
				return SignupResult{}, fmt.Errorf("repository: signup insert property: %w", err)
			}
			result.Properties = append(result.Properties, property)
		}
	}

	if _, err := tx.Exec(ctx,
		`INSERT INTO sessions (token_hash, user_id, expires_at) VALUES ($1, $2, $3)`,
		in.SessionTokenHash, user.ID, in.SessionExpiresAt,
	); err != nil {
		return SignupResult{}, fmt.Errorf("repository: signup insert session: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return SignupResult{}, fmt.Errorf("repository: commit signup: %w", err)
	}
	return result, nil
}
