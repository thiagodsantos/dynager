package usecase

import (
	"context"

	"github.com/thiagosantos/dynamodb-manager/backend/internal/dynamo"
)

func (s *Service) ListTables(ctx context.Context, id string) ([]string, error) {
	connection, err := s.getConnection(ctx, id)
	if err != nil {
		return nil, err
	}
	tables, err := s.dynamo.ListTables(ctx, toDynamoConnection(connection))
	if err != nil {
		return nil, upstream(err)
	}
	return tables, nil
}

func (s *Service) DescribeTable(ctx context.Context, id, table string) (dynamo.TableDescription, error) {
	connection, err := s.getConnection(ctx, id)
	if err != nil {
		return dynamo.TableDescription{}, err
	}
	if err := validateTable(table); err != nil {
		return dynamo.TableDescription{}, err
	}
	description, err := s.dynamo.DescribeTable(ctx, toDynamoConnection(connection), table)
	if err != nil {
		return dynamo.TableDescription{}, upstream(err)
	}
	return description, nil
}

func (s *Service) ListItems(ctx context.Context, id, table string, limit int32, token string, filter *dynamo.Filter) (dynamo.Page, error) {
	connection, err := s.getConnection(ctx, id)
	if err != nil {
		return dynamo.Page{}, err
	}
	if err := validateTable(table); err != nil {
		return dynamo.Page{}, err
	}
	if limit < MinPageSize || limit > MaxPageSize {
		return dynamo.Page{}, invalid("Limit must be between 1 and 100")
	}
	if filter != nil {
		if filter.Attribute == "" || len(filter.Attribute) > MaxFilterAttributeSize {
			return dynamo.Page{}, invalid("Filter attribute must be between 1 and 255 characters")
		}
		if filter.Operator != dynamo.FilterOperatorEqual {
			return dynamo.Page{}, invalid("Filter operator must be eq")
		}
		if filter.Value == nil {
			return dynamo.Page{}, invalid("Filter value is required")
		}
	}
	start, err := decodeToken(token)
	if err != nil {
		return dynamo.Page{}, err
	}
	page, err := s.dynamo.Scan(ctx, toDynamoConnection(connection), table, limit, start, filter)
	if err != nil {
		return dynamo.Page{}, upstream(err)
	}
	return page, nil
}
