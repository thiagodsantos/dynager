package dynamo

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/expression"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

type Connection struct {
	Region   string
	Profile  string
	Endpoint string
}

type AWSClient struct{}

func NewAWSClient() *AWSClient { return &AWSClient{} }

func (a *AWSClient) client(ctx context.Context, connection Connection) (*dynamodb.Client, error) {
	options := []func(*awsconfig.LoadOptions) error{awsconfig.WithRegion(connection.Region)}
	if connection.Profile != "" {
		options = append(options, awsconfig.WithSharedConfigProfile(connection.Profile))
	}
	cfg, err := awsconfig.LoadDefaultConfig(ctx, options...)
	if err != nil {
		return nil, err
	}
	return dynamodb.NewFromConfig(cfg, func(options *dynamodb.Options) {
		if connection.Endpoint != "" {
			options.BaseEndpoint = aws.String(connection.Endpoint)
		}
	}), nil
}

func (a *AWSClient) Test(ctx context.Context, connection Connection) error {
	client, err := a.client(ctx, connection)
	if err != nil {
		return err
	}
	_, err = client.ListTables(ctx, &dynamodb.ListTablesInput{Limit: aws.Int32(1)})
	return err
}

func (a *AWSClient) ListTables(ctx context.Context, connection Connection) ([]string, error) {
	client, err := a.client(ctx, connection)
	if err != nil {
		return nil, err
	}
	var names []string
	paginator := dynamodb.NewListTablesPaginator(client, &dynamodb.ListTablesInput{})
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, err
		}
		names = append(names, page.TableNames...)
	}
	if names == nil {
		names = []string{}
	}
	return names, nil
}

func (a *AWSClient) DescribeTable(ctx context.Context, connection Connection, table string) (TableDescription, error) {
	client, err := a.client(ctx, connection)
	if err != nil {
		return TableDescription{}, err
	}
	output, err := client.DescribeTable(ctx, &dynamodb.DescribeTableInput{TableName: aws.String(table)})
	if err != nil {
		return TableDescription{}, err
	}
	if output.Table == nil {
		return TableDescription{}, errMissingTableDescription
	}
	attributeTypes := make(map[string]string, len(output.Table.AttributeDefinitions))
	for _, definition := range output.Table.AttributeDefinitions {
		attributeTypes[aws.ToString(definition.AttributeName)] = string(definition.AttributeType)
	}
	var description TableDescription
	for _, key := range output.Table.KeySchema {
		keyAttribute := KeyAttribute{Name: aws.ToString(key.AttributeName), Type: attributeTypes[aws.ToString(key.AttributeName)]}
		switch key.KeyType {
		case types.KeyTypeHash:
			description.PartitionKey = keyAttribute
		case types.KeyTypeRange:
			description.SortKey = &keyAttribute
		}
	}
	if description.PartitionKey.Name == "" || description.PartitionKey.Type == "" {
		return TableDescription{}, errInvalidTableKeySchema
	}
	return description, nil
}

func (a *AWSClient) Scan(ctx context.Context, connection Connection, table string, limit int32, start Item, filter *Filter) (Page, error) {
	client, err := a.client(ctx, connection)
	if err != nil {
		return Page{}, err
	}
	input := &dynamodb.ScanInput{TableName: aws.String(table), Limit: aws.Int32(limit), ExclusiveStartKey: map[string]types.AttributeValue(start)}
	if filter != nil {
		builder := expression.NewBuilder().WithFilter(expression.Name(filter.Attribute).Equal(expression.Value(filter.Value)))
		expr, err := builder.Build()
		if err != nil {
			return Page{}, err
		}
		input.FilterExpression = expr.Filter()
		input.ExpressionAttributeNames = expr.Names()
		input.ExpressionAttributeValues = expr.Values()
	}
	output, err := client.Scan(ctx, input)
	if err != nil {
		return Page{}, err
	}
	items := make([]Item, len(output.Items))
	for index, item := range output.Items {
		items[index] = Item(item)
	}
	return Page{Items: items, NextKey: Item(output.LastEvaluatedKey), Count: output.Count, ScannedCount: output.ScannedCount}, nil
}

func (a *AWSClient) PutItem(ctx context.Context, connection Connection, table string, item Item) error {
	client, err := a.client(ctx, connection)
	if err != nil {
		return err
	}
	_, err = client.PutItem(ctx, &dynamodb.PutItemInput{TableName: aws.String(table), Item: map[string]types.AttributeValue(item)})
	return err
}

func (a *AWSClient) DeleteItem(ctx context.Context, connection Connection, table string, key Item) error {
	client, err := a.client(ctx, connection)
	if err != nil {
		return err
	}
	_, err = client.DeleteItem(ctx, &dynamodb.DeleteItemInput{TableName: aws.String(table), Key: map[string]types.AttributeValue(key)})
	return err
}
