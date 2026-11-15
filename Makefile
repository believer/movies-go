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

TEST_DATABASE_URL ?= postgres://admin:password@localhost:5433/movies_test?sslmode=disable
PKG ?= ./...

# Packages share the test database, so run them one at a time. Narrow with
# PKG and RUN, e.g. make test PKG=./db RUN=TestGetBestOfTheYear
test:
	@docker-compose up -d --wait postgres_test
	@TEST_DATABASE_URL="$(TEST_DATABASE_URL)" go test -p 1 $(if $(RUN),-run '$(RUN)') $(PKG)

# Runs psql in the postgres_test container. Override it to reach another
# server, e.g. make e2e-db PSQL="psql -h localhost -U admin", and the
# container isn't started.
PSQL ?= docker exec -i postgres_test psql -U admin

# Start postgres_test, unless PSQL points elsewhere
e2e-postgres:
	$(if $(filter default,$(origin PSQL)),@docker-compose up -d --wait postgres_test)

# Each E2E server port gets a database of its own, so checkouts on different
# ports run E2E side by side. tests/env.ts names it the same way.
E2E_PORT ?= 8090
E2E_DB = movies_e2e_$(E2E_PORT)

# Rebuild the E2E database from db/schema.sql and db/seed.sql, as postgres_test
# does for the default port when it starts
e2e-db: e2e-postgres
	@PSQL="$(PSQL)" SCHEMA=db/schema.sql SEED=db/seed.sql DB=$(E2E_DB) ./db/e2e.sh
	@echo "\033[0;32m(✓)\033[0;37m Rebuilt $(E2E_DB)"

TEMPL := go run github.com/a-h/templ/cmd/templ@$(shell go list -m -f '{{.Version}}' github.com/a-h/templ)
TEMPL_SUMS = find . -name '*_templ.go' -not -path './node_modules/*' | sort | xargs shasum

# Pinned, and built with go.mod's Go: an installed mockery built with an older
# Go can't read this Go's export data
MOCKERY := go run github.com/vektra/mockery/v3@v3.9.1
MOCK_SUMS = find mocks -name '*.go' | sort | xargs shasum

# Regenerate mocks/ from the interfaces in db/. Phony, since mocks/ exists.
.PHONY: mocks
mocks:
	@out=$$($(MOCKERY) 2>&1) || { echo "$$out"; exit 1; }

# Fast checks that need no database: generated templ files and mocks are fresh,
# Go files are gofmt'd, go vet and biome pass, and tests pass with database
# tests skipped
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
			echo "❌ mocks/ was stale and has been regenerated with make mocks, commit it"; \
			exit 1; \
		fi
	@unformatted=$$(gofmt -l $$(git ls-files '*.go')); \
		if [ -n "$$unformatted" ]; then \
			echo "❌ Not gofmt'd, run gofmt -w on:"; \
			echo "$$unformatted"; \
			exit 1; \
		fi
	@go vet ./...
	@out=$$(go test -short ./... 2>&1) || { echo "$$out"; exit 1; }
	@pnpm exec biome check .
	@echo "\033[0;32m(✓)\033[0;37m Checks passed"

test-all: test e2e

# Read-only check that production pages render, run after each deploy.
# Override the site with SMOKE_URL, e.g. make smoke SMOKE_URL=http://localhost:8080.
# Phony because the spec lives in smoke/.
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

db-up:
	@docker-compose up -d
	@docker exec \
		-i postgres_local psql \
		-U admin \
		-d movies \
		< ~/Library/Mobile\ Documents/com~apple~CloudDocs/Backups/Movie\ DB/movies_2026-03-06.sql

db-down:
	@docker-compose down -v

# Refresh db/schema.sql from the database in .env (production). pg_dump runs
# in a Postgres 15 container to match production, so the dump loads on 15.
# The restrict lines carry a random token per dump, so they are dropped to
# keep diffs clean.
schema:
	@set -a && . ./.env && set +a && \
		docker run --rm postgres:15 pg_dump "$${DATABASE_URL%\?*}" --schema-only --no-owner --no-privileges \
		| sed -e '/^\\restrict /d' -e '/^\\unrestrict /d' > db/schema.sql
	@echo "\033[0;32m(✓)\033[0;37m Dumped db/schema.sql"
