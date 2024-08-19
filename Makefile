# Build the application
all: build

build:
	@echo "Building..."
	@templ generate
	@tailwindcss -i styles.css -o public/styles.css
	@go build -o main main.go

# Live Reload
watch:
	air


