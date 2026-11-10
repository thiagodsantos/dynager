package dynamodb

import (
	"context"
	"fmt"
	"net"
	"net/url"
	"os"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/thiagosantos/dynamodb-manager/backend/internal/domain"
)

type testItem struct {
	id      string
	version string
	status  string
}

func TestDynamoDBLocalTableItemFlow(t *testing.T) {
	endpoint := os.Getenv(EnvLocalEndpoint)
	if endpoint == "" {
		t.Skipf("set %s to run DynamoDB Local integration tests", EnvLocalEndpoint)
	}
	parsed, err := url.Parse(endpoint)
	if err != nil || parsed.Hostname() == "" || (parsed.Hostname() != "localhost" && (net.ParseIP(parsed.Hostname()) == nil || !net.ParseIP(parsed.Hostname()).IsLoopback())) {
		t.Fatalf("%s must point to localhost", EnvLocalEndpoint)
	}
	t.Setenv(EnvAWSAccessKeyID, "local")
	t.Setenv(EnvAWSSecretAccessKey, "local")
	t.Setenv(EnvAWSEC2MetadataDisabled, "true")
	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()
	cfg, err := awsconfig.LoadDefaultConfig(ctx, awsconfig.WithRegion("us-east-1"))
	if err != nil {
		t.Fatal("load local AWS configuration")
	}
	sdkClient := dynamodb.NewFromConfig(cfg, func(options *dynamodb.Options) { options.BaseEndpoint = aws.String(endpoint) })
	api := NewAWSClient()
	connection := domain.ConnectionSettings{Region: "us-east-1", Endpoint: endpoint}
	table := fmt.Sprintf("manager-it-%d", time.Now().UnixNano())
	_, err = sdkClient.CreateTable(ctx, &dynamodb.CreateTableInput{
		TableName: aws.String(table),
		AttributeDefinitions: []types.AttributeDefinition{
			{AttributeName: aws.String("id"), AttributeType: types.ScalarAttributeTypeS},
			{AttributeName: aws.String("version"), AttributeType: types.ScalarAttributeTypeN},
		},
		KeySchema: []types.KeySchemaElement{
			{AttributeName: aws.String("id"), KeyType: types.KeyTypeHash},
			{AttributeName: aws.String("version"), KeyType: types.KeyTypeRange},
		},
		BillingMode: types.BillingModePayPerRequest,
	})
	if err != nil {
		t.Fatal("create test table")
	}
	defer func() {
		_, _ = sdkClient.DeleteTable(context.Background(), &dynamodb.DeleteTableInput{TableName: aws.String(table)})
	}()
	if err := dynamodb.NewTableExistsWaiter(sdkClient).Wait(ctx, &dynamodb.DescribeTableInput{TableName: aws.String(table)}, 20*time.Second); err != nil {
		t.Fatal("wait for test table")
	}
	if err := api.Test(ctx, connection); err != nil {
		t.Fatalf("test connection: %v", err)
	}
	tables, err := api.ListTables(ctx, connection)
	if err != nil {
		t.Fatalf("list tables: %v", err)
	}
	found := false
	for _, name := range tables {
		if name == table {
			found = true
		}
	}
	if !found {
		t.Fatalf("created table %q missing from list", table)
	}
	description, err := api.DescribeTable(ctx, connection, table)
	if err != nil || description.PartitionKey.Name != "id" || description.PartitionKey.Type != "S" || description.SortKey == nil || description.SortKey.Name != "version" || description.SortKey.Type != "N" {
		t.Fatalf("description=%#v err=%v", description, err)
	}
	for _, value := range []testItem{{"1", "1", "keep"}, {"2", "2", "drop"}} {
		item := domain.Item{"id": domain.StringValue(value.id), "version": domain.NumberValue(value.version), "status": domain.StringValue(value.status), "count": domain.NumberValue("2")}
		if err := api.PutItem(ctx, connection, table, item); err != nil {
			t.Fatalf("put item %s: %v", value.id, err)
		}
	}
	page, err := api.Scan(ctx, connection, table, 1, nil, nil)
	if err != nil || len(page.Items) != 1 || len(page.NextKey) == 0 {
		t.Fatalf("first scan page=%#v err=%v", page, err)
	}
	filtered, err := api.Scan(ctx, connection, table, 10, nil, &domain.Filter{Attribute: "status", Operator: "eq", Value: domain.StringValue("keep")})
	if err != nil || len(filtered.Items) != 1 {
		t.Fatalf("filtered scan page=%#v err=%v", filtered, err)
	}
	key := domain.Item{"id": domain.StringValue("1"), "version": domain.NumberValue("1")}
	if err := api.DeleteItem(ctx, connection, table, key); err != nil {
		t.Fatalf("delete item: %v", err)
	}
	remaining, err := api.Scan(ctx, connection, table, 10, nil, nil)
	if err != nil || len(remaining.Items) != 1 {
		t.Fatalf("remaining scan page=%#v err=%v", remaining, err)
	}
}
