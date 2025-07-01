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

var examOnlyEndpoints = map[string]bool{
	"/proto.Paper/GetQuestionsByIds": true,
	// "/proto.Paper/GetCategoriesByIds": true,
	"/proto.Paper/GetExamQuestion":   true,
	"/proto.Paper/GetQuestionHashes": true,
	"/proto.Paper/GetQuestionIds":    true,
}

// ExamServiceAuthInterceptor validates that only the exam service can access certain endpoints
func ExamServiceAuthInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		// Only intercept exam service endpoints
		if !examOnlyEndpoints[info.FullMethod] {
			return handler(ctx, req)
		}

		// Get metadata from context
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Error(codes.Unauthenticated, "missing metadata")
		}

		// Get API token from metadata
		tokens := md.Get(constants.X_EXAM_API_TOKEN)
		if len(tokens) == 0 {
			return nil, status.Error(codes.Unauthenticated, "missing exam service API token")
		}

		// Validate API token
		if tokens[0] != env.EXAM_API_TOKEN {
			return nil, status.Error(codes.PermissionDenied, "invalid exam service API token")
		}

		return handler(ctx, req)
	}
}
