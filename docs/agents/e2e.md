# E2E tests

Playwright specs live in `tests/`. The comments above `e2e` and `e2e-db` in the `Makefile` cover ports, `PSQL` and running beside another checkout. This file holds what the config doesn't tell you.

## Running

- Run Chromium only: `pnpm exec playwright test --project=chromium`. Firefox can hang in agent sandboxes, and CI runs all three browsers on every PR.
- Each server port gets its own database, `movies_e2e_<port>`. `make e2e` builds it before Playwright starts the server. Before running `pnpm exec playwright test` on a new port, run `make e2e-db` with the same `E2E_PORT`.
- Other checkouts' agents run E2E too. Pick an `E2E_PORT` that no other run is using; the TMDB stub takes the port after it. `lsof -nP -iTCP -sTCP:LISTEN | grep e2e-serve` lists the ports in use.
- `http://localhost:<port>/health is already used` means a server is already running on that port. If `lsof -a -p <pid> -d cwd` shows it running from this checkout, it's left over from an earlier run and may run old code: stop it and run again. If it runs from another checkout, it belongs to another agent's run: pick other ports.

## Writing

- Expected values come from `db/seed.sql`. Add rows there when a test needs data the seed lacks. `tests/seed.ts` names the seed movies that write tests set up data for.
- Specs on a port share one database, so a spec shows what a page displays, not that a row was created. Prove inserts in a Go test.
- Tests that write log in as a fresh user through the `user` fixture in `tests/fixtures.ts`.
- TMDB responses come from `tests/tmdb/fixtures/`, keyed by request path. A new Movie intake test needs a movie of its own there.
- A language page's heading is the native name: the link on a movie page reads "French", the heading "Français".
- `getByLabel("<Detail>")` returns a metadata row's whole value. Narrow it with `.getByRole("link")` when the row also holds buttons, like My rating for a logged-in user.
- A case an open bug breaks is a `test.fixme` whose title names the issue, e.g. `"a re-rated movie shows the latest Rating (#73)"`. The PR that fixes the bug turns it into a `test`.

## Visual baselines

`tests/visual.spec.ts` screenshots the movie metadata and stats sections. The committed baselines are Linux Chromium from CI, so the spec skips locally unless `VISUAL=1`.

- **Check a markup or CSS change for visual change:** run the spec with `VISUAL=1` and `--update-snapshots` on `origin/main`'s code (`git checkout origin/main -- components views styles`, then restore) to write local baselines that git ignores, then with `VISUAL=1` alone on your change. A diff fails the test.
- **Change a baseline on purpose:** delete its PNG in `tests/visual.spec.ts-snapshots/` and push. The Chromium E2E job fails and writes a new one. Commit it from the job's `e2e-report-chromium` artifact.
