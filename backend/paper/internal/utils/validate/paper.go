package validate

import (
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"pariksha/common/pkg/constants"
)

// MaxScore checks if the given score is within valid range (0 to MAX_SCORE_PER_QUESTION)
func MaxScore(score int32) error {
	if score < 0 || score > constants.MAX_SCORE_PER_QUESTION {
		return status.Errorf(codes.InvalidArgument, "max score must be between 0 and %d", constants.MAX_SCORE_PER_QUESTION)
	}
	return nil
}

// PaperDuration checks if the given duration in minutes is within valid range
func PaperDuration(durationMinutes int32) error {
	if durationMinutes < 0 {
		return status.Error(codes.InvalidArgument, "duration must be positive")
	}
	if durationMinutes > int32(constants.MAX_EXAM_DURATION_MINUTES) {
		return status.Error(codes.InvalidArgument, "duration cannot exceed 24 hours")
	}
	return nil
}
