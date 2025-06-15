package grpcerror

import (
	"fmt"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Internal constructs a gRPC error with an internal status code,
// combining a custom message and the original error details.
func Internal(err error, message string) error {
	return status.Error(codes.Internal, fmt.Sprintf("%s: %s", message, err.Error()))
}
