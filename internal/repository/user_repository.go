package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/GuilhermePain/agrolang-api/internal/model"
)

// userColumns is qualified with the "u" alias for SELECTs; userReturning is
// the same list, unqualified, for INSERT ... RETURNING. Keep the order in
// sync with scanUser.
const (
	userColumns   = `u.id, u.email, u.password_hash, u.profile, u.name, u.whatsapp_phone, u.organization, u.concerns, u.onboarding, u.created_at`
	userReturning = `id, email, password_hash, profile, name, whatsapp_phone, organization, concerns, onboarding, created_at`
)

type UserRepository struct {
	pool *pgxpool.Pool
}

func NewUserRepository(pool *pgxpool.Pool) *UserRepository {
	return &UserRepository{pool: pool}
}

func scanUser(row pgx.Row) (model.User, error) {
	var u model.User
	if err := row.Scan(
		&u.ID, &u.Email, &u.PasswordHash, &u.Profile, &u.Name, &u.WhatsAppPhone,
		&u.Organization, &u.Concerns, &u.Onboarding, &u.CreatedAt,
	); err != nil {
		return model.User{}, err
	}
	if u.Concerns == nil {
		u.Concerns = []string{}
	}
	if u.Onboarding == nil {
		u.Onboarding = map[string]string{}
	}
	return u, nil
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (model.User, error) {
	const query = `SELECT ` + userColumns + ` FROM users u WHERE lower(u.email) = lower($1)`

	u, err := scanUser(r.pool.QueryRow(ctx, query, email))
	if errors.Is(err, pgx.ErrNoRows) {
		return model.User{}, ErrNotFound
	}
	if err != nil {
		return model.User{}, fmt.Errorf("repository: get user by email: %w", err)
	}
	return u, nil
}

// UserBySession returns the owner of a live (non-expired) session.
func (r *UserRepository) UserBySession(ctx context.Context, tokenHash string) (model.User, error) {
	const query = `SELECT ` + userColumns + `
		FROM sessions s
		JOIN users u ON u.id = s.user_id
		WHERE s.token_hash = $1 AND s.expires_at > now()`

	u, err := scanUser(r.pool.QueryRow(ctx, query, tokenHash))
	if errors.Is(err, pgx.ErrNoRows) {
		return model.User{}, ErrNotFound
	}
	if err != nil {
		return model.User{}, fmt.Errorf("repository: get user by session: %w", err)
	}
	return u, nil
}

func (r *UserRepository) CreateSession(ctx context.Context, userID, tokenHash string, expiresAt time.Time) error {
	const query = `INSERT INTO sessions (token_hash, user_id, expires_at) VALUES ($1, $2, $3)`

	if _, err := r.pool.Exec(ctx, query, tokenHash, userID, expiresAt); err != nil {
		return fmt.Errorf("repository: create session: %w", err)
	}
	return nil
}

func (r *UserRepository) DeleteSession(ctx context.Context, tokenHash string) error {
	if _, err := r.pool.Exec(ctx, `DELETE FROM sessions WHERE token_hash = $1`, tokenHash); err != nil {
		return fmt.Errorf("repository: delete session: %w", err)
	}
	return nil
}

func (r *UserRepository) DeleteExpiredSessions(ctx context.Context) error {
	if _, err := r.pool.Exec(ctx, `DELETE FROM sessions WHERE expires_at <= now()`); err != nil {
		return fmt.Errorf("repository: delete expired sessions: %w", err)
	}
	return nil
}
