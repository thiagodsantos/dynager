package usecase

import (
	"encoding/json"
	"fmt"
	"net/url"
	"regexp"
	"strings"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/thiagosantos/dynamodb-manager/backend/internal/dynamo"
)

var (
	regionPattern = regexp.MustCompile(`^[a-z0-9][a-z0-9-]{1,62}[a-z0-9]$`)
	tablePattern  = regexp.MustCompile(`^[a-zA-Z0-9_.-]{3,255}$`)
)

func ValidateConnectionInput(input ConnectionInput) (ConnectionInput, error) {
	input.Name = strings.TrimSpace(input.Name)
	input.Region = strings.TrimSpace(input.Region)
	input.Profile = strings.TrimSpace(input.Profile)
	input.Endpoint = strings.TrimSpace(input.Endpoint)
	if input.Name == "" || len(input.Name) > MaxConnectionNameLength || hasControl(input.Name) {
		return ConnectionInput{}, invalid("Name must be between 1 and 100 characters")
	}
	if !regionPattern.MatchString(input.Region) || !strings.Contains(input.Region, "-") {
		return ConnectionInput{}, invalid("Region is invalid")
	}
	if len(input.Profile) > MaxProfileNameLength || hasControl(input.Profile) {
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
		actual = dynamo.AttributeTypeString
	case *types.AttributeValueMemberN:
		actual = dynamo.AttributeTypeNumber
	case *types.AttributeValueMemberB:
		actual = dynamo.AttributeTypeBinary
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

func hasControl(value string) bool {
	for _, char := range value {
		if char < 0x20 || char == 0x7f {
			return true
		}
	}
	return false
}
