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

seed_records() {
  seed_table=$1
  seed_key_type=$2
  seed_batch="{\"${seed_table}\":["
  seed_batch_count=0
  seed_index=1

  while [ "$seed_index" -le 26 ]; do
    seed_number=$(printf '%03d' "$seed_index")
    if [ "$seed_key_type" = composite ]; then
      seed_item="\"customerId\":{\"S\":\"demo-customer-${seed_number}\"},\"orderId\":{\"S\":\"demo-order-${seed_number}\"}"
    else
      seed_item="\"id\":{\"S\":\"demo-${seed_number}\"}"
    fi
    seed_request="{\"PutRequest\":{\"Item\":{${seed_item},\"name\":{\"S\":\"${seed_table} record ${seed_number}\"},\"status\":{\"S\":\"active\"},\"sequence\":{\"N\":\"${seed_index}\"}}}}"

    if [ "$seed_batch_count" -gt 0 ]; then
      seed_batch="${seed_batch},"
    fi
    seed_batch="${seed_batch}${seed_request}"
    seed_batch_count=$((seed_batch_count + 1))

    if [ "$seed_batch_count" -eq 25 ]; then
      aws --endpoint-url "$DYNAMODB_ENDPOINT" dynamodb batch-write-item \
        --request-items "${seed_batch}]}"
      seed_batch="{\"${seed_table}\":["
      seed_batch_count=0
    fi
    seed_index=$((seed_index + 1))
  done

  if [ "$seed_batch_count" -gt 0 ]; then
    aws --endpoint-url "$DYNAMODB_ENDPOINT" dynamodb batch-write-item \
      --request-items "${seed_batch}]}"
  fi
}

ensure_seed_records() {
  seed_table=$1
  seed_key_type=$2
  seed_count=$(aws --endpoint-url "$DYNAMODB_ENDPOINT" dynamodb scan \
    --table-name "$seed_table" --select COUNT --query Count --output text)

  if [ "$seed_count" -le 25 ]; then
    seed_records "$seed_table" "$seed_key_type"
    printf 'Added pagination records to %s\n' "$seed_table"
  fi
}

if ensure_table DemoProducts \
  AttributeName=id,AttributeType=S \
  --key-schema AttributeName=id,KeyType=HASH \
  --billing-mode PAY_PER_REQUEST; then
  aws --endpoint-url "$DYNAMODB_ENDPOINT" dynamodb batch-write-item \
    --request-items file:///fixtures/products.json
  printf 'Seeded DemoProducts\n'
fi
ensure_seed_records DemoProducts hash

if ensure_table DemoOrders \
  AttributeName=customerId,AttributeType=S \
  AttributeName=orderId,AttributeType=S \
  --key-schema AttributeName=customerId,KeyType=HASH AttributeName=orderId,KeyType=RANGE \
  --billing-mode PAY_PER_REQUEST; then
  aws --endpoint-url "$DYNAMODB_ENDPOINT" dynamodb batch-write-item \
    --request-items file:///fixtures/orders.json
  printf 'Seeded DemoOrders\n'
fi
ensure_seed_records DemoOrders composite

for table_name in \
  DemoCustomers \
  DemoInventory \
  DemoEmployees \
  DemoTickets \
  DemoSessions \
  DemoEvents \
  DemoReviews \
  DemoPayments \
  DemoShipments \
  DemoNotifications \
  DemoDevices \
  DemoArticles \
  DemoComments \
  DemoSubscriptions \
  DemoMetrics \
  DemoTasks \
  DemoProjects \
  DemoMessages \
  DemoLocations \
  DemoInvoices; do
  if ensure_table "$table_name" \
    AttributeName=id,AttributeType=S \
    --key-schema AttributeName=id,KeyType=HASH \
    --billing-mode PAY_PER_REQUEST; then
    printf 'Seeded %s\n' "$table_name"
  fi
  ensure_seed_records "$table_name" hash
done
