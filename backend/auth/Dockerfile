FROM golang:1.23-alpine AS base

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
	go build -o ./bin/pariksha_auth ./cmd/auth/main.go

EXPOSE 4000

CMD ["/app/auth/bin/pariksha_auth"]
