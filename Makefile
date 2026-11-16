# Build the application
all: build

build:
	@templ generate
	@go build -o main main.go

PORT ?= 8080
PROXY_PORT ?= 7331

# Override PORT and PROXY_PORT to run beside another checkout, e.g.
# make dev PORT=8099 PROXY_PORT=7399. A change to any watched file restarts
# --cmd, which rebuilds the embedded stylesheet and works out its hash again.
dev:
	@PORT=$(PORT) templ generate --watch \
		--open-browser=false \
		--proxy=http://localhost:$(PORT) \
		--proxyport=$(PROXY_PORT) \
		--watch-pattern='.+\.(css|go|sql|templ)$$' \
		--ignore-pattern='playwright-report' \
		--cmd='go run .'

# Runs psql in the postgres_test container. Override it to reach another
# server, e.g. make e2e-db PSQL="psql -h localhost -U admin", and the
# container isn't started.
PSQL ?= docker exec -i postgres_test psql -U admin

# Start postgres_test, unless PSQL points elsewhere
test-postgres:
	$(if $(filter file,$(origin PSQL)),@docker-compose up -d --wait postgres_test)

# Each checkout gets test and bench databases of its own, named after its
# directory plus a hash of its path, e.g. movies_test_movies_api_3f9a1c, so
# checkouts run make test and make bench side by side. Databases of removed
# checkouts stay until dropped by hand.
CHECKOUT := $(shell printf '%s' '$(notdir $(CURDIR))' | tr 'A-Z' 'a-z' | tr -c 'a-z0-9' '_' | cut -c1-30)_$(shell printf '%s' '$(CURDIR)' | shasum | cut -c1-6)
CHECKOUT_TEST_DB := movies_test_$(CHECKOUT)
TEST_DB ?= $(CHECKOUT_TEST_DB)

# Rebuild TEST_DB from this checkout's db/schema.sql, so a branch that
# changes the schema tests against it
test-db: test-postgres
	@PSQL="$(PSQL)" SCHEMA=db/schema.sql DB=$(TEST_DB) ./db/rebuild.sh

TEST_DATABASE_URL ?= postgres://admin:password@localhost:5433/$(TEST_DB)?sslmode=disable
PKG ?= ./...

# Packages share the checkout's test database, so run them one at a time.
# Narrow with PKG and RUN, e.g. make test PKG=./db RUN=TestGetBestOfTheYear.
# Pointing TEST_DATABASE_URL at another server needs PSQL pointed there too,
# and TEST_DB set to the database in the URL. -count=1 skips go test's cache,
# which can't tell that test-db rebuilt the database.
test: test-db
	@TEST_DATABASE_URL="$(TEST_DATABASE_URL)" go test -count=1 -p 1 $(if $(RUN),-run '$(RUN)') $(PKG)

# Benchmarks run on a copy of the production backup in postgres_local, so
# they need make db-up once. See docs/agents/performance.md.
BENCH_DB ?= movies_bench_$(CHECKOUT)
BENCH_DATABASE_URL ?= postgres://admin:password@localhost:5432/$(BENCH_DB)?sslmode=disable
LOCAL_PSQL := docker exec -i postgres_local psql -U admin -v ON_ERROR_STOP=1 -q

# Copy movies into BENCH_DB, migrate it with this checkout's migrations and
# VACUUM ANALYZE it, so the planner has statistics. pg_dump reads a snapshot
# while make dev or other clients have movies open, which CREATE DATABASE
# ... TEMPLATE movies can't.
bench-db:
	@$(LOCAL_PSQL) -d postgres -c 'SET client_min_messages = warning' -c 'DROP DATABASE IF EXISTS $(BENCH_DB) WITH (FORCE)' -c 'CREATE DATABASE $(BENCH_DB)'
	@docker exec postgres_local pg_dump -U admin movies | $(LOCAL_PSQL) -d $(BENCH_DB) > /dev/null
	@DATABASE_URL="$(BENCH_DATABASE_URL)" go run ./cmd/migrate
	@$(LOCAL_PSQL) -d $(BENCH_DB) -c 'VACUUM ANALYZE'
	@echo "\033[0;32m(✓)\033[0;37m Built $(BENCH_DB)"

BENCH ?= .
BENCH_COUNT ?= 10
# The slowest queries take under 100ms, so a quarter second still times a
# few calls, and all benchmarks run ten times in three to four minutes
BENCH_TIME ?= 0.25s

# Run the benchmarks in db/ against BENCH_DB, built by make bench-db. Narrow
# with BENCH, e.g. make bench BENCH=Stats/GetRatings, and compare two runs
# with benchstat: make bench > /tmp/new.txt.
bench:
	@BENCH_DATABASE_URL="$(BENCH_DATABASE_URL)" go test ./db -run '^$$' -bench '$(BENCH)' -benchmem -benchtime=$(BENCH_TIME) -count=$(BENCH_COUNT)

# Each E2E server port gets a database of its own, so checkouts on different
# ports run E2E side by side. e2e/env.ts names it the same way.
E2E_PORT ?= 8090
E2E_DB = movies_e2e_$(E2E_PORT)

# Rebuild the E2E database from db/schema.sql and db/seed.sql
e2e-db: test-postgres
	@PSQL="$(PSQL)" SCHEMA=db/schema.sql SEED=db/seed.sql DB=$(E2E_DB) ./db/rebuild.sh
	@echo "\033[0;32m(✓)\033[0;37m Rebuilt $(E2E_DB)"

