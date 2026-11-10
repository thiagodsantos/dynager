package config

import (
	"os"
	"strings"
)

const (
	EnvAddress     = "DYNAGER_ADDR"
	EnvCORSOrigins = "DYNAGER_CORS_ORIGINS"
	EnvConfigPath  = "DYNAGER_CONFIG"

	DefaultAddress     = "127.0.0.1:8080"
	DefaultCORSOrigins = "http://localhost:5173,http://127.0.0.1:5173"
)

func Env(key, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}
	return fallback
}
