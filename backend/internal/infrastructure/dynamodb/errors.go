package dynamodb

import "errors"

var (
	errMissingTableDescription = errors.New("DynamoDB returned no table description")
	errInvalidTableKeySchema   = errors.New("DynamoDB returned an invalid table key schema")
)