TEMPL := go run github.com/a-h/templ/cmd/templ@$(shell go list -m -f '{{.Version}}' github.com/a-h/templ)
TEMPL_SUMS = find . -name '*_templ.go' -not -path './node_modules/*' | sort | xargs shasum

# Pinned, and built with go.mod's Go: an installed mockery built with an older
# Go can't read this Go's export data
MOCKERY := go run github.com/vektra/mockery/v3@v3.9.1
MOCK_SUMS = find db/mocks -name '*.go' | sort | xargs shasum

# Pinned like mockery. Runs go vet with staticcheck, errcheck and the
# forbidigo rules in .golangci.yml.
GOLANGCI_LINT := go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.15.0

# Regenerate db/mocks/ from the interfaces in db/. Phony, since db/mocks/ exists.
.PHONY: mocks
mocks:
	@out=$$($(MOCKERY) 2>&1) || { echo "$$out"; exit 1; }

# Fast checks that need no database: generated templ files and mocks are fresh,
# Go files are gofmt'd, golangci-lint and biome pass, and tests pass with
# database tests skipped
check:
	@before=$$($(TEMPL_SUMS)); \
		$(TEMPL) generate > /dev/null 2>&1; \
		if [ "$$before" != "$$($(TEMPL_SUMS))" ]; then \
			echo "❌ *_templ.go files were stale and have been regenerated, commit them"; \
			exit 1; \
		fi
	@before=$$($(MOCK_SUMS)); \
		$(MAKE) -s mocks || exit 1; \
		if [ "$$before" != "$$($(MOCK_SUMS))" ]; then \
			echo "❌ db/mocks/ was stale and has been regenerated with make mocks, commit it"; \
			exit 1; \
		fi
	@unformatted=$$(gofmt -l $$(git ls-files '*.go')); \
		if [ -n "$$unformatted" ]; then \
			echo "❌ Not gofmt'd, run gofmt -w on:"; \
			echo "$$unformatted"; \
			exit 1; \
		fi
	@out=$$($(GOLANGCI_LINT) run ./... 2>&1) || { echo "$$out"; exit 1; }
	@out=$$(go test -short ./... 2>&1) || { echo "$$out"; exit 1; }
	@pnpm exec biome check .
	@echo "\033[0;32m(✓)\033[0;37m Checks passed"

test-all: test e2e

# Read-only check that production pages render, run after each deploy.
# Override the site with SMOKE_URL, e.g. make smoke SMOKE_URL=http://localhost:8080.
# Phony because the spec lives in e2e/smoke/.
.PHONY: smoke
smoke:
	@pnpm exec playwright test -c playwright.smoke.config.ts

# Playwright starts its own server on port 8090 against movies_e2e_8090, with
# a TMDB stub on 8091, so this runs beside make dev. Playwright's global setup
# rebuilds the database, but the server starts before it, so the database is
# built here first. Pass PSQL to use another Postgres, as for e2e-db.
#
# A server already on the port fails the run; set E2E_REUSE=1 to reuse it.
# Override the port to run beside another checkout, e.g. make e2e E2E_PORT=8190.
# The TMDB stub follows on 8191 unless E2E_TMDB_PORT is set, and the run gets
# movies_e2e_8190, so the two don't share data.
e2e: e2e-db
	@pnpm exec playwright test

e2e-ui: e2e-db
	@pnpm exec playwright test --ui

# What a change needs before review: make check, make test and the Chromium
# E2E suite. Stops at the first failure and prints only its output. Pass
# E2E_PORT as for make e2e to run beside another checkout.
verify: check
	@out=$$($(MAKE) -s test 2>&1) || { echo "$$out"; exit 1; }
	@echo "\033[0;32m(✓)\033[0;37m Tests passed"
	@out=$$($(MAKE) -s e2e-db 2>&1) || { echo "$$out"; exit 1; }
	@out=$$(pnpm exec playwright test --project=chromium 2>&1) || { echo "$$out"; exit 1; }; \
		echo "\033[0;32m(✓)\033[0;37m E2E passed (Chromium):$$(echo "$$out" | grep -E '^ +[0-9]+ (passed|skipped|flaky)' | tr -s ' \n' ' ')"

db-up:
	@docker-compose up -d
	@docker exec \
		-i postgres_local psql \
		-U admin \
		-d movies \
		< ~/Library/Mobile\ Documents/com~apple~CloudDocs/Backups/Movie\ DB/movies_2026-03-06.sql

db-down:
	@docker-compose down -v

# Rebuild db/schema.sql from the migrations, in a scratch Postgres 15
# container (db/schema.sh). Run it after adding a migration.
schema:
	@./db/schema.sh > db/schema.sql.tmp && mv db/schema.sql.tmp db/schema.sql
	@echo "\033[0;32m(✓)\033[0;37m Dumped db/schema.sql"

# Add the next migration, e.g. make migration NAME=add_latest_rating.
# Version 1 is the baseline in db/migrate.go, so files start at 2.
migration:
	@test -n "$(NAME)" || { echo "❌ Pass NAME, e.g. make migration NAME=add_latest_rating"; exit 1; }
	@last=$$(ls db/migrations | sed -n 's/^0*\([0-9]*\)_.*\.sql$$/\1/p' | sort -n | tail -n 1); \
		file=db/migrations/$$(printf '%05d' $$(( $${last:-1} + 1 )))_$(NAME).sql; \
		printf -- '-- +goose Up\n' > $$file; \
		echo "\033[0;32m(✓)\033[0;37m Created $$file"
