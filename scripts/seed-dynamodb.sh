#!/bin/sh
set -eu

: "${DYNAMODB_ENDPOINT:=http://dynamodb-local:8000}"

ensure_table() {
  table_name=$1
  shift

  if aws --endpoint-url "$DYNAMODB_ENDPOINT" dynamodb describe-table \
    --table-name "$table_name" >/dev/null 2>&1; then
    printf 'Keeping existing table and data: %s\n' "$table_name"
    return 1
  fi

  aws --endpoint-url "$DYNAMODB_ENDPOINT" dynamodb create-table \
    --table-name "$table_name" \
    --attribute-definitions "$@"
  aws --endpoint-url "$DYNAMODB_ENDPOINT" dynamodb wait table-exists \
    --table-name "$table_name"
  return 0
}

if ensure_table DemoProducts \
  AttributeName=id,AttributeType=S \
  --key-schema AttributeName=id,KeyType=HASH \
  --billing-mode PAY_PER_REQUEST; then
  aws --endpoint-url "$DYNAMODB_ENDPOINT" dynamodb batch-write-item \
    --request-items file:///fixtures/products.json
  printf 'Seeded DemoProducts\n'
fi

if ensure_table DemoOrders \
  AttributeName=customerId,AttributeType=S \
  AttributeName=orderId,AttributeType=S \
  --key-schema AttributeName=customerId,KeyType=HASH AttributeName=orderId,KeyType=RANGE \
  --billing-mode PAY_PER_REQUEST; then
  aws --endpoint-url "$DYNAMODB_ENDPOINT" dynamodb batch-write-item \
    --request-items file:///fixtures/orders.json
  printf 'Seeded DemoOrders\n'
fi
