FROM golang:1.23-bookworm AS base

WORKDIR /app/common

COPY common/go.mod ./

RUN go mod download

COPY common/. .

WORKDIR /app/server

COPY server/go.mod server/go.sum ./

RUN go mod download

COPY server/. .

RUN go build -o ./bin/pariksha ./cmd/pariksha/main.go

EXPOSE 4000

CMD ["/app/server/bin/pariksha"]
