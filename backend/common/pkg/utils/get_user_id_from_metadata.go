package utils

import (
	"context"
	"strconv"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func GetUserIDFromMetadata(ctx context.Context) (int64, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return 0, status.Error(codes.Unauthenticated, "missing metadata")
	}

	userIDs := md.Get("user_id")
	if len(userIDs) == 0 {
		return 0, status.Error(codes.Unauthenticated, "missing user id")
	}

	userID, err := strconv.Atoi(userIDs[0])
	if err != nil || userID == 0 {
		return 0, status.Error(codes.InvalidArgument, "invalid user id")
	}

	return int64(userID), nil
}
