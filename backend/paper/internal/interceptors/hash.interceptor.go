package interceptors

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"pariksha/common/pkg/models"
	"pariksha/common/pkg/proto"
	"pariksha/common/pkg/types"
	"pariksha/paper/internal/config/db"
)

type hashContextKey string

const (
	paperIDKey     hashContextKey = "paperIdFromHash"
	questionIDKey  hashContextKey = "questionIdFromHash"
	paperIDsKey    hashContextKey = "paperIdsFromHashes"
	questionIDsKey hashContextKey = "questionIdsFromHashes"
)

// getIDFromHash fetches entity ID for a given hash from the database
func getIDFromHash(hash string, table string) (int64, error) {
	var id int64
	err := db.DB.Table(table).
		Select("id").
		Where("hash = ?", hash).
		Take(&id).Error
	return id, err
}

// getIDsFromHashes fetches entity IDs for given hashes from the database
func getIDsFromHashes(hashes []string, table string) ([]int64, error) {
	var ids []int64
	err := db.DB.Table(table).
		Select("id").
		Where("hash IN ?", hashes).
		Find(&ids).Error
	return ids, err
}

// getPaperHashFromRequest extracts PaperHash field from requests
func getPaperHashFromRequest(req any) (string, bool) {
	switch r := req.(type) {
	case *proto.PaperRequest:
		return r.PaperHash, true
	case *proto.UpdatePaperRequest:
		return r.PaperHash, true
	case *proto.CreateQuestionRequest:
		return r.PaperHash, true
	case *proto.CreateCategoryRequest:
		return r.PaperHash, true
	case *proto.ReorderCategoriesRequest:
		return r.PaperHash, true
	default:
		return "", false
	}
}

// getQuestionHashFromRequest extracts QuestionHash field from requests
func getQuestionHashFromRequest(req any) (string, bool) {
	switch r := req.(type) {
	case *proto.QuestionRequest:
		return r.QuestionHash, true
	case *proto.UpdateQuestionRequest:
		return r.QuestionHash, true
	case *proto.GetBoilerplateRequest:
		return r.QuestionHash, true
	case *proto.UpsertTestCasesRequest:
		return r.QuestionHash, true
	default:
		return "", false
	}
}

// getPaperHashesFromRequest extracts hash array from batch requests
func getPaperHashesFromRequest(req any) ([]string, bool) {
	switch r := req.(type) {
	case *proto.DeletePapersRequest:
		return r.PaperHashes, true
	default:
		return nil, false
	}
}

// getQuestionHashesFromRequest extracts hash array from batch requests
func getQuestionHashesFromRequest(req any) ([]string, bool) {
	switch r := req.(type) {
	case *proto.ReorderQuestionsRequest:
		return r.QuestionHashes, true
	case *proto.GetQuestionIdsRequest:
		return r.QuestionHashes, true
	default:
		return nil, false
	}
}

// SinglePaperHashInterceptor converts a single paper_hash to ID
func SinglePaperHashInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, _ *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		hash, ok := getPaperHashFromRequest(req)
		if !ok {
			return handler(ctx, req)
		}

		paperID, err := getIDFromHash(hash, models.PaperHash{}.TableName())
		if err != nil {
			return nil, status.Error(codes.NotFound, "paper not found")
		}

		// Store paper ID in context
		ctx = context.WithValue(ctx, paperIDKey, types.PaperID(paperID))
		return handler(ctx, req)
	}
}

// SingleQuestionHashInterceptor converts a single question_hash to ID
func SingleQuestionHashInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, _ *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		hash, ok := getQuestionHashFromRequest(req)
		if !ok {
			return handler(ctx, req)
		}

		questionID, err := getIDFromHash(hash, models.QuestionHash{}.TableName())
		if err != nil {
			return nil, status.Error(codes.NotFound, "question not found")
		}

		// Store question ID in context
		ctx = context.WithValue(ctx, questionIDKey, types.QuestionID(questionID))
		return handler(ctx, req)
	}
}

// BatchPaperHashInterceptor converts array of paper_hashes to IDs
func BatchPaperHashInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, _ *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		hashes, ok := getPaperHashesFromRequest(req)
		if !ok {
			return handler(ctx, req)
		}

		// Get all hashes, store in map for lookup
		var paperHashModels []models.PaperHash
		if err := db.DB.Where("hash IN ?", hashes).Find(&paperHashModels).Error; err != nil {
			return nil, status.Error(codes.Internal, "failed to fetch paper IDs")
		}
		hashToID := make(map[string]types.PaperID)
		for _, ph := range paperHashModels {
			hashToID[ph.Hash] = ph.ID
		}

		// Create response maintaining input order
		typedPaperIDs := make([]types.PaperID, 0, len(hashes))
		for _, hash := range hashes {
			if id, exists := hashToID[hash]; exists {
				typedPaperIDs = append(typedPaperIDs, id)
			}
		}

		ctx = context.WithValue(ctx, paperIDsKey, typedPaperIDs)
		return handler(ctx, req)
	}
}

// BatchQuestionHashInterceptor converts array of question_hashes to IDs
func BatchQuestionHashInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, _ *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		hashes, ok := getQuestionHashesFromRequest(req)
		if !ok {
			return handler(ctx, req)
		}

		// Get all hashes, store in map for lookup
		var questionHashModels []models.QuestionHash
		if err := db.DB.Where("hash IN ?", hashes).Find(&questionHashModels).Error; err != nil {
			return nil, status.Error(codes.Internal, "failed to fetch question IDs")
		}
		hashToID := make(map[string]types.QuestionID)
		for _, qh := range questionHashModels {
			hashToID[qh.Hash] = qh.ID
		}

		// Create response maintaining input order
		typedQuestionIDs := make([]types.QuestionID, 0, len(hashes))
		for _, hash := range hashes {
			if id, exists := hashToID[hash]; exists {
				typedQuestionIDs = append(typedQuestionIDs, id)
			}
		}

		ctx = context.WithValue(ctx, questionIDsKey, typedQuestionIDs)
		return handler(ctx, req)
	}
}

// GetPaperIDFromContext retrieves the paper ID from context
func GetPaperIDFromContext(ctx context.Context) (types.PaperID, bool) {
	id, ok := ctx.Value(paperIDKey).(types.PaperID)
	return id, ok
}

// GetPaperIDsFromContext retrieves the paper IDs array from context
func GetPaperIDsFromContext(ctx context.Context) ([]types.PaperID, bool) {
	ids, ok := ctx.Value(paperIDsKey).([]types.PaperID)
	return ids, ok
}

// GetQuestionIDFromContext retrieves the question ID from context
func GetQuestionIDFromContext(ctx context.Context) (types.QuestionID, bool) {
	id, ok := ctx.Value(questionIDKey).(types.QuestionID)
	return id, ok
}

// GetQuestionIDsFromContext retrieves the question IDs array from context
func GetQuestionIDsFromContext(ctx context.Context) ([]types.QuestionID, bool) {
	ids, ok := ctx.Value(questionIDsKey).([]types.QuestionID)
	return ids, ok
}
