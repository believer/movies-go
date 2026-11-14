# Coding standards

Judgement calls for review. Anything a tool can check belongs in `make check` instead.

## SQL

- **Deterministic order.** A query that feeds a user-visible list ends its `ORDER BY` with a unique key, usually the row id. Ties then resolve the same way on every load and at any `LIMIT` cutoff.
- **Latest rating.** A user can rate a movie again, and the latest rating counts (`GLOSSARY.md`). Read ratings through a subquery that keeps one row per movie: `DISTINCT ON (movie_id) … WHERE user_id = … ORDER BY movie_id, created_at DESC, id DESC`. A plain join on `rating` returns one row per rating.

## CSS

The stylesheet is built from `styles/`, `components/` and `views/`, and a file's directory decides its cascade layer (`docs/adr/0001-go-builds-the-stylesheet.md`).

- **Beside its one user.** A rule used by one component or view lives in a CSS file beside that component or view.
- **Shared view classes.** A views-layer class used by several views gets its own file in `views/`, named after the class (`views/overview.css`).
- **Shared component classes.** A components-layer class with no component of its own goes in `styles/components.css`.
- **Moves change layers.** Moving a rule between `styles/`, `components/` and `views/` moves it to another layer, so it can win or lose against different rules. The PR says why the new layer is right.

## Tests

- **Real database for queries.** Test repository queries against the test database with `db/dbtest`, so the SQL itself runs. Mocks stand in for repositories only in handler tests.
- **E2E tests assert on the seed.** Every value an E2E test expects comes from `db/seed.sql`, never from whatever data a database happens to hold. A test that needs more data adds rows to the seed. Assert on fixed past years; check views of the current year for structure only.
- **E2E writes run as a fresh user.** A test that writes uses the `user` fixture from `tests/fixtures.ts`, so it only changes data of its own and runs in parallel with the rest. Expected values then come from the seed or the test's own setup. Movie intake uses a made-up movie from `tests/tmdb/fixtures/`, a different one per test.
- **Rule out the near miss.** A test of a selection or ordering rule includes data where the plausible wrong rule gives a different answer. For "latest rating wins", make the later rating lower than the earlier one, so "highest rating wins" fails.
