FROM golang:1.23-bookworm AS base

WORKDIR /app/common

COPY common/go.mod ./

RUN go mod download

COPY common/ .

WORKDIR /app/mail

# https://medium.com/@marcin.niemira/optimise-docker-build-for-go-c03d6eb8b4b

ENV CGO_ENABLED=1
ENV GOCACHE=/go-cache
ENV GOMODCACHE=/gomod-cache

COPY mail/go.mod mail/go.sum ./
COPY mail/cmd/ ./cmd/
COPY mail/internal/ ./internal/

RUN --mount=type=cache,target=/gomod-cache --mount=type=cache,target=/go-cache go build -o ./bin/pariksha_mail ./cmd/mail/main.go

EXPOSE 4010

CMD ["/app/mail/bin/pariksha_mail"]
