package modules

import (
	"context"

	"google.golang.org/protobuf/types/known/emptypb"

	"pariksha/common/pkg/proto"
	"pariksha/question/internal/controllers"
)

type questionServer struct {
	proto.UnimplementedQuestionServer

	questionCtrl *controllers.Question
	categoryCtrl *controllers.Category
}

// __________________________QUESTION HANDLERS_________________________

func (s *questionServer) CreateQuestion(ctx context.Context, req *proto.CreateQuestionRequest) (*proto.CreateQuestionResponse, error) {
	return s.questionCtrl.HandleCreateQuestion(ctx, req)
}

func (s *questionServer) UpdateQuestion(ctx context.Context, req *proto.UpdateQuestionRequest) (*proto.UpdateQuestionResponse, error) {
	return s.questionCtrl.HandleUpdateQuestion(ctx, req)
}

func (s *questionServer) GetQuestionsMetaByHashes(ctx context.Context, req *proto.QuestionHashesRequest) (*proto.QuestionsMetaResponse, error) {
	return s.questionCtrl.HandleGetQuestionsMetaByHashes(ctx, req)
}

func (s *questionServer) GetQuestionIdsByHashes(ctx context.Context, req *proto.QuestionHashesRequest) (*proto.GetQuestionIdsByHashesResponse, error) {
	return s.questionCtrl.HandleGetQuestionIdsByHashes(ctx, req)
}

func (s *questionServer) GetQuestionHashesByIds(ctx context.Context, req *proto.QuestionIdsRequest) (*proto.GetQuestionHashesByIdsResponse, error) {
	return s.questionCtrl.HandleGetQuestionHashesByIds(ctx, req)
}

func (s *questionServer) IncQuestionPaperIndegreeByIds(ctx context.Context, req *proto.QuestionIdsRequest) (*emptypb.Empty, error) {
	return s.questionCtrl.HandleIncQuestionPaperIndegreeByIds(ctx, req)
}

func (s *questionServer) DecQuestionPaperIndegreeByIds(ctx context.Context, req *proto.QuestionIdsRequest) (*emptypb.Empty, error) {
	return s.questionCtrl.HandleDecQuestionPaperIndegreeByIds(ctx, req)
}

func (s *questionServer) IncQuestionExamIndegreeByIds(ctx context.Context, req *proto.QuestionIdsRequest) (*emptypb.Empty, error) {
	return s.questionCtrl.HandleIncQuestionExamIndegreeByIds(ctx, req)
}

func (s *questionServer) DecQuestionExamIndegreeByIds(ctx context.Context, req *proto.QuestionIdsRequest) (*emptypb.Empty, error) {
	return s.questionCtrl.HandleDecQuestionExamIndegreeByIds(ctx, req)
}

func (s *questionServer) GetQuestionsByHashes(ctx context.Context, req *proto.QuestionHashesRequest) (*proto.GetQuestionsResponse, error) {
	return s.questionCtrl.HandleGetQuestionsByHashes(ctx, req)
}

func (s *questionServer) GetQuestionsByIds(ctx context.Context, req *proto.QuestionIdsRequest) (*proto.GetQuestionsResponse, error) {
	return s.questionCtrl.HandleGetQuestionsByIds(ctx, req)
}

func (s *questionServer) GetQuestionsMetaByIds(ctx context.Context, req *proto.QuestionIdsRequest) (*proto.QuestionsMetaResponse, error) {
	return s.questionCtrl.HandleGetQuestionsMetaByIds(ctx, req)
}

func (s *questionServer) UpsertTestCases(ctx context.Context, req *proto.UpsertTestCasesRequest) (*emptypb.Empty, error) {
	return s.questionCtrl.HandleUpsertTestCases(ctx, req)
}

func (s *questionServer) GetBoilerplate(ctx context.Context, req *proto.GetBoilerplateRequest) (*proto.BoilerplateResponse, error) {
	return s.questionCtrl.HandleGetBoilerplate(ctx, req)
}

// __________________________CATEGORY HANDLERS_________________________

func (s *questionServer) CreateCategory(ctx context.Context, req *proto.CreateCategoryRequest) (*proto.CategoryResponse, error) {
	return s.categoryCtrl.HandleCreateCategory(ctx, req)
}

func (s *questionServer) GetCategoriesByIds(ctx context.Context, req *proto.CategoryIdsRequest) (*proto.GetCategoriesResponse, error) {
	return s.categoryCtrl.HandleGetCategoriesByIds(ctx, req)
}

func (s *questionServer) UpdateCategoryName(ctx context.Context, req *proto.UpdateCategoryRequest) (*proto.UpdateCategoryResponse, error) {
	return s.categoryCtrl.HandleUpdateCategoryName(ctx, req)
}

// _____________________ENGINE-SPECIFIC HANDLERS_____________________

func (s *questionServer) GetCodingQuestionInputDefinitions(ctx context.Context, req *proto.GetCodingQuestionInputDefinitionsRequest) (*proto.GetCodingQuestionInputDefinitionsResponse, error) {
	return s.questionCtrl.HandleGetCodingQuestionInputDefinitions(ctx, req)
}
