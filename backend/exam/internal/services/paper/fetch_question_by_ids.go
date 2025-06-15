package paper

import (
	"context"
	"fmt"

	"google.golang.org/grpc/metadata"

	"pariksha/common/pkg/constants"
	"pariksha/common/pkg/proto"
	"pariksha/exam/internal/config/env"
)

// FetchQuestionsByIds returns question details for the given question IDs.
// Returns error if any of the question IDs don't have corresponding questions.
var FetchQuestionsByIds = fetchQuestionsByIds

func fetchQuestionsByIds(questionIDs []int64) ([]*proto.QuestionBatchItem, error) {
	ensurePaperService()

	// Create metadata with exam token
	md := metadata.New(map[string]string{
		constants.X_EXAM_API_TOKEN: env.EXAM_API_TOKEN,
	})
	ctx := metadata.NewOutgoingContext(context.Background(), md)

	// Get questions
	resp, err := pSvc.client.GetQuestionsByIds(ctx, &proto.GetQuestionsByIdsRequest{
		QuestionIds: questionIDs,
	})
	if err != nil {
		return nil, err
	}

	// Verify we got all questions
	if len(resp.Questions) != len(questionIDs) {
		return nil, fmt.Errorf("failed to fetch all questions")
	}

	return resp.Questions, nil
}
