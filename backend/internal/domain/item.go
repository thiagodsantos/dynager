package domain

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"regexp"
)

type AttributeKind string

const (
	AttributeString     AttributeKind = "S"
	AttributeNumber     AttributeKind = "N"
	AttributeBinary     AttributeKind = "B"
	AttributeBoolean    AttributeKind = "BOOL"
	AttributeNull       AttributeKind = "NULL"
	AttributeMap        AttributeKind = "M"
	AttributeList       AttributeKind = "L"
	AttributeStringSet  AttributeKind = "SS"
	AttributeNumberSet  AttributeKind = "NS"
	AttributeBinarySet  AttributeKind = "BS"
	FilterOperatorEqual               = "eq"
	AttributeTypeString               = "S"
	AttributeTypeNumber               = "N"
	AttributeTypeBinary               = "B"
)

// AttributeValue is the domain representation of a typed DynamoDB value.
// It deliberately contains no AWS SDK types so adapters can translate at their boundary.
type AttributeValue struct {
	Kind      AttributeKind
	String    string
	Number    string
	Binary    []byte
	Boolean   bool
	Map       map[string]AttributeValue
	List      []AttributeValue
	StringSet []string
	NumberSet []string
	BinarySet [][]byte
}

type Item map[string]AttributeValue

type Filter struct {
	Attribute string
	Operator  string
	Value     AttributeValue
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

var numberPattern = regexp.MustCompile(`^-?(?:0|[1-9][0-9]*)(?:\.[0-9]+)?(?:[eE][+-]?[0-9]+)?$`)

func StringValue(value string) AttributeValue {
	return AttributeValue{Kind: AttributeString, String: value}
}
func NumberValue(value string) AttributeValue {
	return AttributeValue{Kind: AttributeNumber, Number: value}
}
func BinaryValue(value []byte) AttributeValue {
	return AttributeValue{Kind: AttributeBinary, Binary: value}
}
func BooleanValue(value bool) AttributeValue {
	return AttributeValue{Kind: AttributeBoolean, Boolean: value}
}

func (i *Item) UnmarshalJSON(data []byte) error {
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil || raw == nil {
		return fmt.Errorf("item must be a JSON object")
	}
	item := make(Item, len(raw))
	for name, encoded := range raw {
		if name == "" || len(name) > 255 {
			return fmt.Errorf("invalid attribute name %q", name)
		}
		var value AttributeValue
		if err := json.Unmarshal(encoded, &value); err != nil {
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
		encoded, err := json.Marshal(value)
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

func (v *AttributeValue) UnmarshalJSON(data []byte) error {
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(data, &fields); err != nil || len(fields) != 1 {
		return fmt.Errorf("must contain exactly one DynamoDB type")
	}
	for kind, encoded := range fields {
		var value AttributeValue
		value.Kind = AttributeKind(kind)
		switch value.Kind {
		case AttributeString:
			if err := json.Unmarshal(encoded, &value.String); err != nil {
				return fmt.Errorf("S must be a string")
			}
		case AttributeNumber:
			if err := json.Unmarshal(encoded, &value.Number); err != nil || !numberPattern.MatchString(value.Number) {
				return fmt.Errorf("N must be a valid DynamoDB number string")
			}
		case AttributeBinary:
			var text string
			if err := json.Unmarshal(encoded, &text); err != nil {
				return fmt.Errorf("B must be a base64 string")
			}
			binary, err := base64.StdEncoding.DecodeString(text)
			if err != nil {
				return fmt.Errorf("B must be a valid base64 string")
			}
			value.Binary = binary
		case AttributeBoolean:
			if err := json.Unmarshal(encoded, &value.Boolean); err != nil {
				return fmt.Errorf("BOOL must be a boolean")
			}
		case AttributeNull:
			if err := json.Unmarshal(encoded, &value.Boolean); err != nil || !value.Boolean {
				return fmt.Errorf("NULL must be true")
			}
		case AttributeMap:
			var raw map[string]json.RawMessage
			if err := json.Unmarshal(encoded, &raw); err != nil || raw == nil {
				return fmt.Errorf("M must be an object")
			}
			value.Map = make(map[string]AttributeValue, len(raw))
			for name, memberJSON := range raw {
				if name == "" || len(name) > 255 {
					return fmt.Errorf("M contains an invalid attribute name")
				}
				var member AttributeValue
				if err := json.Unmarshal(memberJSON, &member); err != nil {
					return fmt.Errorf("attribute %q: %w", name, err)
				}
				value.Map[name] = member
			}
		case AttributeList:
			var raw []json.RawMessage
			if err := json.Unmarshal(encoded, &raw); err != nil || raw == nil {
				return fmt.Errorf("L must be an array")
			}
			value.List = make([]AttributeValue, len(raw))
			for index, memberJSON := range raw {
				if err := json.Unmarshal(memberJSON, &value.List[index]); err != nil {
					return fmt.Errorf("list element %d: %w", index, err)
				}
			}
		case AttributeStringSet, AttributeNumberSet, AttributeBinarySet:
			var raw []string
			if err := json.Unmarshal(encoded, &raw); err != nil || len(raw) == 0 {
				return fmt.Errorf("%s must be a non-empty string array", kind)
			}
			switch value.Kind {
			case AttributeStringSet:
				value.StringSet = raw
			case AttributeNumberSet:
				for _, number := range raw {
					if !numberPattern.MatchString(number) {
						return fmt.Errorf("NS must contain valid DynamoDB number strings")
					}
				}
				value.NumberSet = raw
			case AttributeBinarySet:
				value.BinarySet = make([][]byte, len(raw))
				for index, encodedBinary := range raw {
					binary, err := base64.StdEncoding.DecodeString(encodedBinary)
					if err != nil {
						return fmt.Errorf("BS must contain valid base64 strings")
					}
					value.BinarySet[index] = binary
				}
			}
		default:
			return fmt.Errorf("unsupported DynamoDB type %q", kind)
		}
		*v = value
		return nil
	}
	return fmt.Errorf("missing DynamoDB type")
}

func (v AttributeValue) MarshalJSON() ([]byte, error) {
	var raw any
	switch v.Kind {
	case AttributeString:
		raw = v.String
	case AttributeNumber:
		raw = v.Number
	case AttributeBinary:
		raw = base64.StdEncoding.EncodeToString(v.Binary)
	case AttributeBoolean, AttributeNull:
		raw = v.Boolean
	case AttributeMap:
		raw = v.Map
	case AttributeList:
		raw = v.List
	case AttributeStringSet:
		raw = v.StringSet
	case AttributeNumberSet:
		raw = v.NumberSet
	case AttributeBinarySet:
		encoded := make([]string, len(v.BinarySet))
		for index, binary := range v.BinarySet {
			encoded[index] = base64.StdEncoding.EncodeToString(binary)
		}
		raw = encoded
	default:
		return nil, fmt.Errorf("unsupported domain attribute kind %q", v.Kind)
	}
	encoded, err := json.Marshal(raw)
	if err != nil {
		return nil, err
	}
	return json.Marshal(map[string]json.RawMessage{string(v.Kind): encoded})
}
