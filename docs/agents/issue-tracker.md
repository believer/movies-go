# Issue tracker: GitHub

Issues and specs for this repo live in GitHub Issues on **`believer/movies-go`**. Use the `gh` CLI for all operations: the GitHub MCP token is read-only, so its writes fail with 403. Run it from the repo root and it picks up the repo automatically. Otherwise pass `-R believer/movies-go`.

## Conventions

- **Create an issue**: `gh issue create --title "..." --body-file <file> [--label ...]`. Add `bug`, `enhancement` (feature) or `Improvement` when the type is clear.
- **Read an issue**: `gh issue view 123 --comments` for the body plus the discussion.
- **List issues**: `gh issue list --state open [--label ...] [--assignee ...] --json number,title,labels,assignees`. Raise `--limit` (default 30) when you need everything.
- **Issue ids vs numbers**: the sub-issue and dependency endpoints take the issue's database **id**, not its number. Get it with `gh api repos/{owner}/{repo}/issues/<n> --jq .id`.
- **Make an issue a sub-issue of a parent**: `gh api repos/{owner}/{repo}/issues/<parent>/sub_issues -F sub_issue_id=<child id>`. List children in order with `gh api repos/{owner}/{repo}/issues/<parent>/sub_issues`.
- **Blocking**: GitHub's native dependencies. `gh api repos/{owner}/{repo}/issues/<n>/dependencies/blocked_by -F issue_id=<blocker id>` marks `#n` as blocked by the blocker. A ticket is unblocked when every blocker is closed.
- **Count open blockers**: `gh api repos/{owner}/{repo}/issues/<n>/dependencies/blocked_by --jq '[.[] | select(.state == "open")] | length'`.
- **Comment on an issue**: `gh issue comment 123 --body-file <file>`.
- **Apply / remove labels**: `gh issue edit 123 --add-label ...` / `--remove-label ...`.
- **Close**: post the explanation as a comment, then `gh issue close 123` (completed), `gh issue close 123 --reason "not planned"` (won't do), or `gh issue close 123 --duplicate-of 456`.

An issue is open or closed. An assignee means someone has claimed it. The triage labels in `triage-labels.md` cover the remaining states.

File issues only in `believer/movies-go`. Never file in `mattpocock/skills` (the repo these skills come from).

## Pull requests as a triage surface

**PRs as a request surface: yes.** Treat open PRs on `believer/movies-go` like issues when triaging: read with `gh pr view <n> --comments`, list with `gh pr list`. Triage labels apply to PRs too. Link a PR to its issue with `Fixes #123` in the PR body so the issue closes on merge.

## When a skill says "publish to the issue tracker"

Create a GitHub issue in `believer/movies-go`.

## When a skill says "fetch the relevant ticket"

Read it as in **Read an issue** above.

## Wayfinding operations

Used by `/wayfinder`. The **map** is a single issue with **sub-issues** as tickets.

- **Map**: a single issue labelled `wayfinder:map`, holding the Notes / Decisions-so-far / Fog body.
- **Child ticket**: a sub-issue of the map (see **Make an issue a sub-issue**). Label: `wayfinder:<type>` (`research`/`prototype`/`grilling`/`task`). Once claimed, the ticket is assigned to the driving dev.
- **Blocking**: see **Blocking** under Conventions.
- **Frontier query**: `gh api repos/{owner}/{repo}/issues/<map>/sub_issues --jq '.[] | select(.state == "open" and (.assignees | length) == 0) | .number'`. Then drop any with an open blocker (see **Count open blockers**). The first one left, in map order, wins.
- **Claim**: `gh issue edit <n> --add-assignee @me`. This is the session's first write.
- **Resolve**: comment the answer, close the issue, then add a context pointer (gist + link) to the map's Decisions-so-far: `gh issue view <map> --json body --jq .body` to read it, then `gh issue edit <map> --body-file <file>` with the updated body.
