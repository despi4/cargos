package config

import (
	"fmt"
	"os"
	"strings"
)

type Config struct {
	Address     string
	DatabaseURL string
	InternalKey string
	JWTSecret   string
	URLs        map[string]string
}

func Env(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
func Load(name string) (Config, error) {
	cfg := Config{Address: Env("HTTP_ADDR", ":8080"), DatabaseURL: Env("DATABASE_URL", "postgres://cargos:cargos@localhost:5432/cargos_auth?sslmode=disable"), InternalKey: os.Getenv("INTERNAL_KEY"), JWTSecret: os.Getenv("JWT_SECRET"), URLs: map[string]string{}}
	if len(cfg.InternalKey) < 32 {
		return cfg, fmt.Errorf("INTERNAL_KEY must contain at least 32 bytes")
	}
	if name == "auth" && len(cfg.JWTSecret) < 32 {
		return cfg, fmt.Errorf("JWT_SECRET must contain at least 32 bytes")
	}
	for i, role := range []string{"auth", "inventory", "procurement", "demand", "prediction", "notification"} {
		cfg.URLs[role] = Env(strings.ToUpper(role)+"_URL", fmt.Sprintf("http://localhost:%d", 8081+i))
	}
	return cfg, nil
}
