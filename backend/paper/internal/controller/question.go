package controllers

import (
	"context"

	"google.golang.org/protobuf/types/known/emptypb"

	"pariksha/common/pkg/proto"
	"pariksha/paper/internal/service"
)

type Question struct {
	questionSvc service.IQuestionService
}

func NewQuestion(s service.IQuestionService) *Question {
	return &Question{
		questionSvc: s,
	}
}

func (c *Question) HandleGetPaperQuestions(ctx context.Context, req *proto.PaperRequest) (*proto.QuestionList, error) {
	return c.questionSvc.GetPaperQuestions(ctx, req.PaperHash)
}

func (c *Question) HandleGetPaperQuestion(ctx context.Context, req *proto.PaperQuestionRequest) (*proto.PaperQuestionResponse, error) {
	return c.questionSvc.GetPaperQuestion(ctx, req)
}

func (c *Question) HandleCreateQuestion(ctx context.Context, req *proto.CreatePaperQuestionRequest) (*proto.CreatePaperQuestionResponse, error) {
	return c.questionSvc.CreatePaperQuestion(ctx, req)
}

func (c *Question) HandleUpdateQuestion(ctx context.Context, req *proto.UpdatePaperQuestionRequest) (*proto.UpdatePaperQuestionResponse, error) {
	return c.questionSvc.UpdatePaperQuestion(ctx, req)
}

func (c *Question) HandleDeleteQuestion(ctx context.Context, req *proto.PaperQuestionRequest) (*emptypb.Empty, error) {
	if err := c.questionSvc.DeletePaperQuestion(ctx, req.PaperHash, req.QuestionHash); err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}

func (c *Question) HandleReorderQuestions(ctx context.Context, req *proto.ReorderPaperQuestionsRequest) (*emptypb.Empty, error) {
	if err := c.questionSvc.ReorderQuestions(ctx, req.CategoryId, req.QuestionHashes); err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}

func (c *Question) HandleGetBoilerplate(ctx context.Context, req *proto.GetPaperBoilerplateRequest) (*proto.BoilerplateResponse, error) {
	return c.questionSvc.GetBoilerplate(ctx, req)
}

func (c *Question) HandleUpsertTestCases(ctx context.Context, req *proto.UpsertPaperTestCasesRequest) (*emptypb.Empty, error) {
	return c.questionSvc.UpsertTestCases(ctx, req)
}

func (c *Question) HandleGetPaperQuestionsMeta(ctx context.Context, req *proto.PaperRequest) (*proto.PaperQuestionsMeta, error) {
	return c.questionSvc.GetPaperQuestionsMeta(ctx, req)
}
