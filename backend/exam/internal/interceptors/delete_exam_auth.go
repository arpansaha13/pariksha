package interceptors

import (
	"context"
	"fmt"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"pariksha/common/pkg/models"
	"pariksha/common/pkg/proto"
	"pariksha/common/pkg/utils"
	"pariksha/exam/internal/config/db"
)

var deleteExamShouldIntercept = map[string]bool{
	"/proto.Exam/DeleteExams": true,
}

// DeleteExamsAuthInterceptor checks if user has WRITE permissions for all exams being deleted
func DeleteExamsAuthInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		methodName := info.FullMethod
		if !deleteExamShouldIntercept[methodName] {
			return handler(ctx, req)
		}

		deleteReq, ok := req.(*proto.DeleteExamsRequest)
		if !ok {
			return nil, status.Error(codes.Internal, "invalid request type")
		}

		if len(deleteReq.ExamHashes) == 0 {
			return nil, status.Error(codes.InvalidArgument, "exam hashes list cannot be empty")
		}

		userID, err := utils.GetUserIDFromMetadata(ctx)
		if err != nil {
			return nil, err
		}

		// Get all permissions for these exams for this user
		var permissions []models.ExamPermission
		if err := db.DB.Joins("INNER JOIN exams ON exams.id = permissions.exam_id").
			Where("exams.hash IN ? AND permissions.user_id = ?", deleteReq.ExamHashes, userID).
			Find(&permissions).Error; err != nil {
			return nil, status.Error(codes.Internal, fmt.Sprintf("failed to fetch permissions: %s", err.Error()))
		}

		// Non-existent exams should return success
		// Because anyway there were supposed to be deleted

		// If a permission exists, it must have WRITE access
		for _, perm := range permissions {
			if !perm.CanWrite() {
				return nil, status.Error(codes.PermissionDenied, "no permission to delete one or more exams")
			}
		}

		return handler(ctx, req)
	}
}
