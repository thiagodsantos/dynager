package dynamo

const (
	EnvLocalEndpoint          = "DYNAMODB_LOCAL_ENDPOINT"
	EnvAWSAccessKeyID         = "AWS_ACCESS_KEY_ID"
	EnvAWSSecretAccessKey     = "AWS_SECRET_ACCESS_KEY"
	EnvAWSEC2MetadataDisabled = "AWS_EC2_METADATA_DISABLED"

	FilterOperatorEqual = "eq"
	AttributeTypeString = "S"
	AttributeTypeNumber = "N"
	AttributeTypeBinary = "B"
)
