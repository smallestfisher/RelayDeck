package config

import (
	"os"
	"time"
)

type Config struct {
	HTTPAddr              string
	AppSecret             string
	GatewayRequestTimeout time.Duration
}

func Load() Config {
	return Config{
		HTTPAddr:              envOrDefault("HTTP_ADDR", ":8080"),
		AppSecret:             envOrDefault("APP_SECRET", "dev-secret"),
		GatewayRequestTimeout: durationOrDefault("GATEWAY_REQUEST_TIMEOUT", 30*time.Second),
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
