module pariksha/mail

go 1.23.4

require (
	github.com/rabbitmq/amqp091-go v1.10.0
	pariksha/common v0.0.0-00010101000000-000000000000
)

replace pariksha/common => ../common

replace pariksha/common/pkg/utils => ../common/pkg/utils
