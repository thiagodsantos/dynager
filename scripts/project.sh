#!/bin/sh
set -eu

GO=${GO:-go}
GOFMT=${GOFMT:-gofmt}
PNPM=${PNPM:-pnpm}
DYNAMODB_LOCAL_PORT=${DYNAMODB_LOCAL_PORT:-8000}
DYNAMODB_DEV_PORT=${DYNAMODB_DEV_PORT:-8001}
DYNAMODB_ENDPOINT=${DYNAMODB_ENDPOINT:-http://127.0.0.1:${DYNAMODB_LOCAL_PORT}}
BACKEND_RUN_PACKAGE=${BACKEND_RUN_PACKAGE:-./cmd/server}
DOCKER_COMPOSE=${DOCKER_COMPOSE:-docker compose}

compose_dev() {
  $DOCKER_COMPOSE -p dynamodb-manager-dev -f docker/dev.yaml "$@"
}

compose_test() {
  $DOCKER_COMPOSE -p dynamodb-manager-test -f docker/test.yaml "$@"
}

frontend_has_script() {
  [ -f frontend/package.json ] && node -e \
    "const p = require('./frontend/package.json'); process.exit(p.scripts && p.scripts[process.argv[1]] ? 0 : 1)" "$1"
}

run_frontend_script() {
  script=$1
  if [ ! -f frontend/package.json ]; then
    printf 'Skipping frontend: frontend/package.json is not present yet.\n'
  elif frontend_has_script "$script"; then
    $PNPM --dir frontend run "$script"
  else
    printf "Skipping frontend: script '%s' is not defined.\n" "$script"
  fi
}

has_go_module() {
  [ -f backend/go.mod ]
}

has_go_source() {
  find backend -type f -name '*.go' -print -quit | grep -q .
}

run_go_module_command() {
  message=$1
  shift
  if ! has_go_module; then
    printf 'Skipping %s: backend/go.mod is not present yet.\n' "$message"
  elif ! has_go_source; then
    printf 'Skipping %s: no Go source files are present yet.\n' "$message"
  else
    (cd backend && "$GO" "$@")
  fi
}

run_help() {
  cat <<'EOF'
DynamoDB Manager project commands:
  help              Show this command list
  install           Install frontend dependencies and download Go modules
  run               Start the frontend dev script, if defined
  run-backend       Start backend/cmd/server (override BACKEND_RUN_PACKAGE if needed)
  build             Build frontend and Go packages
  test              Run unit and Docker-backed integration tests
  test-unit         Run frontend and Go unit tests
  test-integration  Run the DynamoDB Local integration test in Go
  lint              Run frontend ESLint/TypeScript checks and Go vet/format checks
  lint-fix          Apply frontend ESLint autofixes
  format            Run the frontend format script and gofmt
  clean             Run app clean scripts and go clean
  dev-up            Start dev DynamoDB Local and seed demo data
  dev-down          Stop dev containers and preserve their data
  dev-recreate      Recreate dev containers and seed missing tables
  test-up           Start test DynamoDB Local
  test-down         Stop and remove test containers
  test-recreate     Recreate test DynamoDB Local

Frontend scripts are discovered from frontend/package.json; absent scripts are skipped.
EOF
}

run_install() {
  found=0
  if [ -f frontend/package.json ]; then
    if [ -f frontend/pnpm-lock.yaml ]; then
      $PNPM --dir frontend install --frozen-lockfile
    else
      $PNPM --dir frontend install
    fi
    found=1
  fi
  if has_go_module; then
    (cd backend && "$GO" mod download)
    found=1
  fi
  if [ "$found" -eq 0 ]; then
    printf 'Nothing to install yet (frontend/package.json and backend/go.mod are absent).\n'
  fi
}

run_backend() {
  if ! has_go_module; then
    printf 'Skipping backend: backend/go.mod is not present yet.\n'
  elif ! has_go_source; then
    printf 'Skipping backend: no Go source files are present yet.\n'
  else
    (cd backend && "$GO" run "$BACKEND_RUN_PACKAGE")
  fi
}

run_build() {
  run_frontend_script build
  run_go_module_command 'Go build' build ./...
}

run_test_unit() {
  if [ -f frontend/package.json ]; then
    if frontend_has_script test:unit; then
      $PNPM --dir frontend run test:unit
    elif frontend_has_script test; then
      $PNPM --dir frontend run test
    else
      printf 'Skipping frontend tests: neither test:unit nor test is defined.\n'
    fi
  else
    printf 'Skipping frontend tests: frontend/package.json is not present yet.\n'
  fi
  run_go_module_command 'Go unit tests' test ./...
}

run_dev_up() {
  DYNAMODB_DEV_PORT="$DYNAMODB_DEV_PORT" compose_dev up -d --wait dynamodb-local
  DYNAMODB_DEV_PORT="$DYNAMODB_DEV_PORT" compose_dev run --rm seed
}

run_dev_down() {
  DYNAMODB_DEV_PORT="$DYNAMODB_DEV_PORT" compose_dev stop
}

run_dev_recreate() {
  DYNAMODB_DEV_PORT="$DYNAMODB_DEV_PORT" compose_dev down --remove-orphans
  run_dev_up
}

run_test_up() {
  DYNAMODB_LOCAL_PORT="$DYNAMODB_LOCAL_PORT" compose_test up -d --wait dynamodb-local
}

run_test_down() {
  DYNAMODB_LOCAL_PORT="$DYNAMODB_LOCAL_PORT" compose_test down --remove-orphans
}

run_test_recreate() {
  run_test_down
  run_test_up
}

run_test_integration() {
  if ! has_go_module; then
    printf 'Skipping integration tests: backend/go.mod is not present yet.\n'
    return
  fi
  if ! has_go_source; then
    printf 'Skipping integration tests: no Go source files are present yet.\n'
    return
  fi

  cleanup_compose() {
    DYNAMODB_LOCAL_PORT="$DYNAMODB_LOCAL_PORT" compose_test down --remove-orphans >/dev/null 2>&1 || true
  }
  trap cleanup_compose EXIT
  trap 'exit 130' INT
  trap 'exit 143' TERM

  run_test_up
  (
    cd backend
    DYNAMODB_LOCAL_ENDPOINT="$DYNAMODB_ENDPOINT" AWS_REGION=us-east-1 \
      "$GO" test ./internal/dynamo -run '^TestDynamoDBLocalTableItemFlow$' -count=1
  )
}

run_lint() {
  if [ -f frontend/package.json ]; then
    if frontend_has_script lint; then
      $PNPM --dir frontend run lint
    elif frontend_has_script check; then
      $PNPM --dir frontend run check
    else
      printf 'Skipping frontend lint: neither lint nor check is defined.\n'
    fi
  else
    printf 'Skipping frontend lint: frontend/package.json is not present yet.\n'
  fi

  if has_go_module && has_go_source; then
    (cd backend && "$GO" vet ./...)
    unformatted=$(find backend -type f -name '*.go' -exec "$GOFMT" -l {} +)
    if [ -n "$unformatted" ]; then
      printf 'Run make format; Go files need formatting:\n%s\n' "$unformatted"
      return 1
    fi
  elif ! has_go_module; then
    printf 'Skipping Go lint: backend/go.mod is not present yet.\n'
  else
    printf 'Skipping Go lint: no Go source files are present yet.\n'
  fi
}

run_lint_fix() {
  run_frontend_script lint:fix
}

run_format() {
  run_frontend_script format
  if has_go_module && has_go_source; then
    find backend -type f -name '*.go' -exec "$GOFMT" -w {} +
  elif ! has_go_module; then
    printf 'Skipping Go format: backend/go.mod is not present yet.\n'
  else
    printf 'Skipping Go format: no Go source files are present yet.\n'
  fi
}

run_clean() {
  run_frontend_script clean
  if has_go_module; then
    (cd backend && "$GO" clean)
  else
    printf 'Skipping Go clean: backend/go.mod is not present yet.\n'
  fi
}

command=${1:-help}
case "$command" in
  help) run_help ;;
  install) run_install ;;
  run) run_frontend_script dev ;;
  run-backend) run_backend ;;
  build) run_build ;;
  test-unit) run_test_unit ;;
  test-integration) run_test_integration ;;
  lint) run_lint ;;
  lint-fix) run_lint_fix ;;
  format) run_format ;;
  clean) run_clean ;;
  dev-up) run_dev_up ;;
  dev-down) run_dev_down ;;
  dev-recreate) run_dev_recreate ;;
  test-up) run_test_up ;;
  test-down) run_test_down ;;
  test-recreate) run_test_recreate ;;
  *) printf 'Unknown project command: %s\n' "$command" >&2; exit 2 ;;
esac
