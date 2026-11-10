package application

import (
	"context"
	"errors"
	"testing"

	"github.com/thiagosantos/dynamodb-manager/backend/internal/domain"
)

type memoryRepository struct{ connections map[string]Connection }

func (r *memoryRepository) List(context.Context) ([]Connection, error) {
	result := make([]Connection, 0, len(r.connections))
	for _, connection := range r.connections {
		result = append(result, connection)
	}
	return result, nil
}
func (r *memoryRepository) Get(_ context.Context, id string) (Connection, error) {
	connection, ok := r.connections[id]
	if !ok {
		return Connection{}, notFound("Connection not found")
	}
	return connection, nil
}
func (r *memoryRepository) Create(_ context.Context, connection Connection) (Connection, error) {
	r.connections[connection.ID] = connection
	return connection, nil
}
func (r *memoryRepository) Delete(_ context.Context, id string) error {
	if _, ok := r.connections[id]; !ok {
		return notFound("Connection not found")
	}
	delete(r.connections, id)
	return nil
}

type fakeDynamo struct {
	testErr     error
	testCalls   int
	lastFilter  *domain.Filter
	lastKey     domain.Item
	lastItem    domain.Item
	deleted     bool
	description domain.TableDescription
}

func (d *fakeDynamo) Test(context.Context, domain.ConnectionSettings) error {
	d.testCalls++
	return d.testErr
}
func (*fakeDynamo) ListTables(context.Context, domain.ConnectionSettings) ([]string, error) {
	return []string{"records"}, nil
}
func (d *fakeDynamo) DescribeTable(context.Context, domain.ConnectionSettings, string) (domain.TableDescription, error) {
	return d.description, nil
}
func (d *fakeDynamo) Scan(_ context.Context, _ domain.ConnectionSettings, _ string, _ int32, _ domain.Item, filter *domain.Filter) (domain.Page, error) {
	d.lastFilter = filter
	return domain.Page{Items: []domain.Item{}, Count: 0, ScannedCount: 4}, nil
}
func (d *fakeDynamo) PutItem(_ context.Context, _ domain.ConnectionSettings, _ string, item domain.Item) error {
	d.lastItem = item
	return nil
}
func (d *fakeDynamo) DeleteItem(_ context.Context, _ domain.ConnectionSettings, _ string, key domain.Item) error {
	d.lastKey = key
	d.deleted = true
	return nil
}

func setupService() (*Service, *memoryRepository, *fakeDynamo) {
	repository := &memoryRepository{connections: map[string]Connection{"conn": {ID: "conn", Name: "local", Region: "us-east-1"}}}
	client := &fakeDynamo{description: domain.TableDescription{PartitionKey: domain.KeyAttribute{Name: "id", Type: "S"}, SortKey: &domain.KeyAttribute{Name: "version", Type: "N"}}}
	return NewService(repository, client), repository, client
}

func TestCreateConnectionValidatesAndPersistsOnlyConfiguration(t *testing.T) {
	service, repository, _ := setupService()
	connection, err := service.CreateConnection(context.Background(), ConnectionInput{Name: " local ", Region: "us-east-1", Profile: "dev", Endpoint: "http://localhost:8000"})
	if err != nil {
		t.Fatal(err)
	}
	if connection.ID == "" || connection.Name != "local" || connection.Region != "us-east-1" || connection.Profile != "dev" {
		t.Fatalf("unexpected connection: %#v", connection)
	}
	if _, exists := repository.connections[connection.ID]; !exists {
		t.Fatal("connection was not persisted")
	}
	if _, err := service.CreateConnection(context.Background(), ConnectionInput{Name: "", Region: "us-east-1"}); err == nil {
		t.Fatal("expected invalid name")
	}
	if _, err := service.CreateConnection(context.Background(), ConnectionInput{Name: "x", Region: "us-east-1", Endpoint: "file:///etc/passwd"}); err == nil {
		t.Fatal("expected invalid endpoint")
	}
	for _, endpoint := range []string{"https://user:password@example.test", "https://example.test/path?token=secret"} {
		if _, err := service.CreateConnection(context.Background(), ConnectionInput{Name: "x", Region: "us-east-1", Endpoint: endpoint}); err == nil {
			t.Errorf("expected unsafe endpoint %q to fail", endpoint)
		}
	}
}

func TestTestConnectionDoesNotSave(t *testing.T) {
	service, repository, client := setupService()
	if err := service.TestConnection(context.Background(), ConnectionInput{Name: "try", Region: "us-east-1"}); err != nil {
		t.Fatal(err)
	}
	if client.testCalls != 1 {
		t.Fatalf("expected one connectivity test; got %d", client.testCalls)
	}
	if len(repository.connections) != 1 {
		t.Fatalf("test unexpectedly saved connection: %d", len(repository.connections))
	}
	client.testErr = errors.New("offline")
	if err := service.TestConnection(context.Background(), ConnectionInput{Name: "try", Region: "us-east-1"}); err == nil {
		t.Fatal("expected DynamoDB error")
	}
}

func TestWriteRequiresCompleteMatchingPrimaryKey(t *testing.T) {
	service, _, client := setupService()
	key := domain.Item{"id": domain.StringValue("a"), "version": domain.NumberValue("2")}
	item := domain.Item{"id": domain.StringValue("a"), "version": domain.NumberValue("2"), "data": domain.StringValue("value")}
	if err := service.PutItem(context.Background(), "conn", "records", key, item); err != nil {
		t.Fatal(err)
	}
	if client.lastItem == nil {
		t.Fatal("put was not called")
	}
	mismatched := domain.Item{"id": domain.StringValue("b"), "version": domain.NumberValue("2")}
	if err := service.PutItem(context.Background(), "conn", "records", key, mismatched); err == nil {
		t.Fatal("expected mismatched item key to fail")
	}
	if err := service.DeleteItem(context.Background(), "conn", "records", domain.Item{"id": domain.StringValue("a")}); err == nil {
		t.Fatal("expected incomplete key to fail")
	}
	if err := service.DeleteItem(context.Background(), "conn", "records", key); err != nil {
		t.Fatal(err)
	}
	if !client.deleted || client.lastKey["version"].Kind == "" {
		t.Fatal("delete did not receive complete key")
	}
}

func TestListItemsBoundsAndPaginationToken(t *testing.T) {
	service, _, client := setupService()
	if _, err := service.ListItems(context.Background(), "conn", "records", 101, "", nil); err == nil {
		t.Fatal("expected oversized page to fail")
	}
	filter := &domain.Filter{Attribute: "status", Operator: "eq", Value: domain.StringValue("ready")}
	if _, err := service.ListItems(context.Background(), "conn", "records", 20, "", filter); err != nil {
		t.Fatal(err)
	}
	if client.lastFilter != filter {
		t.Fatal("filter was not passed to DynamoDB")
	}
	key := domain.Item{"id": domain.StringValue("next")}
	token, err := EncodeToken(key)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := service.ListItems(context.Background(), "conn", "records", 20, token, nil); err != nil {
		t.Fatal(err)
	}
	if _, err := service.ListItems(context.Background(), "conn", "records", 20, "not-a-token", nil); err == nil {
		t.Fatal("expected malformed token to fail")
	}
}
