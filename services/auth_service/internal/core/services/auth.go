package services

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"net/mail"
	"strings"
	"time"

	"github.com/depsi4/cargos/auth_service/internal/core/domain"
	"github.com/depsi4/cargos/auth_service/internal/core/ports"
	"github.com/depsi4/cargos/pkg/fault"
	"github.com/google/uuid"
)

type Auth struct {
	users     ports.UserRepository
	sessions  ports.SessionRepository
	tokens    ports.TokenService
	passwords ports.PasswordHasher
}

var _ ports.AuthService = (*Auth)(nil)

func New(users ports.UserRepository, sessions ports.SessionRepository, tokens ports.TokenService, passwords ports.PasswordHasher) *Auth {
	return &Auth{users, sessions, tokens, passwords}
}
func (s *Auth) Register(ctx context.Context, username, email, password string) (*domain.User, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	username = strings.TrimSpace(username)
	addr, err := mail.ParseAddress(email)
	if err != nil || addr.Address != email || len(email) > 255 {
		return nil, fault.Invalid("valid email required")
	}
	if len(username) < 3 || len(username) > 100 || len(password) < 8 || len(password) > 128 {
		return nil, fault.Invalid("username must be 3..100 characters; password 8..128 bytes")
	}
	hash, err := s.passwords.HashPassword(password)
	if err != nil {
		return nil, err
	}
	u := &domain.User{Username: username, Email: email, PasswordHash: hash}
	if err = s.users.Create(ctx, u); err != nil {
		return nil, err
	}
	return u, nil
}
func (s *Auth) Login(ctx context.Context, email, password, clientIP, userAgent string) (*domain.TokenPair, error) {
	if len(password) > 128 {
		return nil, domain.ErrInvalidCredentials
	}
	u, err := s.users.GetByEmail(ctx, strings.ToLower(strings.TrimSpace(email)))
	if err != nil {
		if !errors.Is(err, domain.ErrUserNotFound) {
			return nil, err
		}
		_ = s.passwords.ComparePassword(password, "00000000000000000000000000000000:0000000000000000000000000000000000000000000000000000000000000000")
		return nil, domain.ErrInvalidCredentials
	}
	if err = s.passwords.ComparePassword(password, u.PasswordHash); err != nil {
		return nil, domain.ErrInvalidCredentials
	}
	session, pair, err := s.newSession(u.ID, clientIP, userAgent)
	if err != nil {
		return nil, err
	}
	if err = s.sessions.Create(ctx, session); err != nil {
		return nil, err
	}
	return pair, nil
}
func refreshHash(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}
func (s *Auth) newSession(userID uuid.UUID, clientIP, userAgent string) (*domain.Session, *domain.TokenPair, error) {
	id := uuid.New()
	pair, err := s.tokens.GenerateTokenPair(userID, id)
	if err != nil {
		return nil, nil, err
	}
	session := &domain.Session{ID: id, UserID: userID, RefreshHash: refreshHash(pair.RefreshToken), ClientIP: clientIP, UserAgent: userAgent, CreatedAt: time.Now().UTC(), ExpiresAt: time.Now().UTC().Add(7 * 24 * time.Hour)}
	return session, pair, nil
}
func (s *Auth) RefreshToken(ctx context.Context, refresh, clientIP, userAgent string) (*domain.TokenPair, error) {
	if refresh == "" {
		return nil, domain.ErrUnauthorized
	}
	previous, err := s.sessions.GetByRefreshHash(ctx, refreshHash(refresh))
	if err != nil {
		return nil, err
	}
	next, pair, err := s.newSession(previous.UserID, clientIP, userAgent)
	if err != nil {
		return nil, err
	}
	if err = s.sessions.Rotate(ctx, previous, next); err != nil {
		return nil, err
	}
	return pair, nil
}
func (s *Auth) authenticate(ctx context.Context, token string) (*domain.Claims, error) {
	claims, err := s.tokens.ValidateAccessToken(token)
	if err != nil {
		return nil, err
	}
	session, err := s.sessions.GetByID(ctx, claims.SessionID)
	if err != nil {
		return nil, err
	}
	if session.UserID != claims.UserID || !session.ExpiresAt.After(time.Now()) {
		return nil, domain.ErrUnauthorized
	}
	return claims, nil
}
func (s *Auth) CurrentUser(ctx context.Context, token string) (*domain.User, error) {
	claims, err := s.authenticate(ctx, token)
	if err != nil {
		return nil, err
	}
	return s.users.GetByID(ctx, claims.UserID)
}
func (s *Auth) Logout(ctx context.Context, sessionID uuid.UUID) error {
	return s.sessions.Delete(ctx, sessionID)
}
func (s *Auth) LogoutAccess(ctx context.Context, token string) error {
	claims, err := s.authenticate(ctx, token)
	if err != nil {
		return err
	}
	return s.Logout(ctx, claims.SessionID)
}
