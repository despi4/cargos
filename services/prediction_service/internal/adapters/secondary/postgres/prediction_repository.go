package postgres

import (
	"context"

	"github.com/depsi4/cargos/internal/platform/postgres"
	"github.com/depsi4/cargos/pkg/entity"
	"github.com/depsi4/cargos/pkg/fault"
	"github.com/depsi4/cargos/prediction_service/internal/core/domain"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PredictionRepository struct{ pool *pgxpool.Pool }

func NewPredictionRepository(pool *pgxpool.Pool) *PredictionRepository {
	return &PredictionRepository{pool: pool}
}
func scanPrediction(row postgres.Scanner) (*domain.Prediction, error) {
	var v domain.Prediction
	err := row.Scan(&v.ID, &v.CreatedAt, &v.UpdatedAt, &v.WindowDays, &v.Recommendations, &v.NotificationSent)
	if err != nil {
		return nil, postgres.Error(err)
	}
	return &v, nil
}
func (r *PredictionRepository) Create(ctx context.Context, owner string, v *domain.Prediction) error {
	_, err := r.pool.Exec(ctx, `INSERT INTO predictions(id,created_at,updated_at,window_days,recommendations,notification_sent,owner_id) VALUES($1,$2,$3,$4,$5,$6,$7)`, v.ID, v.CreatedAt, v.UpdatedAt, v.WindowDays, v.Recommendations, v.NotificationSent, owner)
	return postgres.Error(err)
}
func (r *PredictionRepository) Get(ctx context.Context, owner, id string) (*domain.Prediction, error) {
	if !entity.ValidID(id) {
		return nil, fault.ErrNotFound
	}
	return scanPrediction(r.pool.QueryRow(ctx, `SELECT id::text,created_at,updated_at,window_days,recommendations,notification_sent FROM predictions WHERE owner_id=$1 AND id=$2`, owner, id))
}
func (r *PredictionRepository) List(ctx context.Context, owner string, limit, offset int) ([]domain.Prediction, error) {
	rows, err := r.pool.Query(ctx, `SELECT id::text,created_at,updated_at,window_days,recommendations,notification_sent FROM predictions WHERE owner_id=$1 ORDER BY created_at,id LIMIT $2 OFFSET $3`, owner, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []domain.Prediction{}
	for rows.Next() {
		v, err := scanPrediction(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, *v)
	}
	return items, rows.Err()
}
func (r *PredictionRepository) Pending(ctx context.Context) ([]domain.PendingDelivery, error) {
	rows, err := r.pool.Query(ctx, `SELECT owner_id::text,id::text,jsonb_array_length(recommendations) FROM predictions WHERE NOT notification_sent ORDER BY created_at LIMIT 100`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []domain.PendingDelivery{}
	for rows.Next() {
		var v domain.PendingDelivery
		if err = rows.Scan(&v.OwnerID, &v.PredictionID, &v.Count); err != nil {
			return nil, err
		}
		items = append(items, v)
	}
	return items, rows.Err()
}
func (r *PredictionRepository) MarkSent(ctx context.Context, owner, id string) error {
	_, err := r.pool.Exec(ctx, `UPDATE predictions SET notification_sent=true,updated_at=now() WHERE owner_id=$1 AND id=$2`, owner, id)
	return postgres.Error(err)
}
