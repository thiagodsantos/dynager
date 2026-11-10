package app

import "github.com/thiagosantos/dynamodb-manager/backend/internal/dynamo"

func toDynamoConnection(connection Connection) dynamo.Connection {
	return dynamo.Connection{Region: connection.Region, Profile: connection.Profile, Endpoint: connection.Endpoint}
}
