package controllers

import (
	"context"

	"google.golang.org/protobuf/types/known/emptypb"

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
	paperCtrl    *Paper
	categoryCtrl *Category
	questionCtrl *Question
)

// InitializeHandlers sets up all handler dependencies.
// Must be called before using any handlers.
func InitializeHandlers() {
	// Initialize repositories
	paperRepo := repositories.NewPaper(db.DB)
	paperCatRepo := repositories.NewPaperCategory(db.DB)
	paperPermRepo := repositories.NewPaperPermission(db.DB)
	paperQuestRepo := repositories.NewPaperQuestion(db.DB)

	// Initialize services
	paperSvc := services.NewPaper(paperRepo, paperCatRepo, paperPermRepo, paperQuestRepo)
	questionSvc := services.NewQuestion(paperRepo, paperQuestRepo)
	categorySvc := services.NewCategory(paperRepo, paperCatRepo, paperQuestRepo)

	// Initialize controllers
	paperCtrl = NewPaper(paperSvc)
	categoryCtrl = NewCategory(categorySvc)
	questionCtrl = NewQuestion(questionSvc)
}

// _________________________PAPER HANDLERS__________________________

func (s *PaperServer) GetUserPapers(ctx context.Context, req *emptypb.Empty) (*proto.PaperList, error) {
	return paperCtrl.HandleGetUserPapers(ctx, req)
}

func (s *PaperServer) CreatePaper(ctx context.Context, req *emptypb.Empty) (*proto.CreatePaperResponse, error) {
	return paperCtrl.HandleCreatePaper(ctx, req)
}

func (s *PaperServer) GetPaper(ctx context.Context, req *proto.PaperRequest) (*proto.PaperResponse, error) {
	return paperCtrl.HandleGetPaper(ctx, req)
}

func (s *PaperServer) UpdatePaper(ctx context.Context, req *proto.UpdatePaperRequest) (*emptypb.Empty, error) {
	return paperCtrl.HandleUpdatePaper(ctx, req)
}

func (s *PaperServer) GetPaperPermissions(ctx context.Context, req *proto.PaperRequest) (*proto.PaperPermissionsResponse, error) {
	return paperCtrl.HandleGetPaperPermissions(ctx, req)
}

func (s *PaperServer) DeletePapers(ctx context.Context, req *proto.DeletePapersRequest) (*emptypb.Empty, error) {
	return paperCtrl.HandleDeletePapers(ctx, req)
}

// ________________________QUESTION HANDLERS________________________

func (s *PaperServer) GetPaperQuestions(ctx context.Context, req *proto.PaperRequest) (*proto.QuestionList, error) {
	return questionCtrl.HandleGetPaperQuestions(ctx, req)
}

func (s *PaperServer) GetPaperQuestion(ctx context.Context, req *proto.PaperQuestionRequest) (*proto.PaperQuestionResponse, error) {
	return questionCtrl.HandleGetPaperQuestion(ctx, req)
}

func (s *PaperServer) CreatePaperQuestion(ctx context.Context, req *proto.CreatePaperQuestionRequest) (*proto.CreatePaperQuestionResponse, error) {
	return questionCtrl.HandleCreateQuestion(ctx, req)
}

func (s *PaperServer) UpdatePaperQuestion(ctx context.Context, req *proto.UpdatePaperQuestionRequest) (*proto.UpdatePaperQuestionResponse, error) {
	return questionCtrl.HandleUpdateQuestion(ctx, req)
}

func (s *PaperServer) DeletePaperQuestion(ctx context.Context, req *proto.PaperQuestionRequest) (*emptypb.Empty, error) {
	return questionCtrl.HandleDeleteQuestion(ctx, req)
}

func (s *PaperServer) ReorderPaperQuestions(ctx context.Context, req *proto.ReorderPaperQuestionsRequest) (*emptypb.Empty, error) {
	return questionCtrl.HandleReorderQuestions(ctx, req)
}

func (s *PaperServer) GetPaperBoilerplate(ctx context.Context, req *proto.GetPaperBoilerplateRequest) (*proto.BoilerplateResponse, error) {
	return questionCtrl.HandleGetBoilerplate(ctx, req)
}

func (s *PaperServer) UpsertPaperTestCases(ctx context.Context, req *proto.UpsertPaperTestCasesRequest) (*emptypb.Empty, error) {
	return questionCtrl.HandleUpsertTestCases(ctx, req)
}

// ________________________CATEGORY HANDLERS________________________

func (s *PaperServer) GetPaperCategories(ctx context.Context, req *proto.PaperRequest) (*proto.PaperCategoryList, error) {
	return categoryCtrl.HandleGetPaperCategories(ctx, req)
}

func (s *PaperServer) CreatePaperCategory(ctx context.Context, req *proto.CreatePaperCategoryRequest) (*proto.PaperCategoryResponse, error) {
	return categoryCtrl.HandleCreateCategory(ctx, req)
}

func (s *PaperServer) UpdatePaperCategory(ctx context.Context, req *proto.UpdatePaperCategoryRequest) (*emptypb.Empty, error) {
	return categoryCtrl.HandleUpdateCategory(ctx, req)
}

func (s *PaperServer) DeletePaperCategory(ctx context.Context, req *proto.PaperCategoryRequest) (*emptypb.Empty, error) {
	return categoryCtrl.HandleDeleteCategory(ctx, req)
}

func (s *PaperServer) ReorderPaperCategories(ctx context.Context, req *proto.ReorderPaperCategoriesRequest) (*emptypb.Empty, error) {
	return categoryCtrl.HandleReorderCategories(ctx, req)
}

// __________________________EXAM HANDLERS__________________________

func (s *PaperServer) GetPaperQuestionsMeta(ctx context.Context, req *proto.PaperRequest) (*proto.PaperQuestionsMeta, error) {
	return questionCtrl.HandleGetPaperQuestionsMeta(ctx, req)
}

func (s *PaperServer) GetPaperCategoriesMeta(ctx context.Context, req *proto.PaperRequest) (*proto.PaperCategoriesMeta, error) {
	return categoryCtrl.HandleGetPaperCategoriesMeta(ctx, req)
}
