package security

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"strings"
	"time"

	"github.com/depsi4/cargos/auth_service/internal/core/domain"
	"github.com/google/uuid"
)

type Tokens struct{ secret []byte }

func NewTokens(secret string) *Tokens { return &Tokens{secret: []byte(secret)} }
func (t *Tokens) GenerateTokenPair(userID, sessionID uuid.UUID) (*domain.TokenPair, error) {
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return nil, err
	}
	payload, _ := json.Marshal(domain.Claims{UserID: userID, SessionID: sessionID, ExpiresAt: time.Now().Add(15 * time.Minute).Unix()})
	prefix := base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"HS256","typ":"JWT"}`)) + "." + base64.RawURLEncoding.EncodeToString(payload)
	mac := hmac.New(sha256.New, t.secret)
	mac.Write([]byte(prefix))
	return &domain.TokenPair{AccessToken: prefix + "." + base64.RawURLEncoding.EncodeToString(mac.Sum(nil)), RefreshToken: base64.RawURLEncoding.EncodeToString(raw), TokenType: "Bearer", ExpiresIn: 900}, nil
}
func (t *Tokens) ValidateAccessToken(token string) (*domain.Claims, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return nil, domain.ErrUnauthorized
	}
	header, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil || string(header) != `{"alg":"HS256","typ":"JWT"}` {
		return nil, domain.ErrUnauthorized
	}
	mac := hmac.New(sha256.New, t.secret)
	mac.Write([]byte(parts[0] + "." + parts[1]))
	signature, err := base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil || !hmac.Equal(signature, mac.Sum(nil)) {
		return nil, domain.ErrUnauthorized
	}
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	var claims domain.Claims
	if err != nil || json.Unmarshal(payload, &claims) != nil || claims.ExpiresAt <= time.Now().Unix() || claims.UserID == uuid.Nil || claims.SessionID == uuid.Nil {
		return nil, domain.ErrUnauthorized
	}
	return &claims, nil
}
