module github.com/arpansaha13/mail

go 1.23.4

require (
	github.com/arpansaha13/common v0.0.0-00010101000000-000000000000
	github.com/joho/godotenv v1.5.1
	google.golang.org/grpc v1.70.0
	google.golang.org/protobuf v1.36.5
)

require (
	golang.org/x/net v0.35.0 // indirect
	golang.org/x/sys v0.30.0 // indirect
	golang.org/x/text v0.22.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20250212204824-5a70512c5d8b // indirect
)

replace github.com/arpansaha13/common => ../common
replace github.com/arpansaha13/common/pkg/utils => ../common/pkg/utils
