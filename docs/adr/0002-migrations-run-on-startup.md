# Migrations run on startup

Schema changes are goose migrations in `db/migrations/`, embedded in the binary, and the app applies the ones it's missing on startup, right after it connects and before it listens. A failed migration stops the app, so the new machine never passes its health check (`fly.toml`), the deploy fails, and the old machines keep serving. goose takes a Postgres advisory lock, so two machines starting at once don't both run them. We chose startup over Fly's `release_command` because every database the app meets then takes the same path: production, a local database loaded from a backup, and the E2E databases. A `release_command` would run only on deploy, so a local database restored from an older backup would be one manual step behind. Startup costs one query per start once the database is current.

Migration 1 is the baseline: the schema production had before migrations, in `db/migrations/baseline.sql`. It's a Go migration that skips when the `movie` table exists, so production records it as applied without running it and the first deploy needs no manual step.

`db/schema.sql` stays the file tests and E2E load. `make schema` builds it by migrating an empty Postgres 15 container and dumping it, and ends it with the migration history, so a database loaded from it counts as current and the app applies nothing. CI rebuilds it the same way and fails if it differs from the committed file. Production is no longer read to make it.

There are no down migrations. A bad migration is fixed with a new one.

## Considered Options

- **Fly `release_command`.** It runs once per deploy on a temporary machine, before any new machine starts, and keeps migrations out of the app's startup. But local and E2E databases would need a separate step, and with `min_machines_running = 0` startup is already when the app first touches the database.
- **Apply changes to production by hand, as before.** Nothing records what was run, and the change has to land before the code that needs it deploys.
- **A baseline that's an SQL migration.** goose can't skip an SQL migration, so production would need its version inserted by hand before the first deploy.
