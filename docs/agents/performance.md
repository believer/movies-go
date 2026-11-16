# Performance checks

Time queries on a copy of the production backup with the benchmarks in `db/bench_test.go`. The comments above `bench-db` and `bench` in the `Makefile` cover the variables. This file holds the procedure.

## Setup

- The backup lives in `postgres_local`'s `movies`. `make db-up` loads it once.
- `make bench-db` copies `movies` into `movies_bench_<checkout>`, migrates the copy with the checkout's migrations and runs `VACUUM ANALYZE`, so the planner has statistics. It works while `make dev` has `movies` open. It prints the database's name; open it with `docker exec -it postgres_local psql -U admin -d <name>` to `EXPLAIN (ANALYZE, BUFFERS)` a query.
- `make bench` runs every benchmark ten times, in three to four minutes. Narrow it with `BENCH`, a `-bench` pattern: `make bench BENCH='Stats/GetRatings$|Genre/'`. It prints results only, so redirect it to a file outside the checkout.
- Without `BENCH_DATABASE_URL` the benchmarks skip, so `make check` and `make test` never run them.

## Compare a branch with `origin/main`

Each checkout benchmarks its own copy, migrated by its own migrations. So a branch whose migration changes a function or view, like `get_person_role_with_seen_json`, is timed against the old function in the base checkout and the new one in the branch. A branch without migrations gets two identical copies.

1. Add a base worktree beside the branch's: `git worktree add --detach ../movies-<slug>-base origin/main`, then `make bench-db` in both.
2. If the branch adds or changes benchmarks, copy them into the base: `git -C ../movies-<slug>-base restore --source=<branch> -- db/bench_test.go`. If the branch changed a method's signature, the copy won't compile, so time only the benchmarks the two share.
3. Alternate the runs, so drift on the machine (another agent's E2E run, a build) hits both sides. Narrow `BENCH` to the queries the branch changes:

   ```sh
   for i in 1 2; do
     make -C ../movies-<slug>-base bench BENCH='<pattern>' BENCH_COUNT=5 >> /tmp/<slug>-old.txt
     make bench BENCH='<pattern>' BENCH_COUNT=5 >> /tmp/<slug>-new.txt
   done
   ```

4. Compare: `go run golang.org/x/perf/cmd/benchstat@v0.0.0-20261009192801-be2c69fb417e /tmp/<slug>-old.txt /tmp/<slug>-new.txt`. benchstat merges the rounds in each file and prints the median, its spread, and `~` when the difference isn't significant. Paste its table into the PR's evidence.
5. Clean up: `git worktree remove ../movies-<slug>-base` (add `--force` when step 2 copied the benchmarks), then drop both copies, e.g. `docker exec postgres_local dropdb -U admin --force movies_bench_<checkout>`.

## Benchmarks

- A benchmark per repository, `Benchmark<Repository>`, holds a case per method, named after the method: `BenchmarkStats/GetRatings`. `TestBenchmarksCoverUserMethods` fails when a `Querier` method taking a `userID` has no case, so a new method gets one.
- Arguments come from `benchArgsQuery`, from the data rather than fixed IDs: the user with the most Watches, their most watched movie, the movie with the most cast and crew, and the genre, language, company, country, list, series (counting its sub-series) and person with the most movies. Ties go to the lowest ID, so both copies pick the same rows. Start there for a representative-ID query of your own. The join tables are `movie_genre`, `movie_language`, `movie_company`, `movie_country`, `movie_person`, `movie_series` and `official_list_movie`.
- The year is the one the user watched the most movies in, from 2012 on: Watches dated 2011 are a bulk import, about 1,900 of them on 36 days.
- Writes put the data back off the timer: deletes and inserts taking a transaction are rolled back, and the rest restore the rows or write their current values. `UpdateRating` still moves the Rating's `updated_at` to the time of the run, which no read compares on. An interrupted run can leave a change behind, so run `make bench-db` again before a comparison that matters.
