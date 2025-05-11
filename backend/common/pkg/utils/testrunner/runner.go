package testrunner

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Runner executes a gRPC test case with common error handling
func Runner[Req any, Resp any](
	t *testing.T,
	ctx context.Context,
	expectedCode codes.Code,
	req Req,
	call func(context.Context, Req, ...grpc.CallOption) (Resp, error),
	validate func(*testing.T, Resp),
) {
	resp, err := call(ctx, req)

	if expectedCode != codes.OK {
		assert.Equal(t, expectedCode, status.Code(err))
		return
	}

	require.NoError(t, err)
	require.NotNil(t, resp)
	if validate != nil {
		validate(t, resp)
	}
}
