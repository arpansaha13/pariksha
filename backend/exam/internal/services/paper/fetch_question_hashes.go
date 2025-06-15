package paper

import (
	"context"
	"fmt"

	"google.golang.org/grpc/metadata"

	"pariksha/common/pkg/constants"
	"pariksha/common/pkg/proto"
	"pariksha/exam/internal/config/env"
)

// FetchQuestionHashesForIds returns question hashes for the given question IDs.
//
// Returns error if any of the question IDs don't have corresponding hashes.
var FetchQuestionHashesForIds = fetchQuestionHashesForIds

func fetchQuestionHashesForIds(questionIDs []int64) ([]string, error) {
	ensurePaperService()

	// Create metadata with exam token
	md := metadata.New(map[string]string{
		constants.X_EXAM_API_TOKEN: env.EXAM_API_TOKEN,
	})
	ctx := metadata.NewOutgoingContext(context.Background(), md)

	// Get question hashes
	resp, err := pSvc.client.GetQuestionHashes(ctx, &proto.GetQuestionHashesRequest{
		QuestionIds: questionIDs,
	})
	if err != nil {
		return nil, err
	}

	// Verify we got all hashes
	if len(resp.QuestionHashes) != len(questionIDs) {
		return nil, fmt.Errorf("failed to fetch all question hashes")
	}

	return resp.QuestionHashes, nil
}
