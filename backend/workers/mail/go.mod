module pariksha/mail

go 1.23.4

require (
	github.com/rabbitmq/amqp091-go v1.10.0
	pariksha/common v0.0.0-00010101000000-000000000000
)

require (
	golang.org/x/sys v0.29.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20250115164207-1a7da9e5054f // indirect
	google.golang.org/grpc v1.71.0 // indirect
	google.golang.org/protobuf v1.36.6 // indirect
)

replace pariksha/common => ../../common
