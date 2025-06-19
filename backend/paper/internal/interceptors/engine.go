package interceptors

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"pariksha/common/pkg/constants"
	"pariksha/paper/internal/config/env"
)

var engineOnlyEndpoints = map[string]bool{}

// EngineServiceAuthInterceptor validates that only the engine service can access certain endpoints
func EngineServiceAuthInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		// Only intercept engine service endpoints
		if !engineOnlyEndpoints[info.FullMethod] {
			return handler(ctx, req)
		}

		// Get metadata from context
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Error(codes.Unauthenticated, "missing metadata")
		}

		// Get API token from metadata
		tokens := md.Get(constants.X_ENGINE_API_TOKEN)
		if len(tokens) == 0 {
			return nil, status.Error(codes.Unauthenticated, "missing engine service API token")
		}

		// Validate API token
		if tokens[0] != env.ENGINE_API_TOKEN {
			return nil, status.Error(codes.PermissionDenied, "invalid engine service API token")
		}

		return handler(ctx, req)
	}
}
