package ports

import (
	"context"

	"github.com/depsi4/cargos/auth_service/internal/core/domain"
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

// SessionRepository stores revocable sessions and atomically rotates refresh tokens.
type SessionRepository interface {
	Create(ctx context.Context, session *domain.Session) error
	GetByID(ctx context.Context, sessionID uuid.UUID) (*domain.Session, error)
	GetByRefreshHash(ctx context.Context, refreshHash string) (*domain.Session, error)
	Delete(ctx context.Context, sessionID uuid.UUID) error
	DeleteByUserID(ctx context.Context, userID uuid.UUID) error
	Rotate(ctx context.Context, previous *domain.Session, next *domain.Session) error
}

// TokenService определяет контракт для генерации и верификации токенов (JWT).
type TokenService interface {
	GenerateTokenPair(userID, sessionID uuid.UUID) (*domain.TokenPair, error)
	ValidateAccessToken(token string) (*domain.Claims, error)
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
	CurrentUser(ctx context.Context, accessToken string) (*domain.User, error)
	LogoutAccess(ctx context.Context, accessToken string) error
}
