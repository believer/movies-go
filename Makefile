.PHONY: css

# Build the application
all: build

CSS_SRC := $(shell find components -type f -name '*.css' | sort)
BASENAMES := $(notdir $(CSS_SRC))
TARGET := public/styles.css
TMPDIR := tmp

build:
	@./hasher.sh
	@templ generate
	@go build -o main main.go

dev:
	@mkdir -p "./public/components/" && cp $(CSS_SRC) "./public/components/" 
	@awk 'BEGIN {in_section=0} \
	{ \
		if ($$0 ~ /\/\* *START:GENERATED *\*\//) { \
			in_section=1; \
			print; \
			n = split("$(BASENAMES)", files, " "); \
			for (i = 1; i <= n; i++) { \
				printf "@import \"components/%s\" layer(components);\n", files[i] \
			} \
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
	@sed -i "" "s/styles\(\.[a-z0-9]\{6\}\)\{0,1\}\.css/styles.css/g" ./components/layout/html.templ
	@templ generate --watch \
		--open-browser=false \
		--proxy=http://localhost:8080 \
		--watch-pattern='.+\.(css|go|sql|templ)$$' \
		--cmd='go run .'

test:
	@go test ./...

clean:
	rm -f $(CSS_OUTPUT)
