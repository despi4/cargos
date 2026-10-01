package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/despi4/cargos/auth_service/internal/core/domain"
	"github.com/despi4/cargos/auth_service/internal/core/ports"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

var (
	ErrSessionNotFound = domain.ErrSessionNotFound
	ErrSessionExpired  = domain.ErrSessionExpired
)

type RedisSessionRepo struct {
	client *redis.Client
}

type redisSessionDTO struct {
	ID           uuid.UUID `json:"id"`
	UserID       uuid.UUID `json:"user_id"`
	RefreshToken string    `json:"refresh_token"`
	ClientIP     string    `json:"client_ip"`
	UserAgent    string    `json:"user_agent"`
	IsBlocked    bool      `json:"is_blocked"`
	ExpiresAt    time.Time `json:"expires_at"`
	CreatedAt    time.Time `json:"created_at"`
}

func NewRedisSessionRepo(client *redis.Client) *RedisSessionRepo {
	return &RedisSessionRepo{client: client}
}

var _ ports.SessionRepository = (*RedisSessionRepo)(nil)

func sessionKey(refreshToken string) string {
	return fmt.Sprintf("session:%s", refreshToken)
}

func sessionIDKey(sessionID uuid.UUID) string {
	return fmt.Sprintf("session_id:%s", sessionID.String())
}

func userSessionsKey(userID uuid.UUID) string {
	return fmt.Sprintf("user_sessions:%s", userID.String())
}

// Create сохраняет сессию в Redis, связывает session_id с refresh_token и добавляет в множество сессий пользователя.
func (r *RedisSessionRepo) Create(ctx context.Context, session *domain.Session) error {
	ttl := time.Until(session.ExpiresAt)
	if ttl <= 0 {
		return ErrSessionExpired
	}

	dto := redisSessionDTO{
		ID:           session.ID,
		UserID:       session.UserID,
		RefreshToken: session.RefreshToken,
		ClientIP:     session.ClientIP,
		UserAgent:    session.UserAgent,
		IsBlocked:    session.IsBlocked,
		ExpiresAt:    session.ExpiresAt,
		CreatedAt:    session.CreatedAt,
	}

	data, err := json.Marshal(dto)
	if err != nil {
		return fmt.Errorf("marshal session: %w", err)
	}

	sKey := sessionKey(session.RefreshToken)
	sidKey := sessionIDKey(session.ID)
	uKey := userSessionsKey(session.UserID)

	pipe := r.client.Pipeline()
	// 1. Сохраняем сессию по refresh токену с TTL
	pipe.Set(ctx, sKey, data, ttl)
	// 2. Индекс session_id -> refresh_token для поиска и удаления по sessionID
	pipe.Set(ctx, sidKey, session.RefreshToken, ttl)
	// 3. Добавляем refresh_token в список сессий пользователя
	pipe.SAdd(ctx, uKey, session.RefreshToken)

	pipe.Expire(ctx, uKey, ttl)

	_, err = pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("execute redis pipeline: %w", err)
	}

	return nil
}

// GetByRefreshToken возвращает сессию по refresh токену.
func (r *RedisSessionRepo) GetByRefreshToken(ctx context.Context, refreshToken string) (*domain.Session, error) {
	val, err := r.client.Get(ctx, sessionKey(refreshToken)).Result()
	if errors.Is(err, redis.Nil) {
		return nil, ErrSessionNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get session by refresh token: %w", err)
	}

	var dto redisSessionDTO
	if err := json.Unmarshal([]byte(val), &dto); err != nil {
		return nil, fmt.Errorf("unmarshal session: %w", err)
	}

	return toDomainSession(&dto), nil
}

// GetByID возвращает сессию по её уникальному UUID идентификатору.
func (r *RedisSessionRepo) GetByID(ctx context.Context, sessionID uuid.UUID) (*domain.Session, error) {
	refreshToken, err := r.client.Get(ctx, sessionIDKey(sessionID)).Result()
	if errors.Is(err, redis.Nil) {
		return nil, ErrSessionNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get session id index: %w", err)
	}

	return r.GetByRefreshToken(ctx, refreshToken)
}

// Delete удаляет сессию по sessionID (очищая данные, индекс и удаляя из списка пользователя).
func (r *RedisSessionRepo) Delete(ctx context.Context, sessionID uuid.UUID) error {
	session, err := r.GetByID(ctx, sessionID)
	if errors.Is(err, ErrSessionNotFound) {
		return nil
	}
	if err != nil {
		return err
	}

	pipe := r.client.Pipeline()
	pipe.Del(ctx, sessionKey(session.RefreshToken))
	pipe.Del(ctx, sessionIDKey(session.ID))
	pipe.SRem(ctx, userSessionsKey(session.UserID), session.RefreshToken)

	_, err = pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("delete session: %w", err)
	}

	return nil
}

func (r *RedisSessionRepo) DeleteByUserID(ctx context.Context, userID uuid.UUID) error {
    uKey := userSessionsKey(userID)

    tokens, err := r.client.SMembers(ctx, uKey).Result()
    if err != nil {
        return fmt.Errorf("get user sessions: %w", err)
    }

    if len(tokens) == 0 {
        return nil
    }

    pipe := r.client.Pipeline()
    for _, token := range tokens {
        // Просто удаляем сессии. sessionIDKey умрет сам по TTL.
        pipe.Del(ctx, sessionKey(token))
    }
    // Удаляем множество пользователя
    pipe.Del(ctx, uKey)

    _, err = pipe.Exec(ctx)
    if err != nil {
        return fmt.Errorf("delete all user sessions: %w", err)
    }

    return nil
}

func toDomainSession(dto *redisSessionDTO) *domain.Session {
	return &domain.Session{
		ID:           dto.ID,
		UserID:       dto.UserID,
		RefreshToken: dto.RefreshToken,
		ClientIP:     dto.ClientIP,
		UserAgent:    dto.UserAgent,
		IsBlocked:    dto.IsBlocked,
		ExpiresAt:    dto.ExpiresAt,
		CreatedAt:    dto.CreatedAt,
	}
}