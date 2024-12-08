# Build the application
all: build

build:
	@echo "Building..."
	@templ generate
	@npx @tailwindcss/cli -i styles.css -o public/styles.css
	@go build -o main main.go

# Live Reload
watch:
	air


