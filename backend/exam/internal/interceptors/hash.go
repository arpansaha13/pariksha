package interceptors

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"pariksha/common/pkg/proto"
	"pariksha/common/pkg/types"
	"pariksha/exam/internal/interservice"
)

type hashContextKey string

const (
	questionIDContextKey hashContextKey = "questionID"
)

// getQuestionHashFromRequest extracts QuestionHash field from requests
func getQuestionHashFromRequest(req any) (string, bool) {
	switch r := req.(type) {
	case *proto.GetAnswerRequest:
		return r.QuestionHash, true
	case *proto.ParticipantQuestionRequest:
		return r.QuestionHash, true
	case *proto.UpsertAnswersRequest:
		return r.Answer.QuestionHash, true
	default:
		return "", false
	}
}

// SingleQuestionHashInterceptor converts question hash to ID
func SingleQuestionHashInterceptor(questionIntSvc *interservice.Question) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, _ *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		hash, ok := getQuestionHashFromRequest(req)
		if !ok {
			return handler(ctx, req)
		}

		// Get question ID using utility function
		questionIDs, err := questionIntSvc.GetQuestionIDsByHashes([]string{hash})
		if err != nil {
			return nil, status.Error(codes.Internal, "failed to fetch question ID")
		}

		if len(questionIDs) > 0 {
			// Store question ID in context
			ctx = context.WithValue(ctx, questionIDContextKey, types.QuestionID(questionIDs[0]))
		}

		return handler(ctx, req)
	}
}

// GetQuestionIDFromContext retrieves the question ID from context
func GetQuestionIDFromContext(ctx context.Context) (types.QuestionID, bool) {
	id, ok := ctx.Value(questionIDContextKey).(types.QuestionID)
	return id, ok
}
