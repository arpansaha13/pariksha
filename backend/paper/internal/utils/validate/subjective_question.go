package validate

import (
	"strings"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"pariksha/common/pkg/structs"
)

// SubjectiveQuestionData validates subjective question data
func SubjectiveQuestionData(subjective *structs.SubjectiveQuestion) error {
	if strings.TrimSpace(subjective.Statement) == "" {
		return status.Error(codes.InvalidArgument, "question statement cannot be empty")
	}
	return nil
}
