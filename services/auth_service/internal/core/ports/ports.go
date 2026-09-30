package ports

import (
	"context"

	"github.com/despi4/cargos/auth_service/internal/core/domain"
	"github.com/google/uuid"
)

// UserRepository определяет контракт для сохранения и получения пользователей (обычно PostgreSQL/MySQL).
type UserRepository interface {
	Create(ctx context.Context, user *domain.User) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error)
	GetByEmail(ctx context.Context, email string) (*domain.User, error)
	GetByUsername(ctx context.Context, username string) (*domain.User, error)
	Update(ctx context.Context, user *domain.User) error
	Delete(ctx context.Context, id uuid.UUID) error
}

// SessionRepository определяет контракт для работы с сессиями и refresh-токенами (обычно Redis).
type SessionRepository interface {
	Create(ctx context.Context, session *domain.Session) error
	GetByID(ctx context.Context, sessionID uuid.UUID) (*domain.Session, error)
	GetByRefreshToken(ctx context.Context, refreshToken string) (*domain.Session, error)
	Delete(ctx context.Context, sessionID uuid.UUID) error
	DeleteByUserID(ctx context.Context, userID uuid.UUID) error
}

// TokenService определяет контракт для генерации и верификации токенов (JWT).
type TokenService interface {
	GenerateTokenPair(user *domain.User) (*domain.TokenPair, error)
	ValidateAccessToken(token string) (uuid.UUID, error)
}

// PasswordHasher определяет контракт для хэширования и проверки паролей (например, bcrypt/argon2).
type PasswordHasher interface {
	HashPassword(password string) (string, error)
	ComparePassword(password, hash string) error
}

// AuthService определяет входящий порт (Use Case) для бизнес-логики аутентификации.
type AuthService interface {
	Register(ctx context.Context, username, email, password string) (*domain.User, error)
	Login(ctx context.Context, email, password, clientIP, userAgent string) (*domain.TokenPair, error)
	RefreshToken(ctx context.Context, refreshToken, clientIP, userAgent string) (*domain.TokenPair, error)
	Logout(ctx context.Context, sessionID uuid.UUID) error
}
