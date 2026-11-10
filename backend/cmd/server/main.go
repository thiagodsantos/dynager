package main

import (
	"context"
	"errors"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/rs/zerolog"
	"github.com/thiagosantos/dynamodb-manager/backend/internal/application"
	"github.com/thiagosantos/dynamodb-manager/backend/internal/infrastructure/config"
	dynamodb "github.com/thiagosantos/dynamodb-manager/backend/internal/infrastructure/dynamodb"
	"github.com/thiagosantos/dynamodb-manager/backend/internal/infrastructure/httpapi"
)

func main() {
	logger := zerolog.New(zerolog.ConsoleWriter{Out: os.Stdout, NoColor: false}).With().Timestamp().Logger()
	address := config.Env(config.EnvAddress, config.DefaultAddress)
	if err := config.ValidateLoopbackAddress(address); err != nil {
		logger.Error().Msg("server address must use loopback")
		os.Exit(1)
	}
	configPath, err := config.ConfigurationPath()
	if err != nil {
		logger.Error().Err(err).Msg("could not determine configuration location")
		os.Exit(1)
	}
	repository, err := config.NewFileRepository(configPath)
	if err != nil {
		logger.Error().Err(err).Msg("could not initialize connection configuration")
		os.Exit(1)
	}
	if _, err := repository.List(context.Background()); err != nil {
		logger.Error().Err(err).Msg("connection configuration is invalid or unreadable")
		os.Exit(1)
	}
	origins := httpapi.ParseOrigins(config.Env(config.EnvCORSOrigins, config.DefaultCORSOrigins))
	service := application.NewService(repository, dynamodb.NewAWSClient())
	server := &http.Server{
		Addr:              address,
		Handler:           httpapi.NewRouter(service, origins, &logger),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      60 * time.Second,
		IdleTimeout:       60 * time.Second,
	}
	listener, err := net.Listen("tcp", address)
	if err != nil {
		logger.Error().Err(err).Str("address", address).Msg("could not start HTTP server")
		os.Exit(1)
	}
	stopped := make(chan os.Signal, 1)
	signal.Notify(stopped, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-stopped
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := server.Shutdown(ctx); err != nil {
			logger.Error().Err(err).Msg("graceful shutdown failed")
		}
	}()
	logger.Info().Str("address", address).Str("config", configPath).Msg("Dynager API listening")
	if err := server.Serve(listener); err != nil && !errors.Is(err, http.ErrServerClosed) {
		logger.Error().Err(err).Msg("HTTP server stopped unexpectedly")
		os.Exit(1)
	}
}
