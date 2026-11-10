package dynamodb

import (
	"reflect"
	"testing"

	"github.com/thiagosantos/dynamodb-manager/backend/internal/domain"
)

func TestAWSAttributeValueMappingRoundTrip(t *testing.T) {
	item := domain.Item{
		"string":     domain.StringValue("text"),
		"number":     domain.NumberValue("12.5"),
		"binary":     domain.BinaryValue([]byte{1, 2, 3}),
		"boolean":    domain.BooleanValue(true),
		"null":       {Kind: domain.AttributeNull, Boolean: true},
		"map":        {Kind: domain.AttributeMap, Map: map[string]domain.AttributeValue{"nested": domain.StringValue("value")}},
		"list":       {Kind: domain.AttributeList, List: []domain.AttributeValue{domain.StringValue("item"), domain.NumberValue("2")}},
		"string-set": {Kind: domain.AttributeStringSet, StringSet: []string{"a", "b"}},
		"number-set": {Kind: domain.AttributeNumberSet, NumberSet: []string{"1", "2"}},
		"binary-set": {Kind: domain.AttributeBinarySet, BinarySet: [][]byte{{1}, {2, 3}}},
	}

	awsItem, err := toAWSItem(item)
	if err != nil {
		t.Fatal(err)
	}
	actual, err := fromAWSItem(awsItem)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(actual, item) {
		t.Fatalf("round-trip mismatch:\n got: %#v\nwant: %#v", actual, item)
	}
}
