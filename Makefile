.PHONY: css

# Build the application
all: build

CSS_SRC := $(shell find components -type f -name '*.css' | sort)
TARGET := public/styles.css

css:
	@awk 'BEGIN {in_section=0} \
	{ \
		if ($$0 ~ /\/\* *START:GENERATED *\*\//) { \
			in_section=1; \
			print; \
			system("sed \"s/^/	/\" $(CSS_SRC)"); \
			next \
		} \
		if ($$0 ~ /\/\* *END:GENERATED *\*\//) { \
			in_section=0; \
			print; \
			next \
		} \
		if (in_section) { next } \
		print \
	}' $(TARGET) > $(TARGET).new
	@mv $(TARGET).new $(TARGET)
	@echo "\033[0;32m(✓)\033[0;37m Copied component CSS"

build: css
	@./hasher.sh
	@templ generate
	@go build -o main main.go

dev: css
	@sed -i "" "s/styles\(\.[a-z0-9]\{6\}\)\{0,1\}\.css/styles.css/g" ./components/layout/html.templ
	@templ generate --watch \
		--open-browser=false \
		--proxy=http://localhost:8080 \
		--watch-pattern='.+\.(css|go|sql|templ)$$' \
		--ignore-pattern='playwright-report' \
		--cmd='go run .'

TEST_DATABASE_URL ?= postgres://admin:password@localhost:5433/movies_test?sslmode=disable

# Packages share the test database, so run them one at a time
test:
	@docker-compose up -d --wait postgres_test
	@TEST_DATABASE_URL="$(TEST_DATABASE_URL)" go test -p 1 ./...

smoke: test e2e

e2e:
	@E2E_URL=http://localhost:8080 pnpm exec playwright test

e2e-ui:
	@E2E_URL=http://localhost:8080 pnpm exec playwright test --ui

clean:
	rm -f $(CSS_OUTPUT)

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
