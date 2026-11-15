#!/bin/sh
# Builds a database from the schema, and the seed when SEED is set, creating it
# if missing. make test-db and make e2e-db run it from the checkout with PSQL,
# SCHEMA, DB and SEED set.
#
# The database is emptied in place rather than dropped: Playwright starts the
# E2E server before its global setup rebuilds the database, and dropping it
# would cut off the server's connections.
set -e

: "${PSQL:?}" "${SCHEMA:?}" "${DB:?}"

echo "SELECT 'CREATE DATABASE $DB' WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = '$DB') \gexec" |
	$PSQL -v ON_ERROR_STOP=1 -q -d postgres

{
	echo 'SET client_min_messages = warning; DROP SCHEMA public CASCADE; CREATE SCHEMA public;'
	cat "$SCHEMA" ${SEED:+"$SEED"}
} | $PSQL -v ON_ERROR_STOP=1 -q -d "$DB" >/dev/null
