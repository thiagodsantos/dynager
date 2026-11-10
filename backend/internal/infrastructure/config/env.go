package config

import (
	"os"
	"strings"
)

const (
	EnvAddress     = "DYNAMODB_MANAGER_ADDR"
	EnvCORSOrigins = "DYNAMODB_MANAGER_CORS_ORIGINS"
	EnvConfigPath  = "DYNAMODB_MANAGER_CONFIG"

	DefaultAddress     = "127.0.0.1:8080"
	DefaultCORSOrigins = "http://localhost:5173,http://127.0.0.1:5173"
)

func Env(key, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}
	return fallback
}
