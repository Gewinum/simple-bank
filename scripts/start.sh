#!/bin/sh

set -e

echo "running the migration"
/migrate -path /migration -database "$DB_SOURCE" -verbose up

echo "start the app"
exec "$@"