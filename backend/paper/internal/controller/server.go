package controllers

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"

	"pariksha/common/pkg/proto"
	"pariksha/paper/internal/service"
)

type Server struct {
	proto.UnimplementedPaperServer

	paperCtrl    *Paper
	categoryCtrl *Category
	questionCtrl *Question
}

func NewServer(paperSvc service.IPaperService, categorySvc service.ICategoryService, questionSvc service.IQuestionService) *Server {
	return &Server{
		paperCtrl:    NewPaper(paperSvc),
		categoryCtrl: NewCategory(categorySvc),
		questionCtrl: NewQuestion(questionSvc),
	}
}

// SetupPaperServer registers the paper gRPC server with controller wiring.
func SetupPaperServer(
	grpcServer *grpc.Server,
	paperSvc service.IPaperService,
	categorySvc service.ICategoryService,
	questionSvc service.IQuestionService,
) {
	proto.RegisterPaperServer(grpcServer, NewServer(paperSvc, categorySvc, questionSvc))
}

// _________________________PAPER HANDLERS__________________________

func (s *Server) GetUserPapers(ctx context.Context, req *emptypb.Empty) (*proto.PaperList, error) {
	return s.paperCtrl.HandleGetUserPapers(ctx, req)
}

func (s *Server) CreatePaper(ctx context.Context, req *emptypb.Empty) (*proto.CreatePaperResponse, error) {
	return s.paperCtrl.HandleCreatePaper(ctx, req)
}

func (s *Server) GetPaper(ctx context.Context, req *proto.PaperRequest) (*proto.PaperResponse, error) {
	return s.paperCtrl.HandleGetPaper(ctx, req)
}

func (s *Server) UpdatePaper(ctx context.Context, req *proto.UpdatePaperRequest) (*emptypb.Empty, error) {
	return s.paperCtrl.HandleUpdatePaper(ctx, req)
}

func (s *Server) GetPaperPermissions(ctx context.Context, req *proto.PaperRequest) (*proto.PaperPermissionsResponse, error) {
	return s.paperCtrl.HandleGetPaperPermissions(ctx, req)
}

func (s *Server) DeletePapers(ctx context.Context, req *proto.DeletePapersRequest) (*emptypb.Empty, error) {
	return s.paperCtrl.HandleDeletePapers(ctx, req)
}

// ________________________QUESTION HANDLERS________________________

func (s *Server) GetPaperQuestions(ctx context.Context, req *proto.PaperRequest) (*proto.QuestionList, error) {
	return s.questionCtrl.HandleGetPaperQuestions(ctx, req)
}

func (s *Server) GetPaperQuestion(ctx context.Context, req *proto.PaperQuestionRequest) (*proto.PaperQuestionResponse, error) {
	return s.questionCtrl.HandleGetPaperQuestion(ctx, req)
}

func (s *Server) CreatePaperQuestion(ctx context.Context, req *proto.CreatePaperQuestionRequest) (*proto.CreatePaperQuestionResponse, error) {
	return s.questionCtrl.HandleCreateQuestion(ctx, req)
}

func (s *Server) UpdatePaperQuestion(ctx context.Context, req *proto.UpdatePaperQuestionRequest) (*proto.UpdatePaperQuestionResponse, error) {
	return s.questionCtrl.HandleUpdateQuestion(ctx, req)
}

func (s *Server) DeletePaperQuestion(ctx context.Context, req *proto.PaperQuestionRequest) (*emptypb.Empty, error) {
	return s.questionCtrl.HandleDeleteQuestion(ctx, req)
}

func (s *Server) ReorderPaperQuestions(ctx context.Context, req *proto.ReorderPaperQuestionsRequest) (*emptypb.Empty, error) {
	return s.questionCtrl.HandleReorderQuestions(ctx, req)
}

func (s *Server) GetPaperBoilerplate(ctx context.Context, req *proto.GetPaperBoilerplateRequest) (*proto.BoilerplateResponse, error) {
	return s.questionCtrl.HandleGetBoilerplate(ctx, req)
}

func (s *Server) UpsertPaperTestCases(ctx context.Context, req *proto.UpsertPaperTestCasesRequest) (*emptypb.Empty, error) {
	return s.questionCtrl.HandleUpsertTestCases(ctx, req)
}

// ________________________CATEGORY HANDLERS________________________

func (s *Server) GetPaperCategories(ctx context.Context, req *proto.PaperRequest) (*proto.PaperCategoryList, error) {
	return s.categoryCtrl.HandleGetPaperCategories(ctx, req)
}

func (s *Server) CreatePaperCategory(ctx context.Context, req *proto.PaperRequest) (*proto.PaperCategoryResponse, error) {
	return s.categoryCtrl.HandleCreateCategory(ctx, req)
}

func (s *Server) UpdatePaperCategory(ctx context.Context, req *proto.UpdatePaperCategoryRequest) (*emptypb.Empty, error) {
	return s.categoryCtrl.HandleUpdateCategory(ctx, req)
}

func (s *Server) DeletePaperCategory(ctx context.Context, req *proto.PaperCategoryRequest) (*emptypb.Empty, error) {
	return s.categoryCtrl.HandleDeleteCategory(ctx, req)
}

func (s *Server) ReorderPaperCategories(ctx context.Context, req *proto.ReorderPaperCategoriesRequest) (*emptypb.Empty, error) {
	return s.categoryCtrl.HandleReorderCategories(ctx, req)
}

// __________________________EXAM HANDLERS__________________________

func (s *Server) GetPaperQuestionsMeta(ctx context.Context, req *proto.PaperRequest) (*proto.PaperQuestionsMeta, error) {
	return s.questionCtrl.HandleGetPaperQuestionsMeta(ctx, req)
}

func (s *Server) GetPaperCategoriesMeta(ctx context.Context, req *proto.PaperRequest) (*proto.PaperCategoriesMeta, error) {
	return s.categoryCtrl.HandleGetPaperCategoriesMeta(ctx, req)
}
