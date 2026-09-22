package config

import "time"

type AuthConfig struct {
	GRPCPort        string
	DatabaseURL     string
	JWTSecret       string
	AccessTokenTTL  time.Duration
	RefreshTokenTTL time.Duration
}

func LoadAuthConfig() AuthConfig {
	return AuthConfig{
		GRPCPort:        getEnv("GRPC_PORT", "50051"),
		DatabaseURL:     getEnv("DATABASE_URL", "postgres://postgres:postgres@localhost:54322/listenly?sslmode=disable"),
		JWTSecret:       getEnv("JWT_SECRET", "change-me-in-production"),
		AccessTokenTTL:  15 * time.Minute,
		RefreshTokenTTL: 7 * 24 * time.Hour,
	}
}
