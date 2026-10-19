package security

import (
	"crypto/pbkdf2"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"strings"

	"github.com/depsi4/cargos/auth_service/internal/core/domain"
)

type PasswordHasher struct{}

func (PasswordHasher) HashPassword(password string) (string, error) {
	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}
	key, err := pbkdf2.Key(sha256.New, password, salt, 120000, 32)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(salt) + ":" + hex.EncodeToString(key), nil
}
func (PasswordHasher) ComparePassword(password, hash string) error {
	parts := strings.Split(hash, ":")
	if len(parts) != 2 {
		return domain.ErrInvalidCredentials
	}
	salt, err := hex.DecodeString(parts[0])
	if err != nil {
		return domain.ErrInvalidCredentials
	}
	expected, err := hex.DecodeString(parts[1])
	if err != nil {
		return domain.ErrInvalidCredentials
	}
	key, err := pbkdf2.Key(sha256.New, password, salt, 120000, 32)
	if err != nil || subtle.ConstantTimeCompare(key, expected) != 1 {
		return domain.ErrInvalidCredentials
	}
	return nil
}
