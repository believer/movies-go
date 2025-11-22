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

test:
	@go test ./...

smoke: test e2e

e2e:
	@E2E_URL=http://localhost:8080 pnpm exec playwright test

e2e-ui:
	@E2E_URL=http://localhost:8080 pnpm exec playwright test --ui

clean:
	rm -f $(CSS_OUTPUT)
