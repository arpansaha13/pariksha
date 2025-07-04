module pariksha/server

go 1.23.4

require (
	github.com/cenkalti/backoff/v5 v5.0.2
	github.com/gorilla/mux v1.8.1
	google.golang.org/grpc v1.71.0
	google.golang.org/protobuf v1.36.6
	gopkg.in/go-playground/validator.v8 v8.18.2
	pariksha/common v0.0.0-00010101000000-000000000000
)

require (
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/jinzhu/now v1.1.5 // indirect
	golang.org/x/net v0.34.0 // indirect
	golang.org/x/sys v0.29.0 // indirect
	golang.org/x/text v0.21.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20250115164207-1a7da9e5054f // indirect
	gopkg.in/go-playground/assert.v1 v1.2.1 // indirect
	gorm.io/gorm v1.25.12 // indirect
)

replace pariksha/common => ../common
