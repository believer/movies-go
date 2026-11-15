# Schema changes

The schema changes only through a migration (`docs/adr/0002-migrations-run-on-startup.md`).

1. `make migration NAME=add_latest_rating` creates the next `db/migrations/NNNNN_add_latest_rating.sql`. Write the change under `-- +goose Up`. There are no down migrations: fix a bad one with a new one.
2. `make schema` rebuilds `db/schema.sql` from the migrations, in a scratch Postgres 15 container. Commit both files. CI fails if `db/schema.sql` doesn't match the migrations.
3. `make e2e` rebuilds its database from `db/schema.sql` on every run, but `postgres_test`, which `make test` uses, loads it only when the container starts. Restart it with `docker-compose up -d --force-recreate --wait postgres_test`.

The app applies pending migrations when it starts, so production gets them on deploy. Keep a migration safe to run while the previous release still serves traffic: add a column or view before the code that reads it, and drop one only after no deployed code uses it.

Never edit a migration that has been merged, or `db/migrations/baseline.sql`: production has already run them.
