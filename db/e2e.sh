#!/bin/sh
# Builds movies_e2e from the schema and the E2E seed, creating it if missing.
# postgres_test runs this when it starts, with the defaults below, and
# make e2e-db runs it from the checkout with PSQL, SCHEMA and SEED set.
#
# The database is emptied in place rather than dropped: Playwright starts the
# E2E server before its global setup rebuilds the database, and dropping it
# would cut off the server's connections.
set -e

PSQL=${PSQL:-psql -U $POSTGRES_USER}
SCHEMA=${SCHEMA:-/docker-entrypoint-initdb.d/schema.sql}
SEED=${SEED:-/db/seed.sql}

echo "SELECT 'CREATE DATABASE movies_e2e' WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = 'movies_e2e') \gexec" |
	$PSQL -v ON_ERROR_STOP=1 -q -d postgres

{
	echo 'SET client_min_messages = warning; DROP SCHEMA public CASCADE; CREATE SCHEMA public;'
	cat "$SCHEMA" "$SEED"
} | $PSQL -v ON_ERROR_STOP=1 -q -d movies_e2e >/dev/null
