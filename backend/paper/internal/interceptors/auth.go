package interceptors

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"

	"pariksha/common/pkg/models"
	"pariksha/common/pkg/proto"
	"pariksha/common/pkg/types"
	"pariksha/common/pkg/utils"
	"pariksha/common/pkg/utils/grpcerror"
	"pariksha/paper/internal/config/db"
)

type paperContextKey string

const (
	questionContextKey   paperContextKey = "question"
	categoryContextKey   paperContextKey = "category"
	permissionContextKey paperContextKey = "permission"
)

var requiresRead = map[string]bool{
	"/proto.Paper/GetPaper":            true,
	"/proto.Paper/GetPaperCategories":  true,
	"/proto.Paper/GetPaperQuestions":   true,
	"/proto.Paper/GetPaperQuestion":    true,
	"/proto.Paper/GetPaperPermissions": true,
}

var requiresWrite = map[string]bool{
	"/proto.Paper/UpdatePaper":          true,
	"/proto.Paper/CreateCategory":       true,
	"/proto.Paper/UpdateCategory":       true,
	"/proto.Paper/DeleteCategory":       true,
	"/proto.Paper/ReorderCategories":    true,
	"/proto.Paper/UpdateQuestion":       true,
	"/proto.Paper/DeleteQuestion":       true,
	"/proto.Paper/CreateQuestion":       true,
	"/proto.Paper/UpsertPaperTestCases": true,
}

func PaperAuthInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		methodName := info.FullMethod
		if !requiresRead[methodName] && !requiresWrite[methodName] {
			return handler(ctx, req)
		}

		// Try to get paper ID from paper hash in request
		if paperHash := getPaperHashFromRequest(req); paperHash != "" {
			var paperID int64
			if err := db.DB.Model(&models.Paper{}).
				Where("hash = ?", paperHash).
				Pluck("id", &paperID).Error; err != nil {
				if err == gorm.ErrRecordNotFound {
					return nil, status.Error(codes.NotFound, "paper not found")
				}
				return nil, grpcerror.Internal(err, "failed to fetch paper")
			}
			return handlePaperAuth(ctx, methodName, types.PaperID(paperID), handler, req)
		}

		// Try to get question from question hash in request
		if questionHash := getQuestionHashFromRequest(req); questionHash != "" {
			var question models.Question
			if err := db.DB.Where("hash = ?", questionHash).Take(&question).Error; err != nil {
				if err == gorm.ErrRecordNotFound {
					return nil, status.Error(codes.NotFound, "question not found")
				}
				return nil, grpcerror.Internal(err, "failed to fetch question")
			}
			ctx = context.WithValue(ctx, questionContextKey, question)
			return handlePaperAuth(ctx, methodName, types.PaperID(question.PaperID.Int64), handler, req)
		}

		var paperID int64
		switch r := req.(type) {
		case *proto.UpdateCategoryRequest:
			category, err := fetchCategoryData(types.CategoryID(r.CategoryId))
			if err != nil {
				return nil, err
			}
			paperID = category.PaperID.Int64
			ctx = context.WithValue(ctx, categoryContextKey, category)
		case *proto.CategoryRequest:
			category, err := fetchCategoryData(types.CategoryID(r.CategoryId))
			if err != nil {
				return nil, err
			}
			paperID = category.PaperID.Int64
			ctx = context.WithValue(ctx, categoryContextKey, category)
		default:
			return nil, status.Error(codes.Internal, "unable to determine paper ID")
		}

		return handlePaperAuth(ctx, methodName, types.PaperID(paperID), handler, req)
	}
}

