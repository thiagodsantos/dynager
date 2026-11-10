package httpapi

import (
	"context"
	"encoding/json"
	"errors"

	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/rs/zerolog"
	"github.com/thiagosantos/dynamodb-manager/backend/internal/dynamo"
	"github.com/thiagosantos/dynamodb-manager/backend/internal/usecase"
)

type testRepository struct{ connections map[string]usecase.Connection }

func (r *testRepository) List(context.Context) ([]usecase.Connection, error) {
	result := []usecase.Connection{}
	for _, item := range r.connections {
		result = append(result, item)
	}
	return result, nil
}
func (r *testRepository) Get(_ context.Context, id string) (usecase.Connection, error) {
	item, ok := r.connections[id]
	if !ok {
		return usecase.Connection{}, &usecase.Error{Code: usecase.NotFound, Message: "Connection not found"}
	}
	return item, nil
}
func (r *testRepository) Create(_ context.Context, item usecase.Connection) (usecase.Connection, error) {
	r.connections[item.ID] = item
	return item, nil
}
func (r *testRepository) Delete(_ context.Context, id string) error {
	delete(r.connections, id)
	return nil
}

type testDynamo struct {
	tests   int
	filter  *dynamo.Filter
	testErr error
}

func (d *testDynamo) Test(context.Context, dynamo.Connection) error { d.tests++; return d.testErr }
func (*testDynamo) ListTables(context.Context, dynamo.Connection) ([]string, error) {
	return []string{"records"}, nil
}
func (*testDynamo) DescribeTable(context.Context, dynamo.Connection, string) (dynamo.TableDescription, error) {
	return dynamo.TableDescription{PartitionKey: dynamo.KeyAttribute{Name: "id", Type: "S"}}, nil
}
func (d *testDynamo) Scan(_ context.Context, _ dynamo.Connection, _ string, _ int32, _ dynamo.Item, filter *dynamo.Filter) (dynamo.Page, error) {
	d.filter = filter
	return dynamo.Page{Items: []dynamo.Item{}, Count: 0, ScannedCount: 2}, nil
}
func (*testDynamo) PutItem(context.Context, dynamo.Connection, string, dynamo.Item) error { return nil }
func (*testDynamo) DeleteItem(context.Context, dynamo.Connection, string, dynamo.Item) error {
	return nil
}

func newTestHandler() (http.Handler, *testDynamo) {
	repository := &testRepository{connections: map[string]usecase.Connection{"conn": {ID: "conn", Name: "test", Region: "us-east-1"}}}
	client := &testDynamo{}
	logger := zerolog.New(&strings.Builder{})
	return NewRouter(usecase.NewService(repository, client), []string{"http://localhost:5173"}, &logger), client
}

func TestHealthDoesNotRequireVersionAndAPIDoes(t *testing.T) {
	handler, _ := newTestHandler()
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/health", nil))
	if response.Code != http.StatusOK {
		t.Fatalf("health status=%d", response.Code)
	}
	response = httptest.NewRecorder()
	handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/api/connections", nil))
	if response.Code != http.StatusBadRequest || !strings.Contains(response.Body.String(), "unsupported_api_version") {
		t.Fatalf("missing-version response=%d %s", response.Code, response.Body.String())
	}
}

func TestCreateConnectionDoesNotReturnCredentialFieldsAndVersionIsAccepted(t *testing.T) {
	handler, _ := newTestHandler()
	request := httptest.NewRequest(http.MethodPost, "/api/connections", strings.NewReader(`{"name":"dev","region":"us-east-1","profile":"local"}`))
	request.Header.Set("X-API-Version", "v1")
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusCreated {
		t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
	}
	var result map[string]any
	if err := json.Unmarshal(response.Body.Bytes(), &result); err != nil {
		t.Fatal(err)
	}
	if _, exists := result["secret"]; exists {
		t.Fatal("response exposed a secret field")
	}
	if result["profile"] != "local" {
		t.Fatalf("profile missing: %v", result)
	}
}

func TestTestConnectionFailureIsSanitizedAndNotPersisted(t *testing.T) {
	repository := &testRepository{connections: map[string]usecase.Connection{}}
	client := &testDynamo{testErr: errors.New("secret credential value")}
	logger := zerolog.New(&strings.Builder{})
	handler := NewRouter(usecase.NewService(repository, client), nil, &logger)
	request := httptest.NewRequest(http.MethodPost, "/api/connections/test", strings.NewReader(`{"name":"try","region":"us-east-1"}`))
	request.Header.Set("X-API-Version", "v1")
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusOK || strings.Contains(response.Body.String(), "secret credential value") || !strings.Contains(response.Body.String(), `"success":false`) {
		t.Fatalf("unexpected test response: status=%d body=%s", response.Code, response.Body.String())
	}
	if len(repository.connections) != 0 {
		t.Fatal("connection test unexpectedly persisted a connection")
	}
}

func TestCreateRejectsCredentialFieldsWithoutEchoingThem(t *testing.T) {
	handler, _ := newTestHandler()
	request := httptest.NewRequest(http.MethodPost, "/api/connections", strings.NewReader(`{"name":"dev","region":"us-east-1","secretAccessKey":"do-not-echo"}`))
	request.Header.Set("X-API-Version", "v1")
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusBadRequest || strings.Contains(response.Body.String(), "do-not-echo") {
		t.Fatalf("credential field was not safely rejected: status=%d body=%s", response.Code, response.Body.String())
	}
}

func TestTableDescriptionUsesFrontendContractFieldNames(t *testing.T) {
	handler, _ := newTestHandler()
	request := httptest.NewRequest(http.MethodGet, "/api/connections/conn/tables/records", nil)
	request.Header.Set("X-API-Version", "v1")
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
	}
	var result map[string]any
	if err := json.Unmarshal(response.Body.Bytes(), &result); err != nil {
		t.Fatal(err)
	}
	partitionKey, ok := result["partitionKey"].(map[string]any)
	if !ok || partitionKey["name"] != "id" || partitionKey["type"] != "S" {
		t.Fatalf("unexpected partition key contract: %v", result["partitionKey"])
	}
}

func TestFilterUsesTypedValueAndRejectsUnsupportedOperator(t *testing.T) {
	handler, client := newTestHandler()
	request := httptest.NewRequest(http.MethodGet, `/api/connections/conn/tables/records/items?filter=%7B%22attribute%22%3A%22status%22%2C%22operator%22%3A%22eq%22%2C%22value%22%3A%7B%22S%22%3A%22ready%22%7D%7D`, nil)
	request.Header.Set("X-API-Version", "v1")
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
	}
	if _, ok := client.filter.Value.(*types.AttributeValueMemberS); !ok {
		t.Fatalf("filter value was not typed: %T", client.filter.Value)
	}
	request = httptest.NewRequest(http.MethodGet, `/api/connections/conn/tables/records/items?filter=%7B%22attribute%22%3A%22status%22%2C%22operator%22%3A%22contains%22%2C%22value%22%3A%7B%22S%22%3A%22ready%22%7D%7D`, nil)
	request.Header.Set("X-API-Version", "v1")
	response = httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusBadRequest {
		t.Fatalf("unsupported operator status=%d", response.Code)
	}
	request = httptest.NewRequest(http.MethodGet, "/api/connections/conn/tables/records/items?filter=%7B%", nil)
	request.Header.Set("X-API-Version", "v1")
	response = httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusBadRequest {
		t.Fatalf("malformed query status=%d", response.Code)
	}
}
