package paper

import (
	"context"

	"google.golang.org/grpc/metadata"

	"pariksha/common/pkg/constants"
	"pariksha/common/pkg/proto"
	"pariksha/exam/internal/config/env"
)

// FetchQuestionByHash returns question details for the given question hash.
var FetchQuestionByHash = fetchQuestionByHash

func fetchQuestionByHash(questionHash string) (*proto.ExamQuestionResponse, error) {
	ensurePaperService()

	// Create metadata with exam token
	md := metadata.New(map[string]string{
		constants.X_EXAM_API_TOKEN: env.EXAM_API_TOKEN,
	})
	ctx := metadata.NewOutgoingContext(context.Background(), md)

	// Get questions
	resp, err := pSvc.client.GetExamQuestionByHash(ctx, &proto.QuestionRequest{
		QuestionHash: questionHash,
	})
	if err != nil {
		return nil, err
	}

	return resp, nil
}
