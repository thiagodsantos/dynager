package main

import (
	"context"
	"errors"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/thiagosantos/dynamodb-manager/backend/internal/app"
	"github.com/thiagosantos/dynamodb-manager/backend/internal/config"
	"github.com/thiagosantos/dynamodb-manager/backend/internal/dynamo"
	"github.com/thiagosantos/dynamodb-manager/backend/internal/httpapi"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	address := env("DYNAMODB_MANAGER_ADDR", "127.0.0.1:8080")
	if err := validateLoopbackAddress(address); err != nil {
		logger.Error("server address must use loopback", "error", "invalid listen address")
		os.Exit(1)
	}
	configPath, err := configurationPath()
	if err != nil {
		logger.Error("could not determine configuration location")
		os.Exit(1)
	}
	repository, err := config.NewFileRepository(configPath)
	if err != nil {
		logger.Error("could not initialize connection configuration")
		os.Exit(1)
	}
	if _, err := repository.List(context.Background()); err != nil {
		logger.Error("connection configuration is invalid or unreadable")
		os.Exit(1)
	}
	origins := parseOrigins(env("DYNAMODB_MANAGER_CORS_ORIGINS", "http://localhost:5173,http://127.0.0.1:5173"))
	service := app.NewService(repository, dynamo.NewAWSClient())
	server := &http.Server{
		Addr:              address,
		Handler:           httpapi.NewRouter(service, origins, logger),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      60 * time.Second,
		IdleTimeout:       60 * time.Second,
	}
	listener, err := net.Listen("tcp", address)
	if err != nil {
		logger.Error("could not start HTTP server", "address", address, "error", err)
		os.Exit(1)
	}
	stopped := make(chan os.Signal, 1)
	signal.Notify(stopped, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-stopped
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := server.Shutdown(ctx); err != nil {
			logger.Error("graceful shutdown failed")
		}
	}()
	logger.Info("DynamoDB Manager API listening", "address", address, "config", configPath)
	if err := server.Serve(listener); err != nil && !errors.Is(err, http.ErrServerClosed) {
		logger.Error("HTTP server stopped unexpectedly")
		os.Exit(1)
	}
}

func configurationPath() (string, error) {
	if configured := os.Getenv("DYNAMODB_MANAGER_CONFIG"); configured != "" {
		return filepath.Clean(configured), nil
	}
	base, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(base, "dynamodb-manager", "connections.json"), nil
}

func parseOrigins(value string) []string {
	var origins []string
	for _, origin := range strings.Split(value, ",") {
		if origin = strings.TrimSpace(origin); origin != "" && origin != "*" {
			origins = append(origins, origin)
		}
	}
	return origins
}

func validateLoopbackAddress(address string) error {
	host, _, err := net.SplitHostPort(address)
	if err != nil {
		return err
	}
	if host == "localhost" {
		return nil
	}
	ip := net.ParseIP(host)
	if ip == nil || !ip.IsLoopback() {
		return errors.New("address must be loopback")
	}
	return nil
}

func env(key, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}
	return fallback
}
