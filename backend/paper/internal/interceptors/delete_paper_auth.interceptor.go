package interceptors

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"pariksha/common/pkg/models"
	"pariksha/common/pkg/proto"
	"pariksha/common/pkg/utils"
	"pariksha/paper/internal/config/db"
)

const deletePaperPath = "/proto.Paper/DeletePapers"

// DeletePaperAuthInterceptor returns a new unary server interceptor that handles
// permission checks for the DeletePapers endpoint.
func DeletePaperAuthInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		// Only intercept DeletePapers requests
		if info.FullMethod != deletePaperPath {
			return handler(ctx, req)
		}

		// Get user ID from context
		userID, err := utils.GetUserIDFromMetadata(ctx)
		if err != nil {
			return nil, err
		}

		// Get paper IDs from request
		deleteReq, ok := req.(*proto.DeletePapersRequest)
		if !ok {
			return nil, status.Error(codes.Internal, "invalid request type")
		}

		if len(deleteReq.PaperHashes) == 0 {
			return nil, status.Error(codes.InvalidArgument, "paper_ids cannot be empty")
		}

		// Take paperIDs from context added by hash interceptor
		paperIDs, ok := GetPaperIDsFromContext(ctx)
		if !ok {
			return nil, status.Error(codes.Internal, "paper_ids not found in context")
		}

		// Get permissions for all papers
		var permissions []models.PaperPermission
		if err := db.DB.Where("paper_id IN ? AND user_id = ?", paperIDs, userID).Find(&permissions).Error; err != nil {
			return nil, status.Error(codes.Internal, "failed to fetch permissions")
		}

		// Check WRITE permission for papers
		for _, perm := range permissions {
			if !perm.CanWrite() {
				return nil, status.Error(codes.PermissionDenied, "WRITE permission required for all papers")
			}
		}

		return handler(ctx, req)
	}
}
