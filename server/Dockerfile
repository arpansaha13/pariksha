FROM golang:1.23-bookworm AS base

WORKDIR /app/common

COPY common/go.mod ./

RUN go mod download

COPY common/ .

WORKDIR /app/server

# https://medium.com/@marcin.niemira/optimise-docker-build-for-go-c03d6eb8b4b

ENV CGO_ENABLED=1
ENV GOCACHE=/go-cache
ENV GOMODCACHE=/gomod-cache

COPY server/go.mod server/go.sum ./
COPY server/gowatch.yml ./
COPY server/cmd/ ./cmd/
COPY server/internal/ ./internal/

RUN --mount=type=cache,target=/gomod-cache --mount=type=cache,target=/go-cache go build -o ./bin/pariksha ./cmd/pariksha/main.go

EXPOSE 4000

CMD ["/app/server/bin/pariksha"]
