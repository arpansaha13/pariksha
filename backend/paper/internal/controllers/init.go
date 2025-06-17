package controllers

import (
	"context"

	"pariksha/common/pkg/constants"
	"pariksha/common/pkg/proto"
	"pariksha/paper/internal/config/db"
	"pariksha/paper/internal/config/env"
	"pariksha/paper/internal/repositories"
	"pariksha/paper/internal/services"
)

type PaperServer struct {
	proto.UnimplementedPaperServer
}

func init() {
	if env.GO_ENV != constants.GO_ENV_TEST {
		InitializeHandlers()
	}
}

var (
	paperCtrl       *Paper
	categoryCtrl    *Category
	questionCtrl    *Question
	boilerplateCtrl *Boilerplate
	testCaseCtrl    *TestCase // Add new controller var
)

// InitializeHandlers sets up all handler dependencies.
// Must be called before using any handlers.
func InitializeHandlers() {
	// Initialize repositories
	paperRepo := repositories.NewPaper(db.DB)
	questionRepo := repositories.NewQuestion(db.DB)
	categoryRepo := repositories.NewCategory(db.DB)
	boilerplateRepo := repositories.NewBoilerplate(db.DB)
	testCaseRepo := repositories.NewTestCase(db.DB)

	// Initialize services
	paperSvc := services.NewPaper(paperRepo)
	questionSvc := services.NewQuestion(paperRepo, questionRepo)
	categorySvc := services.NewCategory(categoryRepo, paperRepo, questionRepo)
	boilerplateSvc := services.NewBoilerplate(boilerplateRepo)
	testCaseSvc := services.NewTestCase(questionRepo, testCaseRepo)

	// Initialize controllers
	paperCtrl = NewPaper(paperSvc)
	categoryCtrl = NewCategory(categorySvc)
	questionCtrl = NewQuestion(questionSvc)
	boilerplateCtrl = NewBoilerplate(boilerplateSvc)
	testCaseCtrl = NewTestCase(testCaseSvc)
}

// _________________________PAPER HANDLERS__________________________

func (s *PaperServer) GetUserPapers(ctx context.Context, req *proto.Empty) (*proto.PaperList, error) {
	return paperCtrl.HandleGetUserPapers(ctx, req)
}

func (s *PaperServer) CreatePaper(ctx context.Context, req *proto.Empty) (*proto.PaperResponse, error) {
	return paperCtrl.HandleCreatePaper(ctx, req)
}

func (s *PaperServer) GetPaper(ctx context.Context, req *proto.PaperRequest) (*proto.PaperResponse, error) {
	return paperCtrl.HandleGetPaper(ctx, req)
}

func (s *PaperServer) UpdatePaper(ctx context.Context, req *proto.UpdatePaperRequest) (*proto.Empty, error) {
	return paperCtrl.HandleUpdatePaper(ctx, req)
}

func (s *PaperServer) GetPaperPermissions(ctx context.Context, req *proto.PaperRequest) (*proto.PaperPermissionsResponse, error) {
	return paperCtrl.HandleGetPaperPermissions(ctx, req)
}

func (s *PaperServer) DeletePapers(ctx context.Context, req *proto.DeletePapersRequest) (*proto.Empty, error) {
	return paperCtrl.HandleDeletePapers(ctx, req)
}

// ________________________QUESTION HANDLERS________________________

func (s *PaperServer) GetPaperQuestions(ctx context.Context, req *proto.PaperRequest) (*proto.QuestionList, error) {
	return questionCtrl.HandleGetPaperQuestions(ctx, req)
}

func (s *PaperServer) GetPaperQuestion(ctx context.Context, req *proto.QuestionRequest) (*proto.QuestionResponse, error) {
	return questionCtrl.HandleGetPaperQuestion(ctx, req)
}

func (s *PaperServer) CreateQuestion(ctx context.Context, req *proto.CreateQuestionRequest) (*proto.CreateQuestionResponse, error) {
	return questionCtrl.HandleCreateQuestion(ctx, req)
}

