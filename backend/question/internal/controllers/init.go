package controllers

import (
	"context"

	"google.golang.org/protobuf/types/known/emptypb"

	"pariksha/common/pkg/proto"
	"pariksha/question/internal/config/db"
	"pariksha/question/internal/repositories"
	"pariksha/question/internal/services"
)

type QuestionServer struct {
	proto.UnimplementedQuestionServer
}

var (
	questionCtrl *Question
	categoryCtrl *Category
)

// Init sets up all handler dependencies.
// Must be called before using any handlers.
func Init() {
	// Initialize repositories
	questionRepo := repositories.NewQuestion(db.DB)
	categoryRepo := repositories.NewCategory(db.DB)
	boilerplateRepo := repositories.NewBoilerplate(db.DB)
	testcaseRepo := repositories.NewTestcase(db.DB)

	// Initialize services
	questionSvc := services.NewQuestion(questionRepo, boilerplateRepo, testcaseRepo)
	categorySvc := services.NewCategory(categoryRepo)

	// Initialize controllers
	questionCtrl = NewQuestion(questionSvc)
	categoryCtrl = NewCategory(categorySvc)
}

// __________________________QUESTION HANDLERS_________________________

func (s *QuestionServer) CreateQuestion(ctx context.Context, req *proto.CreateQuestionRequest) (*proto.CreateQuestionResponse, error) {
	return questionCtrl.HandleCreateQuestion(ctx, req)
}

func (s *QuestionServer) UpdateQuestion(ctx context.Context, req *proto.UpdateQuestionRequest) (*proto.UpdateQuestionResponse, error) {
	return questionCtrl.HandleUpdateQuestion(ctx, req)
}

func (s *QuestionServer) GetQuestionsMetaByHashes(ctx context.Context, req *proto.QuestionHashesRequest) (*proto.QuestionsMetaResponse, error) {
	return questionCtrl.HandleGetQuestionsMetaByHashes(ctx, req)
}

func (s *QuestionServer) GetQuestionIdsByHashes(ctx context.Context, req *proto.QuestionHashesRequest) (*proto.GetQuestionIdsByHashesResponse, error) {
	return questionCtrl.HandleGetQuestionIdsByHashes(ctx, req)
}

func (s *QuestionServer) GetQuestionHashesByIds(ctx context.Context, req *proto.QuestionIdsRequest) (*proto.GetQuestionHashesByIdsResponse, error) {
	return questionCtrl.HandleGetQuestionHashesByIds(ctx, req)
}

func (s *QuestionServer) IncQuestionPaperIndegreeByIds(ctx context.Context, req *proto.QuestionIdsRequest) (*emptypb.Empty, error) {
	return questionCtrl.HandleIncQuestionPaperIndegreeByIds(ctx, req)
}

func (s *QuestionServer) DecQuestionPaperIndegreeByIds(ctx context.Context, req *proto.QuestionIdsRequest) (*emptypb.Empty, error) {
	return questionCtrl.HandleDecQuestionPaperIndegreeByIds(ctx, req)
}

func (s *QuestionServer) IncQuestionExamIndegreeByIds(ctx context.Context, req *proto.QuestionIdsRequest) (*emptypb.Empty, error) {
	return questionCtrl.HandleIncQuestionExamIndegreeByIds(ctx, req)
}

func (s *QuestionServer) DecQuestionExamIndegreeByIds(ctx context.Context, req *proto.QuestionIdsRequest) (*emptypb.Empty, error) {
	return questionCtrl.HandleDecQuestionExamIndegreeByIds(ctx, req)
}

func (s *QuestionServer) GetQuestionsByHashes(ctx context.Context, req *proto.QuestionHashesRequest) (*proto.GetQuestionsResponse, error) {
	return questionCtrl.HandleGetQuestionsByHashes(ctx, req)
}

func (s *QuestionServer) GetQuestionsByIds(ctx context.Context, req *proto.QuestionIdsRequest) (*proto.GetQuestionsResponse, error) {
	return questionCtrl.HandleGetQuestionsByIds(ctx, req)
}

func (s *QuestionServer) GetQuestionsMetaByIds(ctx context.Context, req *proto.QuestionIdsRequest) (*proto.QuestionsMetaResponse, error) {
	return questionCtrl.HandleGetQuestionsMetaByIds(ctx, req)
}

func (s *QuestionServer) UpsertTestCases(ctx context.Context, req *proto.UpsertTestCasesRequest) (*emptypb.Empty, error) {
	return questionCtrl.HandleUpsertTestCases(ctx, req)
}

func (s *QuestionServer) GetBoilerplate(ctx context.Context, req *proto.GetBoilerplateRequest) (*proto.BoilerplateResponse, error) {
	return questionCtrl.HandleGetBoilerplate(ctx, req)
}

// __________________________CATEGORY HANDLERS_________________________

func (s *QuestionServer) CreateCategory(ctx context.Context, req *proto.CreateCategoryRequest) (*proto.CategoryResponse, error) {
	return categoryCtrl.HandleCreateCategory(ctx, req)
}

func (s *QuestionServer) GetCategoriesByIds(ctx context.Context, req *proto.CategoryIdsRequest) (*proto.GetCategoriesResponse, error) {
	return categoryCtrl.HandleGetCategoriesByIds(ctx, req)
}

func (s *QuestionServer) UpdateCategoryName(ctx context.Context, req *proto.UpdateCategoryRequest) (*proto.UpdateCategoryResponse, error) {
	return categoryCtrl.HandleUpdateCategoryName(ctx, req)
}

// _____________________ENGINE-SPECIFIC HANDLERS_____________________

func (s *QuestionServer) GetCodingQuestionInputDefinitions(ctx context.Context, req *proto.GetCodingQuestionInputDefinitionsRequest) (*proto.GetCodingQuestionInputDefinitionsResponse, error) {
	return questionCtrl.HandleGetCodingQuestionInputDefinitions(ctx, req)
}
