package application

import (
	"context"

	"github.com/thiagosantos/dynamodb-manager/backend/internal/domain"
)

func (s *Service) PutItem(ctx context.Context, id, table string, key, item domain.Item) error {
	connection, description, err := s.tableForWrite(ctx, id, table)
	if err != nil {
		return err
	}
	if err := domain.ValidatePrimaryKey(key, item, description); err != nil {
		return err
	}
	if len(item) == 0 {
		return invalid("Item must not be empty")
	}
	if err := s.dynamo.PutItem(ctx, toConnectionSettings(connection), table, item); err != nil {
		return upstream(err)
	}
	return nil
}

func (s *Service) DeleteItem(ctx context.Context, id, table string, key domain.Item) error {
	connection, description, err := s.tableForWrite(ctx, id, table)
	if err != nil {
		return err
	}
	if err := domain.ValidatePrimaryKey(key, nil, description); err != nil {
		return err
	}
	if err := s.dynamo.DeleteItem(ctx, toConnectionSettings(connection), table, key); err != nil {
		return upstream(err)
	}
	return nil
}

func (s *Service) tableForWrite(ctx context.Context, id, table string) (Connection, domain.TableDescription, error) {
	connection, err := s.tableConnection(ctx, id, table)
	if err != nil {
		return Connection{}, domain.TableDescription{}, err
	}
	description, err := s.tableDescription(ctx, connection, table)
	if err != nil {
		return Connection{}, domain.TableDescription{}, err
	}
	return connection, description, nil
}
