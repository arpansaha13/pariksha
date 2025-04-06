FROM golang:1.23-alpine AS base

# Add required build dependencies
RUN apk add --no-cache gcc musl-dev

WORKDIR /app/common

COPY common/go.mod ./

RUN go mod download

COPY common/ .

WORKDIR /app/workers/exam_questions

# https://medium.com/@marcin.niemira/optimise-docker-build-for-go-c03d6eb8b4b

ENV CGO_ENABLED=1
ENV GOCACHE=/go-cache
ENV GOMODCACHE=/gomod-cache

COPY workers/exam_questions/go.mod workers/exam_questions/go.sum ./
COPY workers/exam_questions/cmd/ ./cmd/
COPY workers/exam_questions/internal/ ./internal/

RUN --mount=type=cache,target=/gomod-cache \
	--mount=type=cache,target=/go-cache \
	go build -o ./bin/pariksha_exam_questions_worker \
	./cmd/exam_questions/main.go

CMD ["/app/workers/exam_questions/bin/pariksha_exam_questions_worker"]
