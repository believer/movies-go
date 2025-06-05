# Install
ARG GO_VERSION=1
FROM golang:${GO_VERSION}-alpine as install

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download && go mod verify

# Templ generate
FROM ghcr.io/a-h/templ:latest AS generate
COPY --chown=65532:65532 . /app
WORKDIR /app
RUN ["templ", "generate"]

# Build
FROM golang:latest AS build
COPY --from=generate /app /app
WORKDIR /app
RUN GOOS=linux go build -v -o /run-app

# App
FROM alpine:latest

COPY --from=generate /app/db /db
COPY --from=generate /app/public /public
COPY --from=generate /app/views /views
COPY --from=generate /app/oscars.csv /
COPY --from=build /run-app /usr/local/bin/run-app

CMD ["run-app"]
