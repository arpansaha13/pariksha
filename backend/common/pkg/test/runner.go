package test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// Runner executes a gRPC test case with common error handling
func Runner[Req any, Resp any](
	t *testing.T,
	tc TestCase[Req, Resp],
	call func(context.Context, Req, ...grpc.CallOption) (Resp, error),
) {
	var setupData map[string]any
	if tc.Setup != nil {
		setupData = tc.Setup(t)
	}

	ctx := metadata.NewOutgoingContext(context.Background(), tc.Metadata)
	resp, err := call(ctx, tc.GetRequest(setupData))

	if tc.ExpectedCode != codes.OK {
		assert.Equal(t, tc.ExpectedCode, status.Code(err))
		return
	}

	require.NoError(t, err)
	require.NotNil(t, resp)

	tc.Validate(t, resp, setupData)
}
