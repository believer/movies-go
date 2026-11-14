# Issue tracker: Linear

Issues and specs for this repo live in Linear, in team **NAT** ("Natt och Dag"). The team is shared with other work, so this repo's issues are scoped by the **`Movies`** label. Use the Linear MCP tools (`tools.linear.*`) for all operations.

## Conventions

- **Create an issue**: `save_issue({ team: "NAT", title, description, labels: ["Movies", ...] })`. Always include `Movies`. Add `Bug`, `Feature` or `Improvement` when the type is clear.
- **Read an issue**: `get_issue({ id: "NAT-123" })`, then `list_comments({ issueId: "NAT-123" })` for the discussion (`get_issue` does not include comments).
- **List issues**: `list_issues({ team: "NAT", label: "Movies", ... })` with `state` / `parentId` / `assignee` filters. Results are paginated; follow the cursor while `hasNextPage` is true.
- **Keep reads lean**: `get_issue` and `list_issues` accept `fields`. Pass only what the step needs, e.g. `fields: ["id", "title", "status", "labels", "assignee"]` for a list; add `"description"` or `"relations"` on `get_issue` when you read one issue in depth (`list_issues` has no `relations` field).
- **Make an issue a sub-issue of a parent**: `save_issue({ ..., parentId: "NAT-<parent>" })`.
- **Blocking**: Linear's native relations. `save_issue({ ..., blockedBy: ["NAT-<n>"] })` on create, or on update with `id`. A ticket is unblocked when every blocker is in a `completed` or `canceled` state.
- **Count open blockers**: `get_issue({ id, fields: ["relations"] })`, then count the blockers whose state type is not `completed` / `canceled`.
- **Comment on an issue**: `save_comment({ issueId: "NAT-123", body })`.
- **Apply / remove labels**: `save_issue({ id, addLabels: [...] })` / `save_issue({ id, removeLabels: [...] })`. Never pass `labels` on update; it replaces the whole set and would drop `Movies`.
- **Close**: post the explanation as a comment, then `save_issue({ id, state: "Done" })` (or `"Canceled"` for won't-do, `"Duplicate"` with `duplicateOf`).

Workflow states in NAT: `Backlog`, `Todo`, `In Progress`, `In Review`, `Done`, `Canceled`, `Duplicate`.

File issues only into team NAT. Never file into `mattpocock/skills` (the repo these skills come from).

## Pull requests as a triage surface

**PRs as a request surface: no.** _(PRs live on GitHub at `believer/movies-go`; issues live in Linear.)_

## When a skill says "publish to the issue tracker"

Create a Linear issue in team NAT with the `Movies` label.

## When a skill says "fetch the relevant ticket"

Read it as in **Read an issue** above.

## Wayfinding operations

Used by `/wayfinder`. The **map** is a single issue with **child** issues as tickets.

- **Map**: a single issue labelled `Movies` + `wayfinder:map`, holding the Notes / Decisions-so-far / Fog body.
- **Child ticket**: an issue with `parentId` set to the map. Labels: `Movies` + `wayfinder:<type>` (`research`/`prototype`/`grilling`/`task`). Once claimed, the ticket is assigned to the driving dev.
- **Blocking**: see **Blocking** under Conventions.
- **Frontier query**: list the map's open children with `list_issues({ parentId: "NAT-<map>", fields: ["id", "title", "status", "assignee"] })`, drop any with an assignee, then any with an open blocker (see **Count open blockers**); first in map order wins.
- **Claim**: `save_issue({ id, assignee: "me", state: "In Progress" })`, the session's first write.
- **Resolve**: comment the answer, set state `Done`, then append a context pointer (gist + link) to the map's Decisions-so-far (use `patch` on the map description).