// handlePaperAuth handles the common paper authorization logic
func handlePaperAuth(ctx context.Context, methodName string, paperID types.PaperID, handler grpc.UnaryHandler, req any) (any, error) {
	userID, err := utils.GetUserIDFromMetadata(ctx)
	if err != nil {
		return nil, err
	}

	// Check if paper exists
	var exists bool
	if err := db.DB.Model(&models.Paper{}).
		Select("1").
		Where("id = ?", paperID).
		Find(&exists).Error; err != nil {
		return nil, status.Error(codes.Internal, "failed to check paper existence")
	}
	if !exists {
		return nil, status.Error(codes.NotFound, "paper not found")
	}

	permissions, err := fetchPaperPermissions(paperID, userID)
	if err != nil {
		return nil, err
	}

	ctx = context.WithValue(ctx, permissionContextKey, permissions)

	// Check if the method requires READ permission
	if requiresRead[methodName] && !permissions.CanRead() {
		return nil, status.Error(codes.PermissionDenied, "READ permission required")
	}

	// Check if the method requires WRITE permission
	if requiresWrite[methodName] && !permissions.CanWrite() {
		return nil, status.Error(codes.PermissionDenied, "WRITE permission required")
	}

	return handler(ctx, req)
}

// Helper function to fetch category data
func fetchCategoryData(categoryID types.CategoryID) (models.QuestionCategory, error) {
	var category models.QuestionCategory
	if err := db.DB.Take(&category, categoryID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return category, status.Error(codes.NotFound, "category not found")
		}
		return category, status.Error(codes.Internal, "failed to find category")
	}
	return category, nil
}

// Helper function to fetch question data
func fetchQuestionData(questionID types.QuestionID) (models.Question, error) {
	var question models.Question
	if err := db.DB.
		Take(&question, questionID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return question, status.Error(codes.NotFound, "question not found")
		}
		return question, status.Error(codes.Internal, "failed to find question")
	}
	return question, nil
}

// fetchPaperPermissions fetches the PaperPermission entry for the given paperId and userId.
func fetchPaperPermissions(paperId types.PaperID, userId types.UserID) (models.PaperPermission, error) {
	var permissions models.PaperPermission
	if err := db.DB.Where("paper_id = ? AND user_id = ?", paperId, userId).Take(&permissions).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return permissions, status.Error(codes.PermissionDenied, "No permission to access this paper")
		}
		return permissions, status.Error(codes.Internal, "failed to fetch permissions")
	}
	return permissions, nil
}

// Getter function to safely access question from context
func GetQuestionFromContext(ctx context.Context) (*models.Question, error) {
	question, ok := ctx.Value(questionContextKey).(models.Question)
	if !ok {
		return nil, status.Error(codes.Internal, "question data not found in context")
	}
	return &question, nil
}

// Getter function to safely access category from context
func GetCategoryFromContext(ctx context.Context) (*models.QuestionCategory, error) {
	category, ok := ctx.Value(categoryContextKey).(models.QuestionCategory)
	if !ok {
		return nil, status.Error(codes.Internal, "category data not found in context")
	}
	return &category, nil
}

// Getter function to safely access permission from context
func GetPermissionFromContext(ctx context.Context) (*models.PaperPermission, error) {
	permission, ok := ctx.Value(permissionContextKey).(models.PaperPermission)
	if !ok {
		return nil, status.Error(codes.Internal, "paper permission data not found in context")
	}
	return &permission, nil
}

// getPaperHashFromRequest extracts paper_hash from supported request types
func getPaperHashFromRequest(req any) string {
	switch r := req.(type) {
	case *proto.PaperRequest:
		return r.PaperHash
	case *proto.UpdatePaperRequest:
		return r.PaperHash
	case *proto.CreateQuestionRequest:
		return r.PaperHash
	case *proto.CreateCategoryRequest:
		return r.PaperHash
	case *proto.ReorderCategoriesRequest:
		return r.PaperHash
	default:
		return ""
	}
}

// getQuestionHashFromRequest extracts question_hash from supported request types
func getQuestionHashFromRequest(req any) string {
	switch r := req.(type) {
	case *proto.QuestionRequest:
		return r.QuestionHash
	case *proto.UpdateQuestionRequest:
		return r.QuestionHash
	case *proto.UpsertTestCasesRequest:
		return r.QuestionHash
	case *proto.GetBoilerplateRequest:
		return r.QuestionHash
	default:
		return ""
	}
}
