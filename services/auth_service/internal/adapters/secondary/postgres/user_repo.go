package postgres

import (
	"context"

	"github.com/despi4/cargos/auth_service/internal/core/domain"
	"github.com/despi4/cargos/auth_service/internal/core/ports"
	"github.com/despi4/cargos/auth_service/internal/adapters/secondary/postgres/db"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepo struct {
	pool *pgxpool.Pool
	q db.Querier
}

func NewUserRepo(pool *pgxpool.Pool) *UserRepo {
	return &UserRepo{
		pool: pool,
		q: db.New(pool),
	}
}

var _ ports.UserRepository = (*UserRepo)(nil)

func (r *UserRepo) Create(ctx context.Context, user *domain.User) error {
    // Формируем параметры только из тех полей, которые требует запрос
    userParams := db.CreateUserParams{
        Username:     user.Username,
        Email:        user.Email,
        PasswordHash: user.PasswordHash,
    }

    // Вызываем базу. Она сама сгенерирует ID и даты, и вернет их через RETURNING
    dbUser, err := r.q.CreateUser(ctx, userParams)
    if err != nil {
        return err // Здесь можно дополнительно оборачивать ошибку, например выявлять дубликат email
    }

    // Обогащаем доменную модель данными, которые сгенерировала БД
    user.ID = uuid.UUID(dbUser.ID.Bytes)
    user.CreatedAt = dbUser.CreatedAt.Time
    user.UpdatedAt = dbUser.UpdatedAt.Time
    
    // Возвращаем только error. Сама структура user (по указателю) уже обновилась
    return nil
}

func(r *UserRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	pgUUID := pgtype.UUID {
		Bytes: id,
		Valid: true,		
	}
	
	pgUser, err := r.q.GetUserByID(ctx, pgUUID)
	if err != nil {
		return nil, err
	}

	user := domain.User {
		ID: pgUUID.Bytes,
		Username: pgUser.Username,
		Email: pgUser.Email,
		PasswordHash: pgUser.PasswordHash,
		CreatedAt: pgUser.CreatedAt.Time,
		UpdatedAt: pgUser.UpdatedAt.Time,
	}

	return &user, nil
}

func(r *UserRepo) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	pgUser, err := r.q.GetUserByEmail(ctx, email)
	if err != nil {
		return nil, err
	}

	user := domain.User {
		ID: pgUser.ID.Bytes,
		Username: pgUser.Username,
		Email: pgUser.Email,
		PasswordHash: pgUser.PasswordHash,
		CreatedAt: pgUser.CreatedAt.Time,
		UpdatedAt: pgUser.UpdatedAt.Time,
	}

	return &user, nil
}

func(r *UserRepo) GetByUsername(ctx context.Context, username string) (*domain.User, error) {
	pgUser, err := r.q.GetUserByUsername(ctx, username)
	if err != nil {
		return nil, err
	}

	user := domain.User {
		ID: pgUser.ID.Bytes,
		Username: pgUser.Username,
		Email: pgUser.Email,
		PasswordHash: pgUser.PasswordHash,
		CreatedAt: pgUser.CreatedAt.Time,
		UpdatedAt: pgUser.UpdatedAt.Time,
	}

	return &user, nil
}

func(r *UserRepo) Update(ctx context.Context, user *domain.User) error {
	pgUUID := pgtype.UUID {
		Bytes: user.ID,
		Valid: true,
	}

	updateParams := db.UpdateUserParams{
		ID: pgUUID,
		Username: user.Username,
		Email: user.Email,
		PasswordHash: user.PasswordHash,
	}

	_, err := r.q.UpdateUser(ctx, updateParams)
	if err != nil {
		return err
	}

	return nil
}

func(r *UserRepo) Delete(ctx context.Context, id uuid.UUID) error {
	pgUUID := pgtype.UUID{
		Bytes: id,
		Valid: true,
	}

	err := r.q.DeleteUser(ctx, pgUUID)
	if err != nil {
		return err
	}

	return nil
}