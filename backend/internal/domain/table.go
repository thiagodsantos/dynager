package domain

import (
	"fmt"
	"reflect"
	"regexp"
)

var tablePattern = regexp.MustCompile(`^[a-zA-Z0-9_.-]{3,255}$`)

func ValidateTableName(table string) error {
	if !tablePattern.MatchString(table) {
		return InvalidInput("Table name is invalid")
	}
	return nil
}

func ValidatePrimaryKey(key, item Item, description TableDescription) error {
	keys := []KeyAttribute{description.PartitionKey}
	if description.SortKey != nil {
		keys = append(keys, *description.SortKey)
	}
	if len(key) != len(keys) {
		return InvalidInput("Key must contain exactly the table's complete primary key")
	}
	for _, attribute := range keys {
		value, exists := key[attribute.Name]
		if !exists || value.Kind == "" {
			return InvalidInput(fmt.Sprintf("Key is missing primary key attribute %q", attribute.Name))
		}
		if err := matchesSchemaType(value, attribute.Type); err != nil {
			return InvalidInput(fmt.Sprintf("Key attribute %q %s", attribute.Name, err.Error()))
		}
		if item != nil {
			itemValue, exists := item[attribute.Name]
			if !exists || !sameValue(value, itemValue) {
				return InvalidInput(fmt.Sprintf("Item primary key attribute %q must match key", attribute.Name))
			}
		}
	}
	return nil
}

func matchesSchemaType(value AttributeValue, attributeType string) error {
	var actual string
	switch value.Kind {
	case AttributeString:
		actual = AttributeTypeString
	case AttributeNumber:
		actual = AttributeTypeNumber
	case AttributeBinary:
		actual = AttributeTypeBinary
	default:
		return errKeyValueType
	}
	if actual != attributeType {
		return fmt.Errorf("must use type %s", attributeType)
	}
	return nil
}

func sameValue(left, right AttributeValue) bool {
	return left.Kind != "" && reflect.DeepEqual(left, right)
}
