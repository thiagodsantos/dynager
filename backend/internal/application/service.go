package application

import (
	"context"

	"github.com/thiagosantos/dynamodb-manager/backend/internal/domain"
)

type Connection = domain.Connection
type ConnectionInput = domain.ConnectionInput

// Repository is the outbound port for saved connection configuration.
type Repository interface {
	List(context.Context) ([]domain.Connection, error)
	Get(context.Context, string) (domain.Connection, error)
	Create(context.Context, domain.Connection) (domain.Connection, error)
	Delete(context.Context, string) error
}

// Dynamo is the outbound port for DynamoDB operations used by application services.
type Dynamo interface {
	Test(context.Context, domain.ConnectionSettings) error
	ListTables(context.Context, domain.ConnectionSettings) ([]string, error)
	DescribeTable(context.Context, domain.ConnectionSettings, string) (domain.TableDescription, error)
	Scan(context.Context, domain.ConnectionSettings, string, int32, domain.Item, *domain.Filter) (domain.Page, error)
	PutItem(context.Context, domain.ConnectionSettings, string, domain.Item) error
	DeleteItem(context.Context, domain.ConnectionSettings, string, domain.Item) error
}

type Service struct {
	repository Repository
	dynamo     Dynamo
}

func NewService(repository Repository, client Dynamo) *Service {
	return &Service{repository: repository, dynamo: client}
}
