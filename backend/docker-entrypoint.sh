#!/bin/bash
set -euo pipefail

echo "=== BPCL Portal API ==="
echo "Waiting for PostgreSQL to be ready..."

# Wait for database to be available
until pg_isready -h "${DB_HOST:-postgres}" -p "${DB_PORT:-5432}" -U "${DB_USER:-bpcl}" 2>/dev/null; do
  echo "PostgreSQL is unavailable - sleeping"
  sleep 2
done

echo "PostgreSQL is ready - running migrations..."

# Build the database URL
DATABASE_URL="postgresql://${DB_USER:-bpcl}:${DB_PASSWORD:-bpcl_secret}@${DB_HOST:-postgres}:${DB_PORT:-5432}/${DB_NAME:-bpcl_portal}?sslmode=disable"

# Run migrations
migrate -path /app/migrations -database "$DATABASE_URL" up

echo "Migrations complete - starting API server..."

# Execute the CMD
exec "$@"
