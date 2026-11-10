package dynamo

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

func TestTypedAttributeRoundTrip(t *testing.T) {
	input := `{"id":{"S":"abc"},"count":{"N":"2"},"tags":{"L":[{"S":"x"},{"BOOL":true}]},"meta":{"M":{"active":{"BOOL":true}}},"raw":{"B":"AQI="}}`
	var item Item
	if err := json.Unmarshal([]byte(input), &item); err != nil {
		t.Fatal(err)
	}
	if _, ok := item["count"].(*types.AttributeValueMemberN); !ok {
		t.Fatalf("number type was not preserved: %T", item["count"])
	}
	encoded, err := json.Marshal(item)
	if err != nil {
		t.Fatal(err)
	}
	var roundTrip Item
	if err := json.Unmarshal(encoded, &roundTrip); err != nil {
		t.Fatal(err)
	}
	if _, ok := roundTrip["tags"].(*types.AttributeValueMemberL); !ok {
		t.Fatalf("list type was not preserved: %T", roundTrip["tags"])
	}
	if !strings.Contains(string(encoded), `"N":"2"`) {
		t.Fatalf("typed number was lost: %s", encoded)
	}
}

func TestInvalidTypedAttributesAreRejected(t *testing.T) {
	invalid := []string{
		`{"x":"plain"}`,
		`{"x":{"S":"a","N":"1"}}`,
		`{"x":{"N":"01"}}`,
		`{"x":{"B":"not base64!"}}`,
		`{"x":{"NULL":false}}`,
		`{"x":{"UNKNOWN":"value"}}`,
	}
	for _, input := range invalid {
		var item Item
		if err := json.Unmarshal([]byte(input), &item); err == nil {
			t.Errorf("expected invalid attribute to fail: %s", input)
		}
	}
}
