# E2E tests

Playwright specs live in `tests/`. The comments above `e2e` and `e2e-db` in the `Makefile` cover ports, `PSQL` and running beside another checkout. This file holds what the config doesn't tell you.

## Running

- Run Chromium only: `pnpm exec playwright test --project=chromium`. Firefox can hang in agent sandboxes, and CI runs all three browsers on every PR.
- In a checkout where `movies_e2e` has never been built, run `make e2e-db` first. Playwright starts the server before its global setup rebuilds the database, so the first run can't connect without it.
- `http://localhost:8090/health is already used` means a server from an earlier run is still up and may run old code. Stop it with `lsof -ti tcp:8090,8091 | xargs kill`, then run again.

## Writing

- Expected values come from `db/seed.sql`. Add rows there when a test needs data the seed lacks. `tests/seed.ts` names the seed movies that write tests set up data for.
- Tests that write log in as a fresh user through the `user` fixture in `tests/fixtures.ts`.
- TMDB responses come from `tests/tmdb/fixtures/`, keyed by request path. A new Movie intake test needs a movie of its own there.
- A language page's heading is the native name: the link on a movie page reads "French", the heading "Français".
