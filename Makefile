# Build the application
all: build

build:
	@./hasher.sh
	@templ generate
	@go build -o main main.go

dev:
	@sed -i "" "s/styles\(\.[a-z0-9]\{6\}\)\{0,1\}\.css/styles.css/g" ./components/layout/html.templ
	@templ generate --watch \
		--open-browser=false \
		--proxy=http://localhost:8080 \
		--watch-pattern='.+\.(css|go|sql|templ)$$' \
		--cmd='go run .'

test:
	@go test ./...
