package service

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/despi4/cargos/auth_service/internal/core/domain"
	"github.com/despi4/cargos/auth_service/internal/core/ports"
	"github.com/google/uuid"
)

type AuthSvc struct {
	userRepo ports.UserRepository
	sessionRepo ports.SessionRepository
	passwordHasher ports.PasswordHasher
	tokenService ports.TokenService
	
}

func NewAuthSvc(
	repo ports.UserRepository,
	passwordHasher ports.PasswordHasher,
	tokenSvc ports.TokenService,
	sessionRepo ports.SessionRepository,
	) *AuthSvc {
	return &AuthSvc{
		userRepo: repo,
		passwordHasher: passwordHasher,
		tokenService: tokenSvc,
		sessionRepo: sessionRepo,
	}
}

var (
	_ ports.AuthService = (*AuthSvc)(nil)

	// Простые (мини) регулярки для конкретных проверок пароля
	hasUpper   = regexp.MustCompile(`[A-Z]`)
	hasLower   = regexp.MustCompile(`[a-z]`)
	hasDigit   = regexp.MustCompile(`[0-9]`)
	hasSpecial = regexp.MustCompile(`[@$!%*?&#]`)
	
	// Базовая проверка структуры email
	emailRegex = regexp.MustCompile(`^[^\s@]+@[^\s@]+\.[^\s@]+$`)
)

func (s *AuthSvc) Register(ctx context.Context, username, email, password string) (*domain.User, error) {
	if err := isValidUsername(&username); err != nil {
		return nil, err
	}

	if err := isValidEmail(email); err != nil {
		return nil, err
	}

	if err := isValidPassword(password); err != nil {
		return nil, err
	}

	existingUser, err := s.userRepo.GetByEmail(ctx, email)
	if err == nil && existingUser != nil {
		return nil, domain.ErrUserAlreadyExists
	}

	hash, err := s.passwordHasher.HashPassword(password)
	if err != nil {
		return nil, fmt.Errorf("generating hash error: %w", err)
	}

	user := domain.User{
		Username:     username,
		Email:        email,
		PasswordHash: hash,
	}

	if err := s.userRepo.Create(ctx, &user); err != nil {
		return nil, fmt.Errorf("user creating error: %w", err)
	}

	return &user, nil
}

func (s *AuthSvc) Login(ctx context.Context, email, password, clientIP, userAgent string) (*domain.TokenPair, error) {
	clientIP = strings.TrimSpace(clientIP)
	if clientIP == "" {
		return nil, domain.ErrInvalidClientIP
	}

	userAgent = strings.TrimSpace(userAgent)
	if userAgent == "" {
		return nil, domain.ErrInvalidUserAgent
	}

	if err := isValidEmail(email); err != nil {
		return nil, err
	}

	if err := isValidPassword(password); err != nil {
		return nil, err
	}

	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return nil, domain.ErrInvalidCredentials
	}

	if err := s.passwordHasher.ComparePassword(password, user.PasswordHash); err != nil {
		return nil, domain.ErrInvalidCredentials
	}

	tokenPair, err := s.tokenService.GenerateTokenPair(user)
	if err != nil {
		return nil, fmt.Errorf("generate tokens error: %w", err)
	}

	session := domain.Session{
		ID:           uuid.New(),
		UserID:       user.ID,
		RefreshToken: tokenPair.RefreshToken,
		ClientIP:     clientIP,
		UserAgent:    userAgent,
		IsBlocked:    false,
		ExpiresAt:    time.Now().Add(time.Hour * 24 * 7),
		CreatedAt:    time.Now(),
	}

	if err := s.sessionRepo.Create(ctx, &session); err != nil {
		return nil, fmt.Errorf("create session error: %w", err)
	}

	return tokenPair, nil
}

func (s *AuthSvc) RefreshToken(ctx context.Context, refreshToken, clientIP, userAgent string) (*domain.TokenPair, error) {
	refreshToken = strings.TrimSpace(refreshToken)
	if refreshToken == "" {
		return nil, domain.ErrInvalidRefreshToken
	}

	clientIP = strings.TrimSpace(clientIP)
	if clientIP == "" {
		return nil, domain.ErrInvalidClientIP
	}

	userAgent = strings.TrimSpace(userAgent)
	if userAgent == "" {
		return nil, domain.ErrInvalidUserAgent
	}

	session, err := s.sessionRepo.GetByRefreshToken(ctx, refreshToken)
	if err != nil {
		return nil, domain.ErrSessionNotFound
	}

	if session.IsBlocked {
		return nil, domain.ErrSessionBlocked
	}

	if time.Now().After(session.ExpiresAt) {
		_ = s.sessionRepo.Delete(ctx, session.ID)
		return nil, domain.ErrSessionExpired
	}

	if err := s.sessionRepo.Delete(ctx, session.ID); err != nil {
		return nil, fmt.Errorf("delete old session error: %w", err)
	}

	user, err := s.userRepo.GetByID(ctx, session.UserID)
	if err != nil {
		return nil, domain.ErrUserNotFound
	}

	newTokenPair, err := s.tokenService.GenerateTokenPair(user)
	if err != nil {
		return nil, fmt.Errorf("generate token pair error: %w", err)
	}

	newSession := domain.Session{
		ID:           uuid.New(),
		UserID:       user.ID,
		RefreshToken: newTokenPair.RefreshToken,
		ClientIP:     clientIP,
		UserAgent:    userAgent,
		IsBlocked:    false,
		ExpiresAt:    time.Now().Add(time.Hour * 24 * 7),
		CreatedAt:    time.Now(),
	}

	if err := s.sessionRepo.Create(ctx, &newSession); err != nil {
		return nil, fmt.Errorf("create session error: %w", err)
	}

	return newTokenPair, nil
}

func (s *AuthSvc) Logout(ctx context.Context, sessionID uuid.UUID) error {
	return s.sessionRepo.Delete(ctx, sessionID)
}

func isValidEmail(email string) error {
	email = strings.TrimSpace(email)
	if len(email) < 6 {
		return domain.ErrEmailTooShort
	}

	if len(email) > 254 {
		return domain.ErrEmailTooLong
	}

	if !emailRegex.MatchString(email) {
		return domain.ErrEmailInvalidFormat
	}

	return nil
}

func isValidPassword(password string) error {
	if len(password) < 8 {
		return domain.ErrPasswordTooShort
	}

	if len(password) > 32 {
		return domain.ErrPasswordTooLong
	}

	if !hasDigit.MatchString(password) {
		return domain.ErrPasswordNoDigit
	}

	if !hasLower.MatchString(password) {
		return domain.ErrPasswordNoLower
	}

	if !hasUpper.MatchString(password) {
		return domain.ErrPasswordNoUpper
	}

	if !hasSpecial.MatchString(password) {
		return domain.ErrPasswordNoSpecial
	}

	return nil
}

func isValidUsername(username *string) error {
	*username = strings.TrimSpace(*username)

	if len(*username) < 3 {
		return domain.ErrUsernameTooShort
	}

	if len(*username) > 25 {
		return domain.ErrUsernameTooLong
	}

	return nil
}