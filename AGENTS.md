Run `make verify E2E_PORT=<free port>` before finishing a change: `make check`, `make test` and the Chromium E2E suite.

**E2E**: running or writing Playwright tests → `docs/agents/e2e.md`.

**Schema changes**: adding a table, column, index or view → `docs/agents/migrations.md`.

Implement in a worktree off a fresh `origin/main`: `git fetch`, then `git worktree add ../movies-<slug> -b <type>/<slug> origin/main`, with `<type>` the commit type and a `<slug>` missing from `git worktree list`, then `pnpm install` in it. Read `AGENTS.md`, `CODING_STANDARDS.md` and `docs/agents/` in the worktree: the main checkout's copies can lag `origin/main`. Once its PR merges, `git worktree remove` it. The main checkout is shared with a human, so leave it on whatever they have checked out.

## Agent skills

### Issue tracker

Issues live in GitHub Issues on `believer/movies-go`, managed with the `gh` CLI. See `docs/agents/issue-tracker.md`.

### Triage labels

Default triage labels. See `docs/agents/triage-labels.md`.

### Domain docs

Single-context: `GLOSSARY.md` and `docs/adr/` at the repo root. See `docs/agents/domain.md`.
