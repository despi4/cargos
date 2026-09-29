package domain

import (
	"time"

	"github.com/google/uuid"
)

// TokenPair — то, что мы отдаем клиенту после логина
type TokenPair struct {
	AccessToken  string
	RefreshToken string
}

// Session — сущность сессии, которую мы будем хранить в Redis на сервере.
// Она привязывается к Refresh токену.
type Session struct {
	ID           uuid.UUID // Уникальный ID сессии
	UserID       uuid.UUID
	RefreshToken string    // Сам токен (или его хэш в целях безопасности)
	ClientIP     string    // Полезно для аналитики и безопасности
	UserAgent    string    // Инфа об устройстве (Chrome, iOS App)
	IsBlocked    bool      // Флаг принудительного отзыва сессии
	ExpiresAt    time.Time
	CreatedAt    time.Time
}