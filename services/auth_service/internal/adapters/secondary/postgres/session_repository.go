package postgres

import (
	"context"
	"errors"

	"github.com/depsi4/cargos/auth_service/internal/core/domain"
	"github.com/depsi4/cargos/internal/platform/postgres"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type SessionRepository struct{ pool *pgxpool.Pool }

func NewSessionRepository(pool *pgxpool.Pool) *SessionRepository { return &SessionRepository{pool} }
func sessionError(err error) error {
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.ErrUnauthorized
	}
	return err
}
func scanSession(row postgres.Scanner) (*domain.Session, error) {
	var s domain.Session
	var id, userID string
	err := row.Scan(&id, &userID, &s.RefreshHash, &s.ClientIP, &s.UserAgent, &s.ExpiresAt, &s.CreatedAt)
	if err != nil {
		return nil, sessionError(err)
	}
	s.ID = uuid.MustParse(id)
	s.UserID = uuid.MustParse(userID)
	return &s, nil
}
func (r *SessionRepository) Create(ctx context.Context, s *domain.Session) error {
	_, err := r.pool.Exec(ctx, `INSERT INTO sessions(id,user_id,refresh_hash,client_ip,user_agent,expires_at,created_at) VALUES($1,$2,$3,$4,$5,$6,$7)`, s.ID.String(), s.UserID.String(), s.RefreshHash, s.ClientIP, s.UserAgent, s.ExpiresAt, s.CreatedAt)
	return err
}
func (r *SessionRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Session, error) {
	return scanSession(r.pool.QueryRow(ctx, `SELECT id::text,user_id::text,refresh_hash,client_ip,user_agent,expires_at,created_at FROM sessions WHERE id=$1 AND expires_at>now()`, id.String()))
}
func (r *SessionRepository) GetByRefreshHash(ctx context.Context, hash string) (*domain.Session, error) {
	return scanSession(r.pool.QueryRow(ctx, `SELECT id::text,user_id::text,refresh_hash,client_ip,user_agent,expires_at,created_at FROM sessions WHERE refresh_hash=$1 AND expires_at>now()`, hash))
}
func (r *SessionRepository) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM sessions WHERE id=$1`, id.String())
	return err
}
func (r *SessionRepository) DeleteByUserID(ctx context.Context, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM sessions WHERE user_id=$1`, id.String())
	return err
}
func (r *SessionRepository) Rotate(ctx context.Context, old, next *domain.Session) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	result, err := tx.Exec(ctx, `DELETE FROM sessions WHERE id=$1 AND refresh_hash=$2 AND expires_at>now()`, old.ID.String(), old.RefreshHash)
	if err != nil {
		return err
	}
	if result.RowsAffected() != 1 {
		return domain.ErrUnauthorized
	}
	_, err = tx.Exec(ctx, `INSERT INTO sessions(id,user_id,refresh_hash,client_ip,user_agent,expires_at,created_at) VALUES($1,$2,$3,$4,$5,$6,$7)`, next.ID.String(), next.UserID.String(), next.RefreshHash, next.ClientIP, next.UserAgent, next.ExpiresAt, next.CreatedAt)
	if err != nil {
		return err
	}
	return tx.Commit(ctx)
}
