FROM golang:1.23-bookworm AS base

WORKDIR /app/common

COPY common/go.mod ./

RUN go mod download

COPY common/ .

WORKDIR /app/mail

COPY mail/go.mod mail/go.sum ./

RUN go mod download

COPY mail/ .

RUN go build -o ./bin/pariksha_mail ./cmd/mail/main.go

EXPOSE 4010

CMD ["/app/mail/bin/pariksha_mail"]
