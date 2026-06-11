package config

import (
	"os"
	"time"
)

type Config struct {
	HTTPAddr               string
	AppSecret              string
	DatabaseURL            string
	RedisURL               string
	UpstreamSecretKey      string
	GatewayRequestTimeout  time.Duration
	BootstrapOwnerEmail    string
	BootstrapOwnerPassword string
}

func Load() Config {
	LoadDotEnv()
	return Config{
		HTTPAddr:               envOrDefault("HTTP_ADDR", ":8080"),
		AppSecret:              envOrDefault("APP_SECRET", "dev-secret"),
		DatabaseURL:            envOrDefault("DATABASE_URL", ""),
		RedisURL:               envOrDefault("REDIS_URL", ""),
		UpstreamSecretKey:      envOrDefault("APP_UPSTREAM_SECRET_KEY", ""),
		GatewayRequestTimeout:  durationOrDefault("GATEWAY_REQUEST_TIMEOUT", 30*time.Second),
		BootstrapOwnerEmail:    envOrDefault("APP_BOOTSTRAP_OWNER_EMAIL", "owner@example.com"),
		BootstrapOwnerPassword: envOrDefault("APP_BOOTSTRAP_OWNER_PASSWORD", "change-me"),
	}
}

func envOrDefault(key string, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func durationOrDefault(key string, fallback time.Duration) time.Duration {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	parsed, err := time.ParseDuration(value)
	if err != nil {
		return fallback
	}
	return parsed
}
