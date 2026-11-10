package app

import (
	"context"
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/thiagosantos/dynamodb-manager/backend/internal/dynamo"
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
	lastFilter  *dynamo.Filter
	lastKey     dynamo.Item
	lastItem    dynamo.Item
	deleted     bool
	description dynamo.TableDescription
}

func (d *fakeDynamo) Test(context.Context, dynamo.Connection) error { d.testCalls++; return d.testErr }
func (*fakeDynamo) ListTables(context.Context, dynamo.Connection) ([]string, error) {
	return []string{"records"}, nil
}
func (d *fakeDynamo) DescribeTable(context.Context, dynamo.Connection, string) (dynamo.TableDescription, error) {
	return d.description, nil
}
func (d *fakeDynamo) Scan(_ context.Context, _ dynamo.Connection, _ string, _ int32, _ dynamo.Item, filter *dynamo.Filter) (dynamo.Page, error) {
	d.lastFilter = filter
	return dynamo.Page{Items: []dynamo.Item{}, Count: 0, ScannedCount: 4}, nil
}
func (d *fakeDynamo) PutItem(_ context.Context, _ dynamo.Connection, _ string, item dynamo.Item) error {
	d.lastItem = item
	return nil
}
func (d *fakeDynamo) DeleteItem(_ context.Context, _ dynamo.Connection, _ string, key dynamo.Item) error {
	d.lastKey = key
	d.deleted = true
	return nil
}

func setupService() (*Service, *memoryRepository, *fakeDynamo) {
	repository := &memoryRepository{connections: map[string]Connection{"conn": {ID: "conn", Name: "local", Region: "us-east-1"}}}
	client := &fakeDynamo{description: dynamo.TableDescription{PartitionKey: dynamo.KeyAttribute{Name: "id", Type: "S"}, SortKey: &dynamo.KeyAttribute{Name: "version", Type: "N"}}}
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
	key := dynamo.Item{"id": &types.AttributeValueMemberS{Value: "a"}, "version": &types.AttributeValueMemberN{Value: "2"}}
	item := dynamo.Item{"id": &types.AttributeValueMemberS{Value: "a"}, "version": &types.AttributeValueMemberN{Value: "2"}, "data": &types.AttributeValueMemberS{Value: "value"}}
	if err := service.PutItem(context.Background(), "conn", "records", key, item); err != nil {
		t.Fatal(err)
	}
	if client.lastItem == nil {
		t.Fatal("put was not called")
	}
	mismatched := dynamo.Item{"id": &types.AttributeValueMemberS{Value: "b"}, "version": &types.AttributeValueMemberN{Value: "2"}}
	if err := service.PutItem(context.Background(), "conn", "records", key, mismatched); err == nil {
		t.Fatal("expected mismatched item key to fail")
	}
	if err := service.DeleteItem(context.Background(), "conn", "records", dynamo.Item{"id": &types.AttributeValueMemberS{Value: "a"}}); err == nil {
		t.Fatal("expected incomplete key to fail")
	}
	if err := service.DeleteItem(context.Background(), "conn", "records", key); err != nil {
		t.Fatal(err)
	}
	if !client.deleted || client.lastKey["version"] == nil {
		t.Fatal("delete did not receive complete key")
	}
}

func TestListItemsBoundsAndPaginationToken(t *testing.T) {
	service, _, client := setupService()
	if _, err := service.ListItems(context.Background(), "conn", "records", 101, "", nil); err == nil {
		t.Fatal("expected oversized page to fail")
	}
	filter := &dynamo.Filter{Attribute: "status", Operator: "eq", Value: &types.AttributeValueMemberS{Value: "ready"}}
	if _, err := service.ListItems(context.Background(), "conn", "records", 20, "", filter); err != nil {
		t.Fatal(err)
	}
	if client.lastFilter != filter {
		t.Fatal("filter was not passed to DynamoDB")
	}
	key := dynamo.Item{"id": &types.AttributeValueMemberS{Value: "next"}}
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
