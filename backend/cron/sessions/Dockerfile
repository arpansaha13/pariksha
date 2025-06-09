FROM golang:1.23-alpine AS builder

# Add required build dependencies
RUN apk add --no-cache gcc musl-dev

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

RUN --mount=type=cache,target=/gomod-cache \
    --mount=type=cache,target=/go-cache \
    go build -ldflags="-w -s" \
    -o ./bin/pariksha_cron_sessions ./cmd/cron/main.go

FROM alpine:3.22.0

RUN apk add --no-cache ca-certificates

COPY --from=builder /app/cron/sessions/bin/pariksha_cron_sessions /usr/local/bin/

CMD ["pariksha_cron_sessions"]
