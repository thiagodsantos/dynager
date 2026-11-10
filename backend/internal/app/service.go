package app

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/url"
	"regexp"
	"strings"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
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

var regionPattern = regexp.MustCompile(`^[a-z0-9][a-z0-9-]{1,62}[a-z0-9]$`)
var tablePattern = regexp.MustCompile(`^[a-zA-Z0-9_.-]{3,255}$`)

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
	if id == "" {
		return invalid("Connection ID is required")
	}
	return s.repository.Delete(ctx, id)
}

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
	if limit < 1 || limit > 100 {
		return dynamo.Page{}, invalid("Limit must be between 1 and 100")
	}
	if filter != nil {
		if filter.Attribute == "" || len(filter.Attribute) > 255 {
			return dynamo.Page{}, invalid("Filter attribute must be between 1 and 255 characters")
		}
		if filter.Operator != "eq" {
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

func (s *Service) getConnection(ctx context.Context, id string) (Connection, error) {
	if id == "" {
		return Connection{}, invalid("Connection ID is required")
	}
	connection, err := s.repository.Get(ctx, id)
	if err != nil {
		return Connection{}, err
	}
	return connection, nil
}

func ValidateConnectionInput(input ConnectionInput) (ConnectionInput, error) {
	input.Name = strings.TrimSpace(input.Name)
	input.Region = strings.TrimSpace(input.Region)
	input.Profile = strings.TrimSpace(input.Profile)
	input.Endpoint = strings.TrimSpace(input.Endpoint)
	if input.Name == "" || len(input.Name) > 100 || hasControl(input.Name) {
		return ConnectionInput{}, invalid("Name must be between 1 and 100 characters")
	}
	if !regionPattern.MatchString(input.Region) || !strings.Contains(input.Region, "-") {
		return ConnectionInput{}, invalid("Region is invalid")
	}
	if len(input.Profile) > 128 || hasControl(input.Profile) {
		return ConnectionInput{}, invalid("Profile must be at most 128 characters")
	}
	if input.Endpoint != "" {
		parsed, err := url.ParseRequestURI(input.Endpoint)
		if err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") || parsed.Host == "" || parsed.User != nil || parsed.RawQuery != "" || parsed.Fragment != "" || (parsed.Path != "" && parsed.Path != "/") {
			return ConnectionInput{}, invalid("Endpoint must be an absolute HTTP(S) URL without credentials, path, query, or fragment")
		}
	}
	return input, nil
}

func validateTable(table string) error {
	if !tablePattern.MatchString(table) {
		return invalid("Table name is invalid")
	}
	return nil
}

func validateKey(key, item dynamo.Item, description dynamo.TableDescription) error {
	keys := []dynamo.KeyAttribute{description.PartitionKey}
	if description.SortKey != nil {
		keys = append(keys, *description.SortKey)
	}
	if len(key) != len(keys) {
		return invalid("Key must contain exactly the table's complete primary key")
	}
	for _, attribute := range keys {
		value, exists := key[attribute.Name]
		if !exists || value == nil {
			return invalid(fmt.Sprintf("Key is missing primary key attribute %q", attribute.Name))
		}
		if err := matchesSchemaType(value, attribute.Type); err != nil {
			return invalid(fmt.Sprintf("Key attribute %q %s", attribute.Name, err.Error()))
		}
		if item != nil {
			itemValue, exists := item[attribute.Name]
			if !exists || !sameValue(value, itemValue) {
				return invalid(fmt.Sprintf("Item primary key attribute %q must match key", attribute.Name))
			}
		}
	}
	return nil
}

func matchesSchemaType(value types.AttributeValue, attributeType string) error {
	var actual string
	switch value.(type) {
	case *types.AttributeValueMemberS:
		actual = "S"
	case *types.AttributeValueMemberN:
		actual = "N"
	case *types.AttributeValueMemberB:
		actual = "B"
	default:
		return fmt.Errorf("must be a string, number, or binary value")
	}
	if actual != attributeType {
		return fmt.Errorf("must use type %s", attributeType)
	}
	return nil
}

func sameValue(left, right types.AttributeValue) bool {
	if left == nil || right == nil {
		return false
	}
	a, errA := json.Marshal(dynamo.Item{"value": left})
	b, errB := json.Marshal(dynamo.Item{"value": right})
	return errA == nil && errB == nil && string(a) == string(b)
}

func decodeToken(token string) (dynamo.Item, error) {
	if token == "" {
		return nil, nil
	}
	if len(token) > 16*1024 {
		return nil, invalid("Pagination token is too large")
	}
	encoded, err := base64.RawURLEncoding.DecodeString(token)
	if err != nil {
		return nil, invalid("Pagination token is invalid")
	}
	var key dynamo.Item
	if err := json.Unmarshal(encoded, &key); err != nil || len(key) == 0 {
		return nil, invalid("Pagination token is invalid")
	}
	return key, nil
}

func EncodeToken(key dynamo.Item) (string, error) {
	if len(key) == 0 {
		return "", nil
	}
	encoded, err := json.Marshal(key)
	if err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(encoded), nil
}

func toDynamoConnection(connection Connection) dynamo.Connection {
	return dynamo.Connection{Region: connection.Region, Profile: connection.Profile, Endpoint: connection.Endpoint}
}

func hasControl(value string) bool {
	for _, char := range value {
		if char < 0x20 || char == 0x7f {
			return true
		}
	}
	return false
}

func newID() (string, error) {
	var bytes [16]byte
	if _, err := rand.Read(bytes[:]); err != nil {
		return "", err
	}
	bytes[6] = bytes[6]&0x0f | 0x40
	bytes[8] = bytes[8]&0x3f | 0x80
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x", bytes[0:4], bytes[4:6], bytes[6:8], bytes[8:10], bytes[10:16]), nil
}
