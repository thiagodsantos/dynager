package dynamo

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

// Item preserves DynamoDB's typed AttributeValue JSON representation at the API boundary.
type Item map[string]types.AttributeValue

var numberPattern = regexp.MustCompile(`^-?(?:0|[1-9][0-9]*)(?:\.[0-9]+)?(?:[eE][+-]?[0-9]+)?$`)

func (i *Item) UnmarshalJSON(data []byte) error {
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil || raw == nil {
		return errors.New("item must be a JSON object")
	}
	item := make(Item, len(raw))
	for name, encoded := range raw {
		if name == "" || len(name) > 255 {
			return fmt.Errorf("invalid attribute name %q", name)
		}
		value, err := decodeAttribute(encoded)
		if err != nil {
			return fmt.Errorf("attribute %q: %w", name, err)
		}
		item[name] = value
	}
	*i = item
	return nil
}

func (i Item) MarshalJSON() ([]byte, error) {
	var buffer bytes.Buffer
	buffer.WriteByte('{')
	first := true
	for name, value := range i {
		encoded, err := encodeAttribute(value)
		if err != nil {
			return nil, fmt.Errorf("attribute %q: %w", name, err)
		}
		if !first {
			buffer.WriteByte(',')
		}
		first = false
		key, _ := json.Marshal(name)
		buffer.Write(key)
		buffer.WriteByte(':')
		buffer.Write(encoded)
	}
	buffer.WriteByte('}')
	return buffer.Bytes(), nil
}

func decodeAttribute(data []byte) (types.AttributeValue, error) {
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(data, &fields); err != nil || len(fields) != 1 {
		return nil, errors.New("must contain exactly one DynamoDB type")
	}
	for kind, value := range fields {
		switch kind {
		case "S":
			var v string
			if err := json.Unmarshal(value, &v); err != nil {
				return nil, errors.New("S must be a string")
			}
			return &types.AttributeValueMemberS{Value: v}, nil
		case "N":
			var v string
			if err := json.Unmarshal(value, &v); err != nil || !numberPattern.MatchString(v) {
				return nil, errors.New("N must be a valid DynamoDB number string")
			}

			return &types.AttributeValueMemberN{Value: v}, nil
		case "B":
			var v string
			if err := json.Unmarshal(value, &v); err != nil {
				return nil, errors.New("B must be a base64 string")
			}
			decoded, err := base64.StdEncoding.DecodeString(v)
			if err != nil {
				return nil, errors.New("B must be a valid base64 string")
			}
			return &types.AttributeValueMemberB{Value: decoded}, nil
		case "BOOL":
			var v bool
			if err := json.Unmarshal(value, &v); err != nil {
				return nil, errors.New("BOOL must be a boolean")
			}
			return &types.AttributeValueMemberBOOL{Value: v}, nil
		case "NULL":
			var v bool
			if err := json.Unmarshal(value, &v); err != nil || !v {
				return nil, errors.New("NULL must be true")
			}
			return &types.AttributeValueMemberNULL{Value: true}, nil
		case "M":
			var object map[string]json.RawMessage
			if err := json.Unmarshal(value, &object); err != nil || object == nil {
				return nil, errors.New("M must be an object")
			}
			members := make(map[string]types.AttributeValue, len(object))
			for name, raw := range object {
				if name == "" || len(name) > 255 {
					return nil, errors.New("M contains an invalid attribute name")
				}
				member, err := decodeAttribute(raw)
				if err != nil {
					return nil, fmt.Errorf("attribute %q: %w", name, err)
				}
				members[name] = member
			}
			return &types.AttributeValueMemberM{Value: members}, nil
		case "L":
			var values []json.RawMessage
			if err := json.Unmarshal(value, &values); err != nil || values == nil {
				return nil, errors.New("L must be an array")
			}
			members := make([]types.AttributeValue, len(values))
			for index, raw := range values {
				member, err := decodeAttribute(raw)
				if err != nil {
					return nil, fmt.Errorf("list element %d: %w", index, err)
				}
				members[index] = member
			}
			return &types.AttributeValueMemberL{Value: members}, nil
		case "SS", "NS", "BS":
			var values []string
			if err := json.Unmarshal(value, &values); err != nil || len(values) == 0 {
				return nil, fmt.Errorf("%s must be a non-empty string array", kind)
			}
			switch kind {
			case "SS":
				return &types.AttributeValueMemberSS{Value: values}, nil
			case "NS":
				for _, v := range values {
					if !numberPattern.MatchString(v) {
						return nil, errors.New("NS must contain valid DynamoDB number strings")
					}
				}
				return &types.AttributeValueMemberNS{Value: values}, nil
			default:
				decoded := make([][]byte, len(values))
				for index, v := range values {
					b, err := base64.StdEncoding.DecodeString(v)
					if err != nil {
						return nil, errors.New("BS must contain valid base64 strings")
					}
					decoded[index] = b
				}
				return &types.AttributeValueMemberBS{Value: decoded}, nil
			}
		default:
			return nil, fmt.Errorf("unsupported DynamoDB type %q", kind)
		}
	}
	return nil, errors.New("missing DynamoDB type")
}

func encodeAttribute(value types.AttributeValue) ([]byte, error) {
	var kind string
	var raw any
	switch v := value.(type) {
	case *types.AttributeValueMemberS:
		kind, raw = "S", v.Value
	case *types.AttributeValueMemberN:
		kind, raw = "N", v.Value
	case *types.AttributeValueMemberB:
		kind, raw = "B", base64.StdEncoding.EncodeToString(v.Value)
	case *types.AttributeValueMemberBOOL:
		kind, raw = "BOOL", v.Value
	case *types.AttributeValueMemberNULL:
		kind, raw = "NULL", v.Value
	case *types.AttributeValueMemberM:
		kind = "M"
		members := Item(v.Value)
		encoded, err := members.MarshalJSON()
		if err != nil {
			return nil, err
		}
		return json.Marshal(map[string]json.RawMessage{kind: encoded})
	case *types.AttributeValueMemberL:
		kind = "L"
		values := make([]json.RawMessage, len(v.Value))
		for index, member := range v.Value {
			encoded, err := encodeAttribute(member)
			if err != nil {
				return nil, err
			}
			values[index] = encoded
		}
		raw = values
	case *types.AttributeValueMemberSS:
		kind, raw = "SS", v.Value
	case *types.AttributeValueMemberNS:
		kind, raw = "NS", v.Value
	case *types.AttributeValueMemberBS:
		kind = "BS"
		values := make([]string, len(v.Value))
		for index, member := range v.Value {
			values[index] = base64.StdEncoding.EncodeToString(member)
		}
		raw = values
	default:
		return nil, fmt.Errorf("unsupported AttributeValue implementation %T", value)
	}
	encoded, err := json.Marshal(raw)
	if err != nil {
		return nil, err
	}
	return json.Marshal(map[string]json.RawMessage{kind: encoded})
}

type Filter struct {
	Attribute string
	Operator  string
	Value     types.AttributeValue
}

type KeyAttribute struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

type TableDescription struct {
	PartitionKey KeyAttribute
	SortKey      *KeyAttribute
}

type Page struct {
	Items        []Item
	NextKey      Item
	Count        int32
	ScannedCount int32
}
