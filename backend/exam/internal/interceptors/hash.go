package interceptors

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"pariksha/common/pkg/models"
	"pariksha/common/pkg/proto"
	"pariksha/common/pkg/types"
	"pariksha/common/pkg/utils"
	"pariksha/exam/internal/config/db"
)

type hashContextKey string

const (
	examIDContextKey  hashContextKey = "examID"
	examIDsContextKey hashContextKey = "examIDs"
)

// getExamHashFromRequest extracts ExamHash field from requests
func getExamHashFromRequest(req any) (string, bool) {
	switch r := req.(type) {
	case *proto.ExamRequest:
		return r.ExamHash, true
	case *proto.UpdateExamRequest:
		return r.ExamHash, true
	case *proto.StartExamRequest:
		return r.ExamHash, true
	case *proto.EndExamRequest:
		return r.ExamHash, true
	case *proto.AddParticipantRequest:
		return r.ExamHash, true
	case *proto.RemoveParticipantRequest:
		return r.ExamHash, true
	case *proto.UpsertAnswersRequest:
		return r.ExamHash, true
	case *proto.GetAnswerRequest:
		return r.ExamHash, true
	case *proto.CheckParticipantRequest:
		return r.ExamHash, true
	case *proto.GetExamParticipantRequest:
		return r.ExamHash, true
	default:
		return "", false
	}
}

// getExamHashesFromRequest extracts array of exam hashes from batch requests
func getExamHashesFromRequest(req any) ([]string, bool) {
	switch r := req.(type) {
	case *proto.DeleteExamsRequest:
		return r.ExamHashes, true
	default:
		return nil, false
	}
}

// SingleExamHashInterceptor converts exam hash to ID
func SingleExamHashInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, _ *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		hash, ok := getExamHashFromRequest(req)
		if !ok {
			return handler(ctx, req)
		}

		examID, err := utils.GetIDFromHash(db.DB, hash, models.ExamHash{}.TableName())
		if err != nil {
			return nil, status.Error(codes.NotFound, "exam not found")
		}

		ctx = context.WithValue(ctx, examIDContextKey, types.ExamID(examID))
		return handler(ctx, req)
	}
}

// BatchExamHashInterceptor converts array of exam_hashes to IDs maintaining sequence
func BatchExamHashInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, _ *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		hashes, ok := getExamHashesFromRequest(req)
		if !ok {
			return handler(ctx, req)
		}

		hashToID, err := utils.GetIDsFromHashes(db.DB, hashes, models.ExamHash{}.TableName())
		if err != nil {
			return nil, status.Error(codes.Internal, "failed to fetch exam IDs")
		}

		// Create response maintaining input order
		typedExamIDs := make([]types.ExamID, 0, len(hashes))
		for _, hash := range hashes {
			if id, exists := hashToID[hash]; exists {
				typedExamIDs = append(typedExamIDs, types.ExamID(id))
			}
		}

		ctx = context.WithValue(ctx, examIDsContextKey, typedExamIDs)
		return handler(ctx, req)
	}
}

// GetExamIDFromContext retrieves the exam ID from context
func GetExamIDFromContext(ctx context.Context) (types.ExamID, bool) {
	id, ok := ctx.Value(examIDContextKey).(types.ExamID)
	return id, ok
}

// GetExamIDsFromContext retrieves the exam IDs array from context
func GetExamIDsFromContext(ctx context.Context) ([]types.ExamID, bool) {
	ids, ok := ctx.Value(examIDsContextKey).([]types.ExamID)
	return ids, ok
}
