# Coding standards

Judgement calls for review. Anything a tool can check belongs in `make check` instead.

## SQL

- **Deterministic order.** A query that feeds a user-visible list ends its `ORDER BY` with a unique key, usually the row id. Ties then resolve the same way on every load and at any `LIMIT` cutoff.
- **Latest rating.** A user can rate a movie again, and the latest rating counts (`GLOSSARY.md`). Read ratings through a subquery that keeps one row per movie: `DISTINCT ON (movie_id) … WHERE user_id = … ORDER BY movie_id, created_at DESC, id DESC`. A plain join on `rating` returns one row per rating.

## Tests

- **Real database for queries.** Test repository queries against the test database with `db/dbtest`, so the SQL itself runs. Mocks stand in for repositories only in handler tests.
- **Rule out the near miss.** A test of a selection or ordering rule includes data where the plausible wrong rule gives a different answer. For "latest rating wins", make the later rating lower than the earlier one, so "highest rating wins" fails.
