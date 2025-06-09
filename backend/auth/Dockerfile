FROM golang:1.23-alpine AS builder

# Add required build dependencies
RUN apk add --no-cache gcc musl-dev

WORKDIR /app/common

COPY common/go.mod ./

RUN go mod download

COPY common/ .

WORKDIR /app/auth

# https://medium.com/@marcin.niemira/optimise-docker-build-for-go-c03d6eb8b4b

ENV CGO_ENABLED=1
ENV GOCACHE=/go-cache
ENV GOMODCACHE=/gomod-cache

COPY auth/go.mod auth/go.sum ./
COPY auth/cmd/ ./cmd/
COPY auth/internal/ ./internal/

RUN --mount=type=cache,target=/gomod-cache \
    --mount=type=cache,target=/go-cache \
    go build -ldflags="-w -s" \
    -o ./bin/pariksha_auth ./cmd/auth/main.go

FROM alpine:3.22.0

RUN apk add --no-cache ca-certificates

COPY --from=builder /app/auth/bin/pariksha_auth /usr/local/bin/

EXPOSE 4000

CMD ["pariksha_auth"]
