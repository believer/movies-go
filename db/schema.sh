#!/bin/sh
# Prints the schema an empty database ends up with after every migration, for
# db/schema.sql. It migrates a scratch Postgres 15 container, to match
# production, and dumps it with that container's pg_dump.
#
# The dump ends with the migration history, so a database loaded from it
# counts as fully migrated and the app doesn't run the migrations again.
set -e

container=$(docker run -d --rm \
	-e POSTGRES_USER=admin -e POSTGRES_PASSWORD=password -e POSTGRES_DB=movies \
	-p 127.0.0.1::5432 postgres:15)
trap 'docker stop "$container" >/dev/null' EXIT

# Over TCP: during init the server only listens on the socket
until docker exec "$container" pg_isready -q -h 127.0.0.1 -U admin -d movies; do
	sleep 0.2
done

port=$(docker port "$container" 5432 | head -n 1 | sed 's/.*://')

DATABASE_URL="postgres://admin:password@localhost:$port/movies?sslmode=disable" \
	go run ./cmd/migrate >&2

# The version lines and the restrict lines, which carry a random token, change
# with every dump, so they are dropped to keep diffs clean.
docker exec "$container" pg_dump -U admin -d movies \
	--schema-only --no-owner --no-privileges |
	sed -e '/^-- Dumped from /d' -e '/^-- Dumped by /d' \
		-e '/^\\restrict /d' -e '/^\\unrestrict /d'

docker exec "$container" psql -U admin -d movies -At -c "
	SELECT 'INSERT INTO public.goose_db_version (version_id, is_applied) VALUES'
		|| E'\n' || string_agg(format('    (%s, %s)', version_id, is_applied::text), E',\n' ORDER BY id)
		|| E';\n'
	FROM goose_db_version"
