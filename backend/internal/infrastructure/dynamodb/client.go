package dynamodb

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/expression"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/thiagosantos/dynamodb-manager/backend/internal/domain"
)

type AWSClient struct{}

func NewAWSClient() *AWSClient { return &AWSClient{} }

func (a *AWSClient) client(ctx context.Context, connection domain.ConnectionSettings) (*dynamodb.Client, error) {
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

func (a *AWSClient) Test(ctx context.Context, connection domain.ConnectionSettings) error {
	client, err := a.client(ctx, connection)
	if err != nil {
		return err
	}
	_, err = client.ListTables(ctx, &dynamodb.ListTablesInput{Limit: aws.Int32(1)})
	return err
}

func (a *AWSClient) ListTables(ctx context.Context, connection domain.ConnectionSettings) ([]string, error) {
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

func (a *AWSClient) DescribeTable(ctx context.Context, connection domain.ConnectionSettings, table string) (domain.TableDescription, error) {
	client, err := a.client(ctx, connection)
	if err != nil {
		return domain.TableDescription{}, err
	}
	output, err := client.DescribeTable(ctx, &dynamodb.DescribeTableInput{TableName: aws.String(table)})
	if err != nil {
		return domain.TableDescription{}, err
	}
	if output.Table == nil {
		return domain.TableDescription{}, errMissingTableDescription
	}
	attributeTypes := make(map[string]string, len(output.Table.AttributeDefinitions))
	for _, definition := range output.Table.AttributeDefinitions {
		attributeTypes[aws.ToString(definition.AttributeName)] = string(definition.AttributeType)
	}
	var description domain.TableDescription
	for _, key := range output.Table.KeySchema {
		keyAttribute := domain.KeyAttribute{Name: aws.ToString(key.AttributeName), Type: attributeTypes[aws.ToString(key.AttributeName)]}
		switch key.KeyType {
		case types.KeyTypeHash:
			description.PartitionKey = keyAttribute
		case types.KeyTypeRange:
			description.SortKey = &keyAttribute
		}
	}
	if description.PartitionKey.Name == "" || description.PartitionKey.Type == "" {
		return domain.TableDescription{}, errInvalidTableKeySchema
	}
	return description, nil
}

func (a *AWSClient) Scan(ctx context.Context, connection domain.ConnectionSettings, table string, limit int32, start domain.Item, filter *domain.Filter) (domain.Page, error) {
	client, err := a.client(ctx, connection)
	if err != nil {
		return domain.Page{}, err
	}
	startKey, err := toAWSItem(start)
	if err != nil {
		return domain.Page{}, err
	}
	input := &dynamodb.ScanInput{TableName: aws.String(table), Limit: aws.Int32(limit), ExclusiveStartKey: startKey}
	if filter != nil {
		value, err := toAWSAttributeValue(filter.Value)
		if err != nil {
			return domain.Page{}, err
		}
		builder := expression.NewBuilder().WithFilter(expression.Name(filter.Attribute).Equal(expression.Value(value)))
		expr, err := builder.Build()
		if err != nil {
			return domain.Page{}, err
		}
		input.FilterExpression = expr.Filter()
		input.ExpressionAttributeNames = expr.Names()
		input.ExpressionAttributeValues = expr.Values()
	}
	output, err := client.Scan(ctx, input)
	if err != nil {
		return domain.Page{}, err
	}
	items := make([]domain.Item, len(output.Items))
	for index, item := range output.Items {
		items[index], err = fromAWSItem(item)
		if err != nil {
			return domain.Page{}, err
		}
	}
	nextKey, err := fromAWSItem(output.LastEvaluatedKey)
	if err != nil {
		return domain.Page{}, err
	}
	return domain.Page{Items: items, NextKey: nextKey, Count: output.Count, ScannedCount: output.ScannedCount}, nil
}

func (a *AWSClient) PutItem(ctx context.Context, connection domain.ConnectionSettings, table string, item domain.Item) error {
	client, err := a.client(ctx, connection)
	if err != nil {
		return err
	}
	awsItem, err := toAWSItem(item)
	if err != nil {
		return err
	}
	_, err = client.PutItem(ctx, &dynamodb.PutItemInput{TableName: aws.String(table), Item: awsItem})
	return err
}

func (a *AWSClient) DeleteItem(ctx context.Context, connection domain.ConnectionSettings, table string, key domain.Item) error {
	client, err := a.client(ctx, connection)
	if err != nil {
		return err
	}
	awsKey, err := toAWSItem(key)
	if err != nil {
		return err
	}
	_, err = client.DeleteItem(ctx, &dynamodb.DeleteItemInput{TableName: aws.String(table), Key: awsKey})
	return err
}

func toAWSItem(item domain.Item) (map[string]types.AttributeValue, error) {
	if item == nil {
		return nil, nil
	}
	result := make(map[string]types.AttributeValue, len(item))
	for name, value := range item {
		converted, err := toAWSAttributeValue(value)
		if err != nil {
			return nil, fmt.Errorf("attribute %q: %w", name, err)
		}
		result[name] = converted
	}
	return result, nil
}

func toAWSAttributeValue(value domain.AttributeValue) (types.AttributeValue, error) {
	switch value.Kind {
	case domain.AttributeString:
		return &types.AttributeValueMemberS{Value: value.String}, nil
	case domain.AttributeNumber:
		return &types.AttributeValueMemberN{Value: value.Number}, nil
	case domain.AttributeBinary:
		return &types.AttributeValueMemberB{Value: value.Binary}, nil
	case domain.AttributeBoolean:
		return &types.AttributeValueMemberBOOL{Value: value.Boolean}, nil
	case domain.AttributeNull:
		return &types.AttributeValueMemberNULL{Value: true}, nil
	case domain.AttributeMap:
		members, err := toAWSItem(domain.Item(value.Map))
		if err != nil {
			return nil, err
		}
		return &types.AttributeValueMemberM{Value: members}, nil
	case domain.AttributeList:
		members := make([]types.AttributeValue, len(value.List))
		for index, member := range value.List {
			converted, err := toAWSAttributeValue(member)
			if err != nil {
				return nil, fmt.Errorf("list element %d: %w", index, err)
			}
			members[index] = converted
		}
		return &types.AttributeValueMemberL{Value: members}, nil
	case domain.AttributeStringSet:
		return &types.AttributeValueMemberSS{Value: value.StringSet}, nil
	case domain.AttributeNumberSet:
		return &types.AttributeValueMemberNS{Value: value.NumberSet}, nil
	case domain.AttributeBinarySet:
		return &types.AttributeValueMemberBS{Value: value.BinarySet}, nil
	default:
		return nil, fmt.Errorf("unsupported domain attribute kind %q", value.Kind)
	}
}

func fromAWSItem(item map[string]types.AttributeValue) (domain.Item, error) {
	if item == nil {
		return nil, nil
	}
	result := make(domain.Item, len(item))
	for name, value := range item {
		converted, err := fromAWSAttributeValue(value)
		if err != nil {
			return nil, fmt.Errorf("attribute %q: %w", name, err)
		}
		result[name] = converted
	}
	return result, nil
}

func fromAWSAttributeValue(value types.AttributeValue) (domain.AttributeValue, error) {
	switch v := value.(type) {
	case *types.AttributeValueMemberS:
		return domain.StringValue(v.Value), nil
	case *types.AttributeValueMemberN:
		return domain.NumberValue(v.Value), nil
	case *types.AttributeValueMemberB:
		return domain.BinaryValue(v.Value), nil
	case *types.AttributeValueMemberBOOL:
		return domain.BooleanValue(v.Value), nil
	case *types.AttributeValueMemberNULL:
		return domain.AttributeValue{Kind: domain.AttributeNull, Boolean: v.Value}, nil
	case *types.AttributeValueMemberM:
		members, err := fromAWSItem(v.Value)
		return domain.AttributeValue{Kind: domain.AttributeMap, Map: map[string]domain.AttributeValue(members)}, err
	case *types.AttributeValueMemberL:
		members := make([]domain.AttributeValue, len(v.Value))
		for index, member := range v.Value {
			converted, err := fromAWSAttributeValue(member)
			if err != nil {
				return domain.AttributeValue{}, fmt.Errorf("list element %d: %w", index, err)
			}
			members[index] = converted
		}
		return domain.AttributeValue{Kind: domain.AttributeList, List: members}, nil
	case *types.AttributeValueMemberSS:
		return domain.AttributeValue{Kind: domain.AttributeStringSet, StringSet: v.Value}, nil
	case *types.AttributeValueMemberNS:
		return domain.AttributeValue{Kind: domain.AttributeNumberSet, NumberSet: v.Value}, nil
	case *types.AttributeValueMemberBS:
		return domain.AttributeValue{Kind: domain.AttributeBinarySet, BinarySet: v.Value}, nil
	default:
		return domain.AttributeValue{}, fmt.Errorf("unsupported AWS AttributeValue implementation %T", value)
	}
}
