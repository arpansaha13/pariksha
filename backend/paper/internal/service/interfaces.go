package service

import (
	"context"

	"google.golang.org/protobuf/types/known/emptypb"

	"pariksha/common/pkg/proto"
	"pariksha/common/pkg/types"
)

// IPaperService defines the interface for paper service operations.
type IPaperService interface {
	GetUserPapers(ctx context.Context, userID types.UserID) (*proto.PaperList, error)
	CreatePaper(ctx context.Context, userID types.UserID) (*proto.CreatePaperResponse, error)
	GetPaper(ctx context.Context, paperHash string) (*proto.PaperResponse, error)
	UpdatePaper(ctx context.Context, req *proto.UpdatePaperRequest) error
	DeletePapers(ctx context.Context, hashes []string) error
}

// ICategoryService defines the interface for category service operations.
type ICategoryService interface {
	GetPaperCategories(ctx context.Context, paperHash string) (*proto.PaperCategoryList, error)
	CreateCategory(ctx context.Context, paperHash string) (*proto.PaperCategoryResponse, error)
	UpdateCategory(ctx context.Context, req *proto.UpdatePaperCategoryRequest) error
	ReorderCategories(ctx context.Context, paperHash string, categoryIDs []int64) error
	DeleteCategory(ctx context.Context, req *proto.PaperCategoryRequest) error
	GetPaperCategoriesMeta(ctx context.Context, req *proto.PaperRequest) (*proto.PaperCategoriesMeta, error)
}

// IQuestionService defines the interface for question service operations.
type IQuestionService interface {
	GetPaperQuestions(ctx context.Context, paperHash string) (*proto.QuestionList, error)
	GetPaperQuestion(ctx context.Context, req *proto.PaperQuestionRequest) (*proto.PaperQuestionResponse, error)
	CreatePaperQuestion(ctx context.Context, req *proto.CreatePaperQuestionRequest) (*proto.CreatePaperQuestionResponse, error)
	UpdatePaperQuestion(ctx context.Context, req *proto.UpdatePaperQuestionRequest) (*proto.UpdatePaperQuestionResponse, error)
	DeletePaperQuestion(ctx context.Context, paperHash string, questionHash string) error
	ReorderQuestions(ctx context.Context, categoryID int64, questionHashes []string) error
	GetBoilerplate(ctx context.Context, req *proto.GetPaperBoilerplateRequest) (*proto.BoilerplateResponse, error)
	UpsertTestCases(ctx context.Context, req *proto.UpsertPaperTestCasesRequest) (*emptypb.Empty, error)
	GetPaperQuestionsMeta(ctx context.Context, req *proto.PaperRequest) (*proto.PaperQuestionsMeta, error)
}
