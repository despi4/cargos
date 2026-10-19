package postgres

import (
	"context"

	"github.com/depsi4/cargos/internal/platform/postgres"
	"github.com/depsi4/cargos/notification_service/internal/core/domain"
	"github.com/depsi4/cargos/pkg/entity"
	"github.com/depsi4/cargos/pkg/fault"
	"github.com/jackc/pgx/v5/pgxpool"
)

type NotificationRepository struct{ pool *pgxpool.Pool }

func NewNotificationRepository(pool *pgxpool.Pool) *NotificationRepository {
	return &NotificationRepository{pool: pool}
}
func scanNotification(row postgres.Scanner) (*domain.Notification, error) {
	var v domain.Notification
	err := row.Scan(&v.ID, &v.CreatedAt, &v.UpdatedAt, &v.EventID, &v.Title, &v.Message, &v.Read)
	if err != nil {
		return nil, postgres.Error(err)
	}
	return &v, nil
}
func (r *NotificationRepository) Create(ctx context.Context, owner string, v *domain.Notification) error {
	_, err := r.pool.Exec(ctx, `INSERT INTO notifications(id,created_at,updated_at,event_id,title,message,is_read,owner_id) VALUES($1,$2,$3,$4,$5,$6,$7,$8)`, v.ID, v.CreatedAt, v.UpdatedAt, v.EventID, v.Title, v.Message, v.Read, owner)
	return postgres.Error(err)
}
func (r *NotificationRepository) Get(ctx context.Context, owner, id string) (*domain.Notification, error) {
	if !entity.ValidID(id) {
		return nil, fault.ErrNotFound
	}
	return scanNotification(r.pool.QueryRow(ctx, `SELECT id::text,created_at,updated_at,event_id::text,title,message,is_read FROM notifications WHERE owner_id=$1 AND id=$2`, owner, id))
}
func (r *NotificationRepository) List(ctx context.Context, owner string, limit, offset int) ([]domain.Notification, error) {
	rows, err := r.pool.Query(ctx, `SELECT id::text,created_at,updated_at,event_id::text,title,message,is_read FROM notifications WHERE owner_id=$1 ORDER BY created_at,id LIMIT $2 OFFSET $3`, owner, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []domain.Notification{}
	for rows.Next() {
		v, err := scanNotification(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, *v)
	}
	return items, rows.Err()
}
func (r *NotificationRepository) SetRead(ctx context.Context, owner, id string, read bool) (*domain.Notification, error) {
	if !entity.ValidID(id) {
		return nil, fault.ErrNotFound
	}
	return scanNotification(r.pool.QueryRow(ctx, `UPDATE notifications SET is_read=$3,updated_at=now() WHERE owner_id=$1 AND id=$2 RETURNING id::text,created_at,updated_at,event_id::text,title,message,is_read`, owner, id, read))
}
