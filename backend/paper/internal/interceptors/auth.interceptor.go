package interceptors

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"pariksha/common/pkg/models"
	"pariksha/common/pkg/proto"
	"pariksha/common/pkg/utils"
	"pariksha/paper/internal/config/db"

	"gorm.io/gorm"
)

type CategoryCtxKey struct{}
type QuestionCtxKey struct{}

var requiresRead = map[string]bool{
	"/proto.PaperService/GetPaper":           true,
	"/proto.PaperService/CheckPaperAccess":   true,
	"/proto.PaperService/GetPaperCategories": true,
	"/proto.PaperService/GetPaperQuestions":  true,
	"/proto.PaperService/GetPaperQuestion":   true,
}

var requiresWrite = map[string]bool{
	"/proto.PaperService/UpdatePaper":       true,
	"/proto.PaperService/CreateCategory":    true,
	"/proto.PaperService/UpdateCategory":    true,
	"/proto.PaperService/DeleteCategory":    true,
	"/proto.PaperService/ReorderCategories": true,
	"/proto.PaperService/UpdateQuestion":    true,
	"/proto.PaperService/DeleteQuestion":    true,
	"/proto.PaperService/CreateQuestion":    true,
}

func PaperAuthInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		// Check if the method needs to be intercepted
		methodName := info.FullMethod
		if !requiresRead[methodName] && !requiresWrite[methodName] {
			return handler(ctx, req)
		}

		// Get paper ID based on the request type and method
		var paperID int64

		switch r := req.(type) {
		case *proto.PaperRequest:
			paperID = r.PaperId
		case *proto.UpdatePaperRequest:
			paperID = r.PaperId
		case *proto.CreateCategoryRequest:
			paperID = r.PaperId
		case *proto.UpdateCategoryRequest:
			category, err := fetchCategoryData(r.CategoryId)
			if err != nil {
				return nil, err
			}

			paperID = category.PaperID.Int64
			ctx = context.WithValue(ctx, CategoryCtxKey{}, category)
		case *proto.CategoryRequest:
			category, err := fetchCategoryData(r.CategoryId)
			if err != nil {
				return nil, err
			}

			paperID = category.PaperID.Int64
			ctx = context.WithValue(ctx, CategoryCtxKey{}, category)
		case *proto.ReorderCategoriesRequest:
			paperID = r.PaperId
		case *proto.QuestionRequest:
			question, err := fetchQuestionData(r.QuestionId)
			if err != nil {
				return nil, err
			}
			paperID = question.PaperID.Int64
			ctx = context.WithValue(ctx, QuestionCtxKey{}, question)
		case *proto.UpdateQuestionRequest:
			question, err := fetchQuestionData(r.QuestionId)
			if err != nil {
				return nil, err
			}
			paperID = question.PaperID.Int64
			ctx = context.WithValue(ctx, QuestionCtxKey{}, question)
		case *proto.CreateQuestionRequest:
			paperID = r.PaperId
		}

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

		// Check if the method requires READ permission
		if requiresRead[methodName] {
			if !permissions.CanRead() {
				return nil, status.Error(codes.PermissionDenied, "READ permission required")
			}
		}

		// Check if the method requires WRITE permission
		if requiresWrite[methodName] {
			if !permissions.CanWrite() {
				return nil, status.Error(codes.PermissionDenied, "WRITE permission required")
			}
		}

		return handler(ctx, req)
	}
}

// Helper function to fetch category data
func fetchCategoryData(categoryID int64) (models.QuestionCategory, error) {
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
func fetchQuestionData(questionID int64) (models.Question, error) {
	var question models.Question
	if err := db.DB.Preload("Paper").Preload("Category").
		Take(&question, questionID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return question, status.Error(codes.NotFound, "question not found")
		}
		return question, status.Error(codes.Internal, "failed to find question")
	}
	return question, nil
}

// fetchPaperPermissions fetches the PaperPermissions entry for the given paperId and userId.
func fetchPaperPermissions(paperId int64, userId int64) (models.PaperPermissions, error) {
	var permissions models.PaperPermissions
	if err := db.DB.Where("paper_id = ? AND user_id = ?", paperId, userId).Take(&permissions).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return permissions, status.Error(codes.PermissionDenied, "No permission to access this paper")
		}
		return permissions, status.Error(codes.Internal, "failed to fetch permissions")
	}
	return permissions, nil
}
