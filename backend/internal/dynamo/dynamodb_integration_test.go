package dynamo

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
)

func TestDynamoDBLocalTableItemFlow(t *testing.T) {
	endpoint := os.Getenv("DYNAMODB_LOCAL_ENDPOINT")
	if endpoint == "" {
		t.Skip("set DYNAMODB_LOCAL_ENDPOINT to run DynamoDB Local integration tests")
	}
	parsed, err := url.Parse(endpoint)
	if err != nil || parsed.Hostname() == "" || (parsed.Hostname() != "localhost" && (net.ParseIP(parsed.Hostname()) == nil || !net.ParseIP(parsed.Hostname()).IsLoopback())) {
		t.Fatal("DYNAMODB_LOCAL_ENDPOINT must point to localhost")
	}
	t.Setenv("AWS_ACCESS_KEY_ID", "local")
	t.Setenv("AWS_SECRET_ACCESS_KEY", "local")
	t.Setenv("AWS_EC2_METADATA_DISABLED", "true")
	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()
	cfg, err := awsconfig.LoadDefaultConfig(ctx, awsconfig.WithRegion("us-east-1"))
	if err != nil {
		t.Fatal("load local AWS configuration")
	}
	sdkClient := dynamodb.NewFromConfig(cfg, func(options *dynamodb.Options) { options.BaseEndpoint = aws.String(endpoint) })
	api := NewAWSClient()
	connection := Connection{Region: "us-east-1", Endpoint: endpoint}
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
	for _, value := range []struct{ id, version, status string }{{"1", "1", "keep"}, {"2", "2", "drop"}} {
		item := Item{"id": &types.AttributeValueMemberS{Value: value.id}, "version": &types.AttributeValueMemberN{Value: value.version}, "status": &types.AttributeValueMemberS{Value: value.status}, "count": &types.AttributeValueMemberN{Value: "2"}}
		if err := api.PutItem(ctx, connection, table, item); err != nil {
			t.Fatalf("put item %s: %v", value.id, err)
		}
	}
	page, err := api.Scan(ctx, connection, table, 1, nil, nil)
	if err != nil || len(page.Items) != 1 || len(page.NextKey) == 0 {
		t.Fatalf("first scan page=%#v err=%v", page, err)
	}
	filtered, err := api.Scan(ctx, connection, table, 10, nil, &Filter{Attribute: "status", Operator: "eq", Value: &types.AttributeValueMemberS{Value: "keep"}})
	if err != nil || len(filtered.Items) != 1 {
		t.Fatalf("filtered scan page=%#v err=%v", filtered, err)
	}
	key := Item{"id": &types.AttributeValueMemberS{Value: "1"}, "version": &types.AttributeValueMemberN{Value: "1"}}
	if err := api.DeleteItem(ctx, connection, table, key); err != nil {
		t.Fatalf("delete item: %v", err)
	}
	remaining, err := api.Scan(ctx, connection, table, 10, nil, nil)
	if err != nil || len(remaining.Items) != 1 {
		t.Fatalf("remaining scan page=%#v err=%v", remaining, err)
	}
}
