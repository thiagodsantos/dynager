package usecase

import (
	"context"

	"github.com/thiagosantos/dynamodb-manager/backend/internal/dynamo"
)

func (s *Service) ListConnections(ctx context.Context) ([]Connection, error) {
	return s.repository.List(ctx)
}

func (s *Service) CreateConnection(ctx context.Context, input ConnectionInput) (Connection, error) {
	clean, err := ValidateConnectionInput(input)
	if err != nil {
		return Connection{}, err
	}
	id, err := newID()
	if err != nil {
		return Connection{}, internal("Could not create connection", err)
	}
	return s.repository.Create(ctx, Connection{ID: id, Name: clean.Name, Region: clean.Region, Profile: clean.Profile, Endpoint: clean.Endpoint})
}

func (s *Service) TestConnection(ctx context.Context, input ConnectionInput) error {
	clean, err := ValidateConnectionInput(input)
	if err != nil {
		return err
	}
	return s.dynamo.Test(ctx, dynamo.Connection{Region: clean.Region, Profile: clean.Profile, Endpoint: clean.Endpoint})
}

func (s *Service) DeleteConnection(ctx context.Context, id string) error {
	if err := validateConnectionID(id); err != nil {
		return err
	}
	return s.repository.Delete(ctx, id)
}

func validateConnectionID(id string) error {
	if id == "" {
		return invalid("Connection ID is required")
	}
	return nil
}

func (s *Service) getConnection(ctx context.Context, id string) (Connection, error) {
	if err := validateConnectionID(id); err != nil {
		return Connection{}, err
	}
	connection, err := s.repository.Get(ctx, id)
	if err != nil {
		return Connection{}, err
	}
	return connection, nil
}
