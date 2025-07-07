package controllers

import (
	"context"

	"google.golang.org/protobuf/types/known/emptypb"

	"pariksha/common/pkg/proto"
	"pariksha/common/pkg/utils/grpcerror"
	"pariksha/paper/internal/services"
)

type Question struct {
	questionSvc *services.Question
}

func NewQuestion(s *services.Question) *Question {
	return &Question{
		questionSvc: s,
	}
}

func (c *Question) HandleGetPaperQuestions(ctx context.Context, req *proto.PaperRequest) (*proto.QuestionList, error) {
	return c.questionSvc.GetPaperQuestions(ctx, req.PaperHash)
}

func (c *Question) HandleGetPaperQuestion(ctx context.Context, req *proto.PaperQuestionRequest) (*proto.PaperQuestionResponse, error) {
	return c.questionSvc.GetPaperQuestion(req)
}

func (c *Question) HandleCreateQuestion(ctx context.Context, req *proto.CreatePaperQuestionRequest) (*proto.CreatePaperQuestionResponse, error) {
	return c.questionSvc.CreatePaperQuestion(ctx, req)
}

func (c *Question) HandleUpdateQuestion(ctx context.Context, req *proto.UpdatePaperQuestionRequest) (*proto.UpdatePaperQuestionResponse, error) {
	return c.questionSvc.UpdatePaperQuestion(req)
}

func (c *Question) HandleDeleteQuestion(ctx context.Context, req *proto.PaperQuestionRequest) (*emptypb.Empty, error) {
	if err := c.questionSvc.DeletePaperQuestion(req.PaperHash, req.QuestionHash); err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}

func (c *Question) HandleReorderQuestions(ctx context.Context, req *proto.ReorderPaperQuestionsRequest) (*emptypb.Empty, error) {
	if err := c.questionSvc.ReorderQuestions(req.CategoryId, req.QuestionHashes); err != nil {
		return nil, grpcerror.Internal(err, "failed to reorder questions")
	}
	return &emptypb.Empty{}, nil
}

func (c *Question) HandleGetBoilerplate(ctx context.Context, req *proto.GetPaperBoilerplateRequest) (*proto.BoilerplateResponse, error) {
	return c.questionSvc.GetBoilerplate(req)
}

func (c *Question) HandleUpsertTestCases(ctx context.Context, req *proto.UpsertPaperTestCasesRequest) (*emptypb.Empty, error) {
	return c.questionSvc.UpsertTestCases(req)
}

func (c *Question) HandleGetPaperQuestionsMeta(ctx context.Context, req *proto.PaperRequest) (*proto.PaperQuestionsMeta, error) {
	return c.questionSvc.GetPaperQuestionsMeta(req)
}
