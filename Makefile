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

# Rebuild movies_e2e from db/schema.sql and db/seed.sql, as postgres_test does
# when it starts
e2e-db: e2e-postgres
	@PSQL="$(PSQL)" SCHEMA=db/schema.sql SEED=db/seed.sql ./db/e2e.sh
	@echo "\033[0;32m(✓)\033[0;37m Rebuilt movies_e2e"

TEMPL := go run github.com/a-h/templ/cmd/templ@$(shell go list -m -f '{{.Version}}' github.com/a-h/templ)
TEMPL_SUMS = find . -name '*_templ.go' -not -path './node_modules/*' | sort | xargs shasum

# Fast checks that need no database: generated templ files are fresh, go vet
# and biome pass, and tests pass with database tests skipped
check:
	@before=$$($(TEMPL_SUMS)); \
		$(TEMPL) generate > /dev/null 2>&1; \
		if [ "$$before" != "$$($(TEMPL_SUMS))" ]; then \
			echo "❌ *_templ.go files were stale and have been regenerated, commit them"; \
			exit 1; \
		fi
	@go vet ./...
	@out=$$(TEST_DATABASE_URL= go test ./... 2>&1) || { echo "$$out"; exit 1; }
	@pnpm exec biome check .
	@echo "\033[0;32m(✓)\033[0;37m Checks passed"

test-all: test e2e

# Read-only check that production pages render, run after each deploy.
# Override the site with SMOKE_URL, e.g. make smoke SMOKE_URL=http://localhost:8080.
# Phony because the spec lives in smoke/.
.PHONY: smoke
smoke:
	@pnpm exec playwright test -c playwright.smoke.config.ts

# Playwright starts its own server on port 8090 against movies_e2e, with a
# TMDB stub on 8091, and rebuilds the database first, so this runs beside
# make dev. The server starts before the rebuild, so postgres_test has to be
# up already. Pass PSQL to use another Postgres, as for e2e-db.
#
# Locally Playwright reuses a server already on its port, even one another
# checkout started. Override the ports to run from a second checkout, e.g.
# make e2e E2E_PORT=8190 E2E_TMDB_PORT=8191. Every run rebuilds the one
# movies_e2e, though, so only one run at a time is supported.
e2e: e2e-postgres
	@pnpm exec playwright test

e2e-ui: e2e-postgres
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
