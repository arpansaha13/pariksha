package utils

import (
	"context"
	"pariksha/common/pkg/types"
	"strconv"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func GetUserIDFromMetadata(ctx context.Context) (types.UserID, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return 0, status.Error(codes.Unauthenticated, "missing metadata")
	}

	userIDs := md.Get("user_id")
	if len(userIDs) == 0 {
		return 0, status.Error(codes.Unauthenticated, "missing user id")
	}

	userID, err := strconv.ParseInt(userIDs[0], 10, 64)
	if err != nil || userID == 0 {
		return 0, status.Error(codes.InvalidArgument, "invalid user id")
	}

	return types.UserID(userID), nil
}
