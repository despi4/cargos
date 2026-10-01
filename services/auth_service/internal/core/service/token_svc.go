package service

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"time"

	"github.com/despi4/cargos/auth_service/internal/core/domain"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type JWTService struct {
	jwtSecret []byte
	accessTTL time.Duration
}

func NewJwtService(jwtSecret []byte, accessTTL time.Duration) *JWTService {
	return &JWTService{
		jwtSecret: jwtSecret,
		accessTTL: accessTTL,
	}
}

func (s *JWTService) GenerateTokenPair(user *domain.User) (*domain.TokenPair, error) {
	now := time.Now()

	// claims - это то что мы кладем в часть токена Payload
	claims := jwt.RegisteredClaims{
		Subject: user.ID.String(),
		IssuedAt: jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(now.Add(s.accessTTL)),
	}

    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString(s.jwtSecret)
	if err != nil {
		return nil, fmt.Errorf("sign access token: %w", err)
	}

	// Случайная строка
	refreshToken, err := generateOpaqueToken(32)
	if err != nil {
		return nil, fmt.Errorf("generate refresh token: %w", err)
	}
	
	tokenPair := domain.TokenPair{
		AccessToken: signedToken,
		RefreshToken: refreshToken,
	}

	return &tokenPair, nil
}

func (s *JWTService) ValidateAccessToken(accessToken string) (uuid.UUID, error) {
	token, err := jwt.ParseWithClaims(accessToken, &jwt.RegisteredClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		return s.jwtSecret, nil
	})

	if err != nil {
		// jwt.Parse возвращает детальные ошибки, можно их проверить
		if errors.Is(err, jwt.ErrTokenExpired) {
			return uuid.Nil, domain.ErrExpiredToken
		}
		return uuid.Nil, domain.ErrInvalidToken
	}

	// Извлекаем Payload (Claims)
	claims, ok := token.Claims.(*jwt.RegisteredClaims)
	if !ok || !token.Valid {
		return uuid.Nil, domain.ErrInvalidToken
	}

	// Достаем ID пользователя из Subject
	userID, err := uuid.Parse(claims.Subject)
	if err != nil {
		return uuid.Nil, fmt.Errorf("invalid user ID in token: %w", err)
	}

	return userID, nil
}

// generateOpaqueToken создает криптографически безопасную случайную строку.
// Используем crypto/rand, а не math/rand, так как это токены безопасности!
func generateOpaqueToken(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	// base64.URLEncoding делает строку безопасной для передачи в URL/Headers
	return base64.RawURLEncoding.EncodeToString(bytes), nil
}