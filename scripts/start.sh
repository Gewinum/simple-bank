#!/bin/sh

set -e

echo "running the migration"
source /app.env
/migrate -path /migration -database "$DB_SOURCE" -verbose up

echo "start the app"
exec "$@"