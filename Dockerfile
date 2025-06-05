# Build
ARG GO_VERSION=1
FROM golang:${GO_VERSION}-alpine as builder

WORKDIR /usr/src/app
COPY go.mod go.sum ./
RUN go mod download && go mod verify
COPY . .
RUN make build-prod

# App
FROM alpine:latest

COPY --from=builder /usr/src/app/db /db
COPY --from=builder /usr/src/app/public /public
COPY --from=builder /usr/src/app/views /views
COPY --from=builder /usr/src/app/oscars.csv /
COPY --from=builder /run-app /usr/local/bin/run-app

CMD ["run-app"]
