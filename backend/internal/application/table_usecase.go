package application

import (
	"context"

	"github.com/thiagosantos/dynamodb-manager/backend/internal/domain"
)

func (s *Service) ListTables(ctx context.Context, id string) ([]string, error) {
	connection, err := s.getConnection(ctx, id)
	if err != nil {
		return nil, err
	}
	tables, err := s.dynamo.ListTables(ctx, toConnectionSettings(connection))
	if err != nil {
		return nil, upstream(err)
	}
	return tables, nil
}

func (s *Service) DescribeTable(ctx context.Context, id, table string) (domain.TableDescription, error) {
	connection, err := s.tableConnection(ctx, id, table)
	if err != nil {
		return domain.TableDescription{}, err
	}
	return s.tableDescription(ctx, connection, table)
}

func (s *Service) tableConnection(ctx context.Context, id, table string) (Connection, error) {
	connection, err := s.getConnection(ctx, id)
	if err != nil {
		return Connection{}, err
	}
	if err := domain.ValidateTableName(table); err != nil {
		return Connection{}, err
	}
	return connection, nil
}

func (s *Service) tableDescription(ctx context.Context, connection Connection, table string) (domain.TableDescription, error) {
	description, err := s.dynamo.DescribeTable(ctx, toConnectionSettings(connection), table)
	if err != nil {
		return domain.TableDescription{}, upstream(err)
	}
	return description, nil
}

func (s *Service) ListItems(ctx context.Context, id, table string, limit int32, token string, filter *domain.Filter) (domain.Page, error) {
	connection, err := s.tableConnection(ctx, id, table)
	if err != nil {
		return domain.Page{}, err
	}
	if limit < MinPageSize || limit > MaxPageSize {
		return domain.Page{}, invalid("Limit must be between 1 and 100")
	}
	if filter != nil {
		if filter.Attribute == "" || len(filter.Attribute) > MaxFilterAttributeSize {
			return domain.Page{}, invalid("Filter attribute must be between 1 and 255 characters")
		}
		if filter.Operator != domain.FilterOperatorEqual {
			return domain.Page{}, invalid("Filter operator must be eq")
		}
		if filter.Value.Kind == "" {
			return domain.Page{}, invalid("Filter value is required")
		}
	}
	start, err := decodeToken(token)
	if err != nil {
		return domain.Page{}, err
	}
	page, err := s.dynamo.Scan(ctx, toConnectionSettings(connection), table, limit, start, filter)
	if err != nil {
		return domain.Page{}, upstream(err)
	}
	return page, nil
}
