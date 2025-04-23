FROM golang:1.23-alpine AS base

# Add required build dependencies
RUN apk add --no-cache gcc musl-dev

WORKDIR /app/common

COPY common/go.mod ./

RUN go mod download

COPY common/ .

WORKDIR /app/workers/exam

# https://medium.com/@marcin.niemira/optimise-docker-build-for-go-c03d6eb8b4b

ENV CGO_ENABLED=1
ENV GOCACHE=/go-cache
ENV GOMODCACHE=/gomod-cache

COPY workers/exam/go.mod workers/exam/go.sum ./
COPY workers/exam/cmd/ ./cmd/
COPY workers/exam/internal/ ./internal/

RUN --mount=type=cache,target=/gomod-cache \
	--mount=type=cache,target=/go-cache \
	go build -o ./bin/pariksha_exam_worker \
	./cmd/exam/main.go

CMD ["/app/workers/exam/bin/pariksha_exam_worker"]
