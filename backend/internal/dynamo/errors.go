package dynamo

import "errors"

var (
	errItemMustBeJSONObject           = errors.New("item must be a JSON object")
	errExactlyOneDynamoDBTypeRequired = errors.New("must contain exactly one DynamoDB type")
	errStringTypeRequired             = errors.New("S must be a string")
	errNumberTypeRequired             = errors.New("N must be a valid DynamoDB number string")
	errBinaryBase64Required           = errors.New("B must be a base64 string")
	errValidBinaryBase64Required      = errors.New("B must be a valid base64 string")
	errBooleanTypeRequired            = errors.New("BOOL must be a boolean")
	errNullMustBeTrue                 = errors.New("NULL must be true")
	errMapTypeRequired                = errors.New("M must be an object")
	errMapInvalidAttributeName        = errors.New("M contains an invalid attribute name")
	errListTypeRequired               = errors.New("L must be an array")
	errNonEmptyStringArrayRequired    = errors.New("must be a non-empty string array")
	errValidNumberSetRequired         = errors.New("NS must contain valid DynamoDB number strings")
	errValidBinarySetRequired         = errors.New("BS must contain valid base64 strings")
	errMissingDynamoDBType            = errors.New("missing DynamoDB type")
	errMissingTableDescription        = errors.New("DynamoDB returned no table description")
	errInvalidTableKeySchema          = errors.New("DynamoDB returned an invalid table key schema")
)
