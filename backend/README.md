# DynamoDB Manager backend

The API is a trusted local single-user service. It only listens on a loopback address and uses the AWS SDK for Go v2 default credential chain (including an optional shared-config profile). AWS credentials are never accepted by the API, persisted, returned, or logged.

## Architecture

The backend follows a dependency-inward, hexagonal layout:

- `internal/domain` contains connection rules, domain errors, DynamoDB item/value types, and key validation. It has no dependency on the AWS SDK or HTTP framework.
- `internal/application` contains use cases and the outbound `Repository` and `Dynamo` ports. It depends on the domain and does not know about storage or transport implementations.
- `internal/infrastructure` contains adapters: `httpapi` is the inbound HTTP adapter, `config` implements the connection repository and runtime configuration, and `dynamodb` implements the DynamoDB port and maps domain values to the AWS SDK.
- `cmd/server` wires the adapters to the application service.

Dependencies point inward: infrastructure depends on application/domain contracts, while the domain and application layers do not import infrastructure packages.

## Run

From `backend/`:

```sh
go run ./cmd/server
```

The default listener is `127.0.0.1:8080`. Configuration is stored as non-secret connection metadata in `${user-config-dir}/dynamodb-manager/connections.json` with private file permissions. Override the path with `DYNAMODB_MANAGER_CONFIG`.

Environment options:

- `DYNAMODB_MANAGER_ADDR` — listener address; must be `localhost` or a loopback IP (default `127.0.0.1:8080`).
- `DYNAMODB_MANAGER_CONFIG` — connection metadata file path.
- `DYNAMODB_MANAGER_CORS_ORIGINS` — comma-separated exact browser origins (default `http://localhost:5173,http://127.0.0.1:5173`). Wildcard origins are ignored.

All `/api` routes require `X-API-Version: v1`. `/health` does not. The frontend can configure a region, optional AWS shared profile, and optional HTTP(S) DynamoDB endpoint override; endpoint URLs containing credentials, paths, queries, or fragments are rejected.

## Verify

```sh
go test ./...
go vet ./...
```

The fast test suite does not need AWS access or Docker. To run the DynamoDB Local integration test, start DynamoDB Local separately and set its loopback URL:

```sh
DYNAMODB_LOCAL_ENDPOINT=http://localhost:8000 go test ./internal/infrastructure/dynamodb -run TestDynamoDBLocalTableItemFlow -count=1
```
