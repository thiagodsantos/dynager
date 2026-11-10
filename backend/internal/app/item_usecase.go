package app

import (
	"context"

	"github.com/thiagosantos/dynamodb-manager/backend/internal/dynamo"
)

func (s *Service) PutItem(ctx context.Context, id, table string, key, item dynamo.Item) error {
	connection, description, err := s.tableForWrite(ctx, id, table)
	if err != nil {
		return err
	}
	if err := validateKey(key, item, description); err != nil {
		return err
	}
	if len(item) == 0 {
		return invalid("Item must not be empty")
	}
	if err := s.dynamo.PutItem(ctx, toDynamoConnection(connection), table, item); err != nil {
		return upstream(err)
	}
	return nil
}

func (s *Service) DeleteItem(ctx context.Context, id, table string, key dynamo.Item) error {
	connection, description, err := s.tableForWrite(ctx, id, table)
	if err != nil {
		return err
	}
	if err := validateKey(key, nil, description); err != nil {
		return err
	}
	if err := s.dynamo.DeleteItem(ctx, toDynamoConnection(connection), table, key); err != nil {
		return upstream(err)
	}
	return nil
}

func (s *Service) tableForWrite(ctx context.Context, id, table string) (Connection, dynamo.TableDescription, error) {
	connection, err := s.getConnection(ctx, id)
	if err != nil {
		return Connection{}, dynamo.TableDescription{}, err
	}
	if err := validateTable(table); err != nil {
		return Connection{}, dynamo.TableDescription{}, err
	}
	description, err := s.dynamo.DescribeTable(ctx, toDynamoConnection(connection), table)
	if err != nil {
		return Connection{}, dynamo.TableDescription{}, upstream(err)
	}
	return connection, description, nil
}
