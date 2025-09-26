# Build the application
all: build

build:
	@templ generate
	@./hasher.sh
	@go build -o main main.go

dev:
	@templ generate --watch \
		--proxy=http://localhost:8080 \
		--watch-pattern='.+\.(css|go|sql|templ)$$' \
		--cmd='go run .'

css:
	@sed -i "" "s/styles\(\.[a-z0-9]\{6\}\)\{0,1\}\.css/styles.css/g" ./components/layout/layout.templ
	@./scripts/tailwindcss -i styles.css -o public/styles.css --watch

build-prod:
	@templ generate
	@./hasher.sh
	@go build -v -o /run-app .

test:
	@go test ./...
