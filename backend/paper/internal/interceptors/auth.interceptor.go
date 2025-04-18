package interceptors

import (
	"context"
	"database/sql"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"pariksha/common/pkg/constants"
	"pariksha/common/pkg/models"
	"pariksha/common/pkg/proto"
	"pariksha/common/pkg/utils"
	"pariksha/paper/internal/config/db"

	"gorm.io/gorm"
)

type CategoryCtxKey struct{}
type QuestionCtxKey struct{}

func PaperAccessInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		// Only intercept specific methods
		methodName := info.FullMethod
		if !shouldIntercept(methodName) {
			return handler(ctx, req)
		}

		userID, err := utils.GetUserIDFromMetadata(ctx)
		if err != nil {
			return nil, err
		}

		// Get paper ID and required ownership type based on the request type and method
		var paperID int64
		ownershipType := constants.PAPER_OWNERSHIP_TYPE_OWNER

		switch r := req.(type) {
		case *proto.PaperRequest:
			paperID = r.PaperId

			if methodName != "/proto.PaperService/CheckPaperAccess" {
				ownershipType = ""
			}
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
			ownershipType = ""
			ctx = context.WithValue(ctx, QuestionCtxKey{}, question)
		case *proto.UpdateQuestionRequest:
			question, err := fetchQuestionData(r.QuestionId)
			if err != nil {
				return nil, err
			}
			paperID = question.PaperID.Int64
			ownershipType = constants.PAPER_OWNERSHIP_TYPE_OWNER
			ctx = context.WithValue(ctx, QuestionCtxKey{}, question)
		case *proto.CreateQuestionRequest:
			paperID = r.PaperId
			ownershipType = constants.PAPER_OWNERSHIP_TYPE_OWNER
		}

		// Verify paper access with ownership type check when required
		if err := verifyPaperAccess(nil, paperID, userID, ownershipType); err != nil {
			return nil, err
		}

		return handler(ctx, req)
	}
}

func shouldIntercept(methodName string) bool {
	methodsToIntercept := []string{
		"/proto.PaperService/GetPaper",
		"/proto.PaperService/UpdatePaper",
		"/proto.PaperService/CheckPaperAccess",
		"/proto.PaperService/GetPaperCategories",
		"/proto.PaperService/CreateCategory",
		"/proto.PaperService/UpdateCategory",
		"/proto.PaperService/DeleteCategory",
		"/proto.PaperService/ReorderCategories",
		"/proto.PaperService/GetPaperQuestions",
		"/proto.PaperService/GetQuestion",
		"/proto.PaperService/UpdateQuestion",
		"/proto.PaperService/DeleteQuestion",
		"/proto.PaperService/CreateQuestion",
		// "/proto.PaperService/TestGetQuestionsByIds",
	}

	for _, method := range methodsToIntercept {
		if strings.HasSuffix(methodName, method) {
			return true
		}
	}
	return false
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

// Helper function to verify paper access
func verifyPaperAccess(tx *gorm.DB, paperID any, userID int64, ownershipType string) error {
	if tx == nil {
		tx = db.DB
	}

	var actualPaperID int64
	switch v := paperID.(type) {
	case sql.NullInt64:
		if !v.Valid {
			return status.Error(codes.InvalidArgument, "invalid paper id")
		}
		actualPaperID = v.Int64
	case int64:
		actualPaperID = v
	default:
		return status.Error(codes.InvalidArgument, "invalid paper id type")
	}

	// Check if paper exists first
	var exists bool
	err := tx.Raw(`SELECT EXISTS (SELECT 1 FROM papers WHERE id = ?)`, actualPaperID).
		Scan(&exists).Error
	if err != nil {
		return status.Error(codes.Internal, "failed to check paper existence")
	}
	if !exists {
		return status.Error(codes.NotFound, "paper not found")
	}

	// Check paper access
	var condition string
	var args []any

	args = append(args, actualPaperID, userID)
	condition = "po.paper_id = ? AND po.user_id = ?"

	if ownershipType != "" {
		condition += " AND po.type = ?"
		args = append(args, ownershipType)
	}

	err = tx.Raw(`SELECT EXISTS (
			SELECT 1 FROM paper_ownerships po
			WHERE `+condition+`)`, args...).
		Scan(&exists).Error

	if err != nil {
		return status.Error(codes.Internal, "failed to check paper access")
	}

	if !exists {
		return status.Error(codes.PermissionDenied, "no permission to perform this action")
	}

	return nil
}
