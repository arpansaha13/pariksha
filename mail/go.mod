module github.com/arpansaha13/mail

go 1.23.4

require (
	github.com/arpansaha13/common v0.0.0-00010101000000-000000000000
	github.com/rabbitmq/amqp091-go v1.10.0
)

replace github.com/arpansaha13/common => ../common

replace github.com/arpansaha13/common/pkg/utils => ../common/pkg/utils
