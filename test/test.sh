#!/bin/bash
set -e

COMPOSE_FILE="test/compose.yaml"

echo "Starting test database..."
docker compose -f $COMPOSE_FILE up -d test-db

echo "Waiting for PostgreSQL to be ready..."
until docker compose -f $COMPOSE_FILE exec -t test-db pg_isready -U tuser -d fingo_test; do
  sleep 1
done

echo "Running migrations..."
goose -dir sql/schema postgres "postgres://tuser:test@localhost:5433/fingo_test?sslmode=disable" up

echo "Running Go tests..."
go test -v ./...

docker compose -f $COMPOSE_FILE down
