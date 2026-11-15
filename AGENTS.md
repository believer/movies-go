Run `make check` before finishing a change. A change to handlers, views, templates or `db/` also needs `make test` and the E2E suite.

**E2E**: running or writing Playwright tests → `docs/agents/e2e.md`.

Implement in a worktree branched from a fresh `origin/main` (`git fetch` first). The main checkout is shared with a human, so leave it on whatever they have checked out. Run `pnpm install` in the worktree before `make check`.

## Agent skills

### Issue tracker

Issues live in GitHub Issues on `believer/movies-go`, managed with the `gh` CLI. See `docs/agents/issue-tracker.md`.

### Triage labels

Default triage labels. See `docs/agents/triage-labels.md`.

### Domain docs

Single-context: `GLOSSARY.md` and `docs/adr/` at the repo root. See `docs/agents/domain.md`.
