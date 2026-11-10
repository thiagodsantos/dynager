package app

import (
	"encoding/base64"
	"encoding/json"

	"github.com/thiagosantos/dynamodb-manager/backend/internal/dynamo"
)

func decodeToken(token string) (dynamo.Item, error) {
	if token == "" {
		return nil, nil
	}
	if len(token) > MaxPaginationTokenSize {
		return nil, invalid("Pagination token is too large")
	}
	encoded, err := base64.RawURLEncoding.DecodeString(token)
	if err != nil {
		return nil, invalid("Pagination token is invalid")
	}
	var key dynamo.Item
	if err := json.Unmarshal(encoded, &key); err != nil || len(key) == 0 {
		return nil, invalid("Pagination token is invalid")
	}
	return key, nil
}

func EncodeToken(key dynamo.Item) (string, error) {
	if len(key) == 0 {
		return "", nil
	}
	encoded, err := json.Marshal(key)
	if err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(encoded), nil
}
