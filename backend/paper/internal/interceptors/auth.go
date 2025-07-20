package interceptors

import (
	"context"
	"fmt"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"

	"pariksha/common/pkg/utils"
	"pariksha/paper/internal/models"
	"pariksha/paper/internal/repositories"
)

type paperContextKey string

const (
	questionContextKey   paperContextKey = "question"
	categoryContextKey   paperContextKey = "category"
	permissionContextKey paperContextKey = "permission"
)

var requiresRead = map[string]struct{}{
	"/proto.Paper/GetPaper":            {},
	"/proto.Paper/GetPaperCategories":  {},
	"/proto.Paper/GetPaperQuestions":   {},
	"/proto.Paper/GetPaperQuestion":    {},
	"/proto.Paper/GetPaperPermissions": {},
}

var requiresWrite = map[string]struct{}{
	"/proto.Paper/UpdatePaper":            {},
	"/proto.Paper/CreatePaperCategory":    {},
	"/proto.Paper/UpdatePaperCategory":    {},
	"/proto.Paper/DeletePaperCategory":    {},
	"/proto.Paper/ReorderPaperCategories": {},
	"/proto.Paper/UpdatePaperQuestion":    {},
	"/proto.Paper/DeletePaperQuestion":    {},
	"/proto.Paper/CreatePaperQuestion":    {},
	"/proto.Paper/UpsertPaperTestCases":   {},
}

func PaperAuthInterceptor(permissionRepo *repositories.PaperPermission) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		methodName := info.FullMethod
		_, needsRead := requiresRead[methodName]
		_, needsWrite := requiresWrite[methodName]
		if !needsRead && !needsWrite {
			return handler(ctx, req)
		}

		// Try to get paper ID from paper hash in request
		paperHash, err := getPaperHashFromRequest(req)
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		if paperHash == "" {
			return nil, status.Error(codes.PermissionDenied, "no permission for this resource")
		}

		userID, err := utils.GetUserIDFromMetadata(ctx)
		if err != nil {
			return nil, err
		}

		permissions, err := permissionRepo.GetByPaperHashAndUserId(nil, paperHash, userID)
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				return permissions, status.Error(codes.PermissionDenied, "No permission to access this paper")
			}
			return permissions, status.Error(codes.Internal, "failed to fetch permissions")
		}

		ctx = context.WithValue(ctx, permissionContextKey, *permissions)

		// Check if the method requires READ permission
		if _, ok := requiresRead[methodName]; ok && !permissions.CanRead() {
			return nil, status.Error(codes.PermissionDenied, "READ permission required")
		}

		// Check if the method requires WRITE permission
		if _, ok := requiresWrite[methodName]; ok && !permissions.CanWrite() {
			return nil, status.Error(codes.PermissionDenied, "WRITE permission required")
		}

		return handler(ctx, req)
	}
}

// Getter function to safely access permission from context
func GetPermissionFromContext(ctx context.Context) (*models.PaperPermission, error) {
	permission, ok := ctx.Value(permissionContextKey).(models.PaperPermission)
	if !ok {
		return nil, status.Error(codes.Internal, "paper permission data not found in context")
	}
	return &permission, nil
}

type RequestWithPaperHash interface {
	GetPaperHash() string
}

func getPaperHashFromRequest(req any) (string, error) {
	if r, ok := req.(RequestWithPaperHash); ok {
		return r.GetPaperHash(), nil
	}
	return "", fmt.Errorf("invalid request type: does not implement RequestWithPaperHash")
}
