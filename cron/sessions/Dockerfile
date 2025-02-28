FROM golang:1.23-bookworm AS base

WORKDIR /app/common

COPY common/go.mod ./

RUN go mod download

COPY common/ .

WORKDIR /app/cron/sessions

# https://medium.com/@marcin.niemira/optimise-docker-build-for-go-c03d6eb8b4b

ENV CGO_ENABLED=1
ENV GOCACHE=/go-cache
ENV GOMODCACHE=/gomod-cache

COPY cron/sessions/go.mod cron/sessions/go.sum ./
COPY cron/sessions/cmd/ ./cmd/
COPY cron/sessions/internal/ ./internal/

RUN --mount=type=cache,target=/gomod-cache --mount=type=cache,target=/go-cache go build -o ./bin/pariksha_cron_sessions ./cmd/cron/main.go

CMD ["/app/cron/sessions/bin/pariksha_cron_sessions"]
