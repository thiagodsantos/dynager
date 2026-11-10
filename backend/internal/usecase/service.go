package usecase

import (
	"context"

	"github.com/thiagosantos/dynamodb-manager/backend/internal/dynamo"
)

type Connection struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Region   string `json:"region"`
	Profile  string `json:"profile,omitempty"`
	Endpoint string `json:"endpoint,omitempty"`
}

type ConnectionInput struct {
	Name     string
	Region   string
	Profile  string
	Endpoint string
}

type Repository interface {
	List(context.Context) ([]Connection, error)
	Get(context.Context, string) (Connection, error)
	Create(context.Context, Connection) (Connection, error)
	Delete(context.Context, string) error
}

type Dynamo interface {
	Test(context.Context, dynamo.Connection) error
	ListTables(context.Context, dynamo.Connection) ([]string, error)
	DescribeTable(context.Context, dynamo.Connection, string) (dynamo.TableDescription, error)
	Scan(context.Context, dynamo.Connection, string, int32, dynamo.Item, *dynamo.Filter) (dynamo.Page, error)
	PutItem(context.Context, dynamo.Connection, string, dynamo.Item) error
	DeleteItem(context.Context, dynamo.Connection, string, dynamo.Item) error
}

type Service struct {
	repository Repository
	dynamo     Dynamo
}

func NewService(repository Repository, client Dynamo) *Service {
	return &Service{repository: repository, dynamo: client}
}