func (s *PaperServer) UpdateQuestion(ctx context.Context, req *proto.UpdateQuestionRequest) (*proto.UpdateQuestionResponse, error) {
	return questionCtrl.HandleUpdateQuestion(ctx, req)
}

func (s *PaperServer) DeleteQuestion(ctx context.Context, req *proto.QuestionRequest) (*proto.Empty, error) {
	return questionCtrl.HandleDeleteQuestion(ctx, req)
}

func (s *PaperServer) ReorderQuestions(ctx context.Context, req *proto.ReorderQuestionsRequest) (*proto.Empty, error) {
	return questionCtrl.HandleReorderQuestions(ctx, req)
}

// ________________________CATEGORY HANDLERS________________________

func (s *PaperServer) GetPaperCategories(ctx context.Context, req *proto.PaperRequest) (*proto.CategoryList, error) {
	return categoryCtrl.HandleGetPaperCategories(ctx, req)
}

func (s *PaperServer) CreateCategory(ctx context.Context, req *proto.CreateCategoryRequest) (*proto.CategoryResponse, error) {
	return categoryCtrl.HandleCreateCategory(ctx, req)
}

func (s *PaperServer) UpdateCategory(ctx context.Context, req *proto.UpdateCategoryRequest) (*proto.Empty, error) {
	return categoryCtrl.HandleUpdateCategory(ctx, req)
}

func (s *PaperServer) DeleteCategory(ctx context.Context, req *proto.CategoryRequest) (*proto.Empty, error) {
	return categoryCtrl.HandleDeleteCategory(ctx, req)
}

func (s *PaperServer) ReorderCategories(ctx context.Context, req *proto.ReorderCategoriesRequest) (*proto.Empty, error) {
	return categoryCtrl.HandleReorderCategories(ctx, req)
}

// __________________________EXAM HANDLERS__________________________
// These handlers are only meant for the Exam Service

func (s *PaperServer) GetQuestionsByIds(ctx context.Context, req *proto.GetQuestionsByIdsRequest) (*proto.GetQuestionsByIdsResponse, error) {
	return questionCtrl.HandleGetQuestionsByIds(ctx, req)
}

func (s *PaperServer) GetExamQuestion(ctx context.Context, req *proto.QuestionRequest) (*proto.QuestionResponse, error) {
	return questionCtrl.HandleGetExamQuestion(ctx, req)
}

func (s *PaperServer) GetQuestionHashes(ctx context.Context, req *proto.GetQuestionHashesRequest) (*proto.GetQuestionHashesResponse, error) {
	return questionCtrl.HandleGetQuestionHashes(ctx, req)
}

func (s *PaperServer) GetQuestionIds(ctx context.Context, req *proto.GetQuestionIdsRequest) (*proto.GetQuestionIdsResponse, error) {
	return questionCtrl.HandleGetQuestionIds(ctx, req)
}

// GetCategoriesByIds retrieves multiple categories by their IDs in a single request
func (s *PaperServer) GetCategoriesByIds(ctx context.Context, req *proto.GetCategoriesByIdsRequest) (*proto.CategoryBatchResponse, error) {
	return categoryCtrl.HandleGetCategoriesByIds(ctx, req)
}

// ______________________BOILERPLATE HANDLERS_______________________

func (s *PaperServer) GetBoilerplate(ctx context.Context, req *proto.GetBoilerplateRequest) (*proto.GetBoilerplateResponse, error) {
	return boilerplateCtrl.HandleGetBoilerplate(ctx, req)
}

// ________________________TESTCASE HANDLERS________________________

func (s *PaperServer) UpsertPaperTestCases(ctx context.Context, req *proto.UpsertTestCasesRequest) (*proto.Empty, error) {
	return testCaseCtrl.HandleUpsertTestCases(ctx, req)
}
