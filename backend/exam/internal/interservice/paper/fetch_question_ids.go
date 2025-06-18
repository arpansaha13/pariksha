package paper

import (
	"context"
	"fmt"

	"google.golang.org/grpc/metadata"

	"pariksha/common/pkg/constants"
	"pariksha/common/pkg/proto"
	"pariksha/exam/internal/config/env"
)

// FetchQuestionIdsForHashes returns question IDs for the given question hashes.
// Returns error if any of the question hashes don't have corresponding IDs.
var FetchQuestionIdsForHashes = fetchQuestionIdsForHashes

func fetchQuestionIdsForHashes(questionHashes []string) ([]int64, error) {
	ensurePaperService()

	// Create metadata with exam token
	md := metadata.New(map[string]string{
		constants.X_EXAM_API_TOKEN: env.EXAM_API_TOKEN,
	})
	ctx := metadata.NewOutgoingContext(context.Background(), md)

	// Get question IDs
	resp, err := pSvc.client.GetQuestionIds(ctx, &proto.GetQuestionIdsRequest{
		QuestionHashes: questionHashes,
	})
	if err != nil {
		return nil, err
	}

	// Verify we got all IDs
	if len(resp.QuestionIds) != len(questionHashes) {
		return nil, fmt.Errorf("failed to fetch all question IDs")
	}

	return resp.QuestionIds, nil
}
