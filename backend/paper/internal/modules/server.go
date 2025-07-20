package modules

import (
	"context"

	"google.golang.org/protobuf/types/known/emptypb"

	"pariksha/common/pkg/proto"
	"pariksha/paper/internal/controllers"
)

type paperServer struct {
	proto.UnimplementedPaperServer

	paperCtrl    *controllers.Paper
	categoryCtrl *controllers.Category
	questionCtrl *controllers.Question
}

// _________________________PAPER HANDLERS__________________________

func (s *paperServer) GetUserPapers(ctx context.Context, req *emptypb.Empty) (*proto.PaperList, error) {
	return s.paperCtrl.HandleGetUserPapers(ctx, req)
}

func (s *paperServer) CreatePaper(ctx context.Context, req *emptypb.Empty) (*proto.CreatePaperResponse, error) {
	return s.paperCtrl.HandleCreatePaper(ctx, req)
}

func (s *paperServer) GetPaper(ctx context.Context, req *proto.PaperRequest) (*proto.PaperResponse, error) {
	return s.paperCtrl.HandleGetPaper(ctx, req)
}

func (s *paperServer) UpdatePaper(ctx context.Context, req *proto.UpdatePaperRequest) (*emptypb.Empty, error) {
	return s.paperCtrl.HandleUpdatePaper(ctx, req)
}

func (s *paperServer) GetPaperPermissions(ctx context.Context, req *proto.PaperRequest) (*proto.PaperPermissionsResponse, error) {
	return s.paperCtrl.HandleGetPaperPermissions(ctx, req)
}

func (s *paperServer) DeletePapers(ctx context.Context, req *proto.DeletePapersRequest) (*emptypb.Empty, error) {
	return s.paperCtrl.HandleDeletePapers(ctx, req)
}

// ________________________QUESTION HANDLERS________________________

func (s *paperServer) GetPaperQuestions(ctx context.Context, req *proto.PaperRequest) (*proto.QuestionList, error) {
	return s.questionCtrl.HandleGetPaperQuestions(ctx, req)
}

func (s *paperServer) GetPaperQuestion(ctx context.Context, req *proto.PaperQuestionRequest) (*proto.PaperQuestionResponse, error) {
	return s.questionCtrl.HandleGetPaperQuestion(ctx, req)
}

func (s *paperServer) CreatePaperQuestion(ctx context.Context, req *proto.CreatePaperQuestionRequest) (*proto.CreatePaperQuestionResponse, error) {
	return s.questionCtrl.HandleCreateQuestion(ctx, req)
}

func (s *paperServer) UpdatePaperQuestion(ctx context.Context, req *proto.UpdatePaperQuestionRequest) (*proto.UpdatePaperQuestionResponse, error) {
	return s.questionCtrl.HandleUpdateQuestion(ctx, req)
}

func (s *paperServer) DeletePaperQuestion(ctx context.Context, req *proto.PaperQuestionRequest) (*emptypb.Empty, error) {
	return s.questionCtrl.HandleDeleteQuestion(ctx, req)
}

func (s *paperServer) ReorderPaperQuestions(ctx context.Context, req *proto.ReorderPaperQuestionsRequest) (*emptypb.Empty, error) {
	return s.questionCtrl.HandleReorderQuestions(ctx, req)
}

func (s *paperServer) GetPaperBoilerplate(ctx context.Context, req *proto.GetPaperBoilerplateRequest) (*proto.BoilerplateResponse, error) {
	return s.questionCtrl.HandleGetBoilerplate(ctx, req)
}

func (s *paperServer) UpsertPaperTestCases(ctx context.Context, req *proto.UpsertPaperTestCasesRequest) (*emptypb.Empty, error) {
	return s.questionCtrl.HandleUpsertTestCases(ctx, req)
}

// ________________________CATEGORY HANDLERS________________________

func (s *paperServer) GetPaperCategories(ctx context.Context, req *proto.PaperRequest) (*proto.PaperCategoryList, error) {
	return s.categoryCtrl.HandleGetPaperCategories(ctx, req)
}

func (s *paperServer) CreatePaperCategory(ctx context.Context, req *proto.CreatePaperCategoryRequest) (*proto.PaperCategoryResponse, error) {
	return s.categoryCtrl.HandleCreateCategory(ctx, req)
}

func (s *paperServer) UpdatePaperCategory(ctx context.Context, req *proto.UpdatePaperCategoryRequest) (*emptypb.Empty, error) {
	return s.categoryCtrl.HandleUpdateCategory(ctx, req)
}

func (s *paperServer) DeletePaperCategory(ctx context.Context, req *proto.PaperCategoryRequest) (*emptypb.Empty, error) {
	return s.categoryCtrl.HandleDeleteCategory(ctx, req)
}

func (s *paperServer) ReorderPaperCategories(ctx context.Context, req *proto.ReorderPaperCategoriesRequest) (*emptypb.Empty, error) {
	return s.categoryCtrl.HandleReorderCategories(ctx, req)
}

// __________________________EXAM HANDLERS__________________________

func (s *paperServer) GetPaperQuestionsMeta(ctx context.Context, req *proto.PaperRequest) (*proto.PaperQuestionsMeta, error) {
	return s.questionCtrl.HandleGetPaperQuestionsMeta(ctx, req)
}

func (s *paperServer) GetPaperCategoriesMeta(ctx context.Context, req *proto.PaperRequest) (*proto.PaperCategoriesMeta, error) {
	return s.categoryCtrl.HandleGetPaperCategoriesMeta(ctx, req)
}
