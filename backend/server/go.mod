module pariksha/server

go 1.23.4

require (
	github.com/cenkalti/backoff/v5 v5.0.2
	github.com/gorilla/mux v1.8.1
	go.uber.org/zap v1.27.1
	google.golang.org/grpc v1.74.2
	google.golang.org/protobuf v1.36.8
	gopkg.in/go-playground/validator.v8 v8.18.2
	pariksha/common v0.0.0-00010101000000-000000000000
)

require (
	github.com/google/uuid v1.6.0 // indirect
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/jinzhu/now v1.1.5 // indirect
	go.uber.org/multierr v1.10.0 // indirect
	golang.org/x/net v0.42.0 // indirect
	golang.org/x/sys v0.35.0 // indirect
	golang.org/x/text v0.27.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20250728155136-f173205681a0 // indirect
	gopkg.in/go-playground/assert.v1 v1.2.1 // indirect
	gorm.io/gorm v1.25.12 // indirect
)

replace pariksha/common => ../common
