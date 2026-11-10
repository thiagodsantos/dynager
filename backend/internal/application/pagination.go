package application

import (
	"encoding/base64"
	"encoding/json"

	"github.com/thiagosantos/dynamodb-manager/backend/internal/domain"
)

func decodeToken(token string) (domain.Item, error) {
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
	var key domain.Item
	if err := json.Unmarshal(encoded, &key); err != nil || len(key) == 0 {
		return nil, invalid("Pagination token is invalid")
	}
	return key, nil
}

func EncodeToken(key domain.Item) (string, error) {
	if len(key) == 0 {
		return "", nil
	}
	encoded, err := json.Marshal(key)
	if err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(encoded), nil
}
