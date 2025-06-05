# Build the application
all: build

build:
	@templ generate
	@./hasher.sh
	@go build -o main main.go

dev:
	@templ generate
	@go build -o main main.go

css:
	@sed -i "" "s/styles\(\.[a-z0-9]\{6\}\)\{0,1\}\.css/styles.css/g" ./components/layout.templ
	@./scripts/tailwindcss -i styles.css -o public/styles.css --watch

# Live Reload
watch:
	air

build-prod:
	@templ generate
	@./hasher.sh
	@go build -v -o /run-app .
