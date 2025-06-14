ARG GO_VERSION=1

# Install
FROM golang:${GO_VERSION}-alpine AS install

COPY . /app
WORKDIR /app
RUN go mod download && go mod verify

# Build
FROM golang:${GO_VERSION} AS build

COPY --from=install /app /app
WORKDIR /app
RUN CGO_ENABLED=0 GOOS=linux go build -buildvcs=false -o movies

# App
FROM alpine:latest AS deploy

WORKDIR /app
COPY --from=build /app/db /app/db
COPY --from=build /app/public /app/public
COPY --from=build /app/views /app/views
COPY --from=build /app/oscars.csv /app
COPY --from=build /app/movies /app

CMD ["/app/movies"]
