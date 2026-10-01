package service

import (
	"github.com/despi4/cargos/auth_service/internal/core/ports"
	"golang.org/x/crypto/bcrypt"
)

type PassHasher struct{}

func NewPasswordHasher() *PassHasher {
	return &PassHasher{}
}

var _ ports.PasswordHasher = (*PassHasher)(nil)

func(p *PassHasher) HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		return "", err
	}

	return string(hash), nil
}

func(p *PassHasher) ComparePassword(password string, hash string) (error) {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	if err != nil {
		return err
	}

	return nil
}
