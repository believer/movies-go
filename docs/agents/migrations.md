# Schema changes

The schema changes only through a migration (`docs/adr/0002-migrations-run-on-startup.md`).

1. `make migration NAME=add_latest_rating` creates the next `db/migrations/NNNNN_add_latest_rating.sql`. Write the change under `-- +goose Up`. There are no down migrations: fix a bad one with a new one.
2. `make schema` rebuilds `db/schema.sql` from the migrations, in a scratch Postgres 15 container. Commit both files. `make test` and `make e2e` load `db/schema.sql` from your checkout, so they pick up the change. CI fails if `db/schema.sql` doesn't match the migrations.

The app applies pending migrations when it starts, so production gets them on deploy. Keep a migration safe to run while the previous release still serves traffic: add a column or view before the code that reads it, and drop one only after no deployed code uses it.

Production's role has a 10 second `statement_timeout`, and migrations run as that role. A migration with a slower statement, like an index on a large table or a backfill, starts with `SET LOCAL statement_timeout = 0;` under `-- +goose Up`.

Never edit a migration that has been merged, or `db/migrations/baseline.sql`: production has already run them.
