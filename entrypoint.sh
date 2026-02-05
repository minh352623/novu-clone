#!/bin/sh
set -e

# Check env vars before running migration
if [ -n "$PG_USERNAME" ] && [ -n "$PG_PASSWORD" ] && [ -n "$PG_HOST" ] && [ -n "$PG_PORT" ] && [ -n "$PG_DATABASE" ]; then
  echo "Running database migrations..."
  GOOSE_DRIVER=postgres \
  GOOSE_DBSTRING=postgres://$PG_USERNAME:$PG_PASSWORD@$PG_HOST:$PG_PORT/$PG_DATABASE \
  goose -dir sql/schema up
else
  echo "Skipping migrations (missing PG_* environment variables)"
fi

# Use a specific config file based on environment variable or default to environment.yml
CONFIG_FILE=${CONFIG_FILE:-/config/environment.yml}

echo "Starting tek-notification with config: $CONFIG_FILE"
/tek-notification "$CONFIG_FILE"
