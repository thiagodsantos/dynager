# DynamoDB Manager backend

The API is a trusted local single-user service. It only listens on a loopback address and uses the AWS SDK for Go v2 default credential chain (including an optional shared-config profile). AWS credentials are never accepted by the API, persisted, returned, or logged.

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
gofmt -w cmd/server/*.go internal/*/*.go
go test ./...
go vet ./...
```

The fast test suite does not need AWS access or Docker. To run the DynamoDB Local integration test, start DynamoDB Local separately and set its loopback URL:

```sh
DYNAMODB_LOCAL_ENDPOINT=http://localhost:8000 go test ./internal/dynamo -run TestDynamoDBLocalTableItemFlow -count=1
```
