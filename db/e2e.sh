#!/bin/sh
# Creates movies_e2e beside movies_test when postgres_test starts, loaded with
# the schema and the E2E seed. make e2e-db rebuilds it the same way.
set -e

psql -v ON_ERROR_STOP=1 -q -U "$POSTGRES_USER" -d postgres -c 'CREATE DATABASE movies_e2e'
psql -v ON_ERROR_STOP=1 -q -U "$POSTGRES_USER" -d movies_e2e \
	-f /docker-entrypoint-initdb.d/schema.sql \
	-f /db/seed.sql
