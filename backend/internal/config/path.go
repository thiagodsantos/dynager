package config

import (
	"os"
	"path/filepath"
)

const (
	ConfigDirectory = "dynamodb-manager"
	ConfigFile      = "connections.json"
)

func ConfigurationPath() (string, error) {
	if configured := os.Getenv(EnvConfigPath); configured != "" {
		return filepath.Clean(configured), nil
	}
	base, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(base, ConfigDirectory, ConfigFile), nil
}
