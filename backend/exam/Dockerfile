FROM golang:1.23-alpine AS builder

# Add required build dependencies
RUN apk add --no-cache gcc musl-dev

WORKDIR /app/common

COPY common/go.mod ./

RUN go mod download

COPY common/ .

WORKDIR /app/exam

# https://medium.com/@marcin.niemira/optimise-docker-build-for-go-c03d6eb8b4b

ENV CGO_ENABLED=1
ENV GOCACHE=/go-cache
ENV GOMODCACHE=/gomod-cache

COPY exam/go.mod exam/go.sum ./
COPY exam/cmd/ ./cmd/
COPY exam/internal/ ./internal/

RUN --mount=type=cache,target=/gomod-cache \
    --mount=type=cache,target=/go-cache \
    go build -ldflags="-w -s" \
    -o ./bin/pariksha_exam ./cmd/exam/main.go

FROM alpine:3.22.0

RUN apk add --no-cache ca-certificates

COPY --from=builder /app/exam/bin/pariksha_exam /usr/local/bin/

EXPOSE 4000

CMD ["pariksha_exam"]
