package validate

import (
	"fmt"
	"strings"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"pariksha/common/pkg/constants"
	"pariksha/common/pkg/structs"
)

// McqQuestionData validates MCQ question data
func McqQuestionData(mcq *structs.MCQQuestion) error {
	if strings.TrimSpace(mcq.Statement) == "" {
		return status.Error(codes.InvalidArgument, "question statement cannot be empty")
	}
	if len(mcq.Options) < int(constants.MIN_MCQ_OPTIONS_COUNT) {
		return status.Error(codes.InvalidArgument, fmt.Sprintf("MCQ questions must have at least %d options", constants.MIN_MCQ_OPTIONS_COUNT))
	}
	if len(mcq.Options) > int(constants.MAX_MCQ_OPTIONS_COUNT) {
		return status.Error(codes.InvalidArgument, fmt.Sprintf("MCQ questions cannot have more than %d options", constants.MAX_MCQ_OPTIONS_COUNT))
	}
	return nil
}
