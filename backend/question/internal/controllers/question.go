package controllers

import (
	"context"

	"pariksha/common/pkg/proto"
	"pariksha/question/internal/services"

	"google.golang.org/protobuf/types/known/emptypb"
)

type Question struct {
	questionSvc *services.Question
}

func NewQuestion(questionSvc *services.Question) *Question {
	return &Question{questionSvc: questionSvc}
}

func (c *Question) HandleCreateQuestion(ctx context.Context, req *proto.CreateQuestionRequest) (*proto.CreateQuestionResponse, error) {
	return c.questionSvc.CreateQuestion(req)
}

func (c *Question) HandleUpdateQuestion(ctx context.Context, req *proto.UpdateQuestionRequest) (*proto.UpdateQuestionResponse, error) {
	return c.questionSvc.UpdateQuestion(req)
}

func (c *Question) HandleGetQuestionsByIds(ctx context.Context, req *proto.QuestionIdsRequest) (*proto.GetQuestionsResponse, error) {
	return c.questionSvc.GetQuestionsByIds(req)
}

func (c *Question) HandleGetQuestionsMetaByIds(ctx context.Context, req *proto.QuestionIdsRequest) (*proto.QuestionsMetaResponse, error) {
	return c.questionSvc.GetQuestionsMetaByIds(req)
}

func (c *Question) HandleGetQuestionsMetaByHashes(ctx context.Context, req *proto.QuestionHashesRequest) (*proto.QuestionsMetaResponse, error) {
	return c.questionSvc.GetQuestionsMetaByHashes(req.Hashes)
}

func (c *Question) HandleGetQuestionHashesByIds(ctx context.Context, req *proto.QuestionIdsRequest) (*proto.GetQuestionHashesByIdsResponse, error) {
	return c.questionSvc.GetQuestionHashesByIds(req)
}

func (c *Question) HandleGetQuestionIdsByHashes(ctx context.Context, req *proto.QuestionHashesRequest) (*proto.GetQuestionIdsByHashesResponse, error) {
	return c.questionSvc.GetQuestionIdsByHashes(req.Hashes)
}

func (c *Question) HandleIncQuestionPaperIndegreeByIds(ctx context.Context, req *proto.QuestionIdsRequest) (*emptypb.Empty, error) {
	return c.questionSvc.IncPaperIndegreeByIds(req.Ids)
}

func (c *Question) HandleDecQuestionPaperIndegreeByIds(ctx context.Context, req *proto.QuestionIdsRequest) (*emptypb.Empty, error) {
	return c.questionSvc.DecPaperIndegreeByIds(req.Ids)
}

func (c *Question) HandleIncQuestionExamIndegreeByIds(ctx context.Context, req *proto.QuestionIdsRequest) (*emptypb.Empty, error) {
	return c.questionSvc.IncExamIndegreeByIds(req.Ids)
}

func (c *Question) HandleDecQuestionExamIndegreeByIds(ctx context.Context, req *proto.QuestionIdsRequest) (*emptypb.Empty, error) {
	return c.questionSvc.DecExamIndegreeByIds(req.Ids)
}

func (c *Question) HandleGetQuestionsByHashes(ctx context.Context, req *proto.QuestionHashesRequest) (*proto.GetQuestionsResponse, error) {
	return c.questionSvc.GetQuestionsByHashes(req.Hashes)
}

func (c *Question) HandleUpsertTestCases(ctx context.Context, req *proto.UpsertTestCasesRequest) (*emptypb.Empty, error) {
	return c.questionSvc.UpsertTestCases(req)
}

func (c *Question) HandleGetBoilerplate(ctx context.Context, req *proto.GetBoilerplateRequest) (*proto.BoilerplateResponse, error) {
	return c.questionSvc.GetBoilerplate(req)
}

func (c *Question) HandleGetCodingQuestionInputDefinitions(ctx context.Context, req *proto.GetCodingQuestionInputDefinitionsRequest) (*proto.GetCodingQuestionInputDefinitionsResponse, error) {
	return c.questionSvc.GetCodingQuestionInputDefinitions(req)
}
