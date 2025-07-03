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

// HandleGetPaperQuestions handles getting all questions for a paper
func (c *Question) HandleGetPaperQuestions(ctx context.Context, req *proto.PaperRequest) (*proto.QuestionList, error) {
	return c.questionSvc.GetPaperQuestions(ctx, req.PaperHash)
}

// HandleGetPaperQuestion handles getting a single question
func (c *Question) HandleGetPaperQuestion(ctx context.Context, req *proto.QuestionRequest) (*proto.QuestionResponse, error) {
	return c.questionSvc.GetPaperQuestion(ctx)
}

// HandleCreateQuestion handles creating a new question.
func (c *Question) HandleCreateQuestion(ctx context.Context, req *proto.CreateQuestionRequest) (*proto.CreateQuestionResponse, error) {
	return c.questionSvc.CreateQuestion(ctx, req)
}

// HandleUpdateQuestion handles question updates
func (c *Question) HandleUpdateQuestion(ctx context.Context, req *proto.UpdateQuestionRequest) (*proto.UpdateQuestionResponse, error) {
	return services.UpdateQuestion(ctx, req)
}

func (c *Question) HandleDeleteQuestion(ctx context.Context, req *proto.QuestionRequest) (*emptypb.Empty, error) {
	if err := c.questionSvc.DeleteQuestion(ctx, req.QuestionHash); err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}

// HandleReorderQuestions handles reordering questions in a category
func (c *Question) HandleReorderQuestions(ctx context.Context, req *proto.ReorderQuestionsRequest) (*emptypb.Empty, error) {
	if err := c.questionSvc.ReorderQuestions(ctx, req.CategoryId, req.QuestionHashes); err != nil {
		return nil, grpcerror.Internal(err, "failed to reorder questions")
	}
	return &emptypb.Empty{}, nil
}

// _____________________EXAM QUESTION HANDLERS_____________________

// HandleGetQuestionsByIds handles getting questions by their IDs for exam service
func (c *Question) HandleGetQuestionsByIds(ctx context.Context, req *proto.GetQuestionsByIdsRequest) (*proto.GetQuestionsByIdsResponse, error) {
	return c.questionSvc.GetQuestionsByIds(ctx, req.QuestionIds)
}

// HandleGetExamQuestion handles getting a question for exam taking
func (c *Question) HandleGetExamQuestionByHash(ctx context.Context, req *proto.QuestionRequest) (*proto.ExamQuestionResponse, error) {
	return c.questionSvc.GetExamQuestionByHash(ctx, req.QuestionHash)
}

func (c *Question) HandleGetQuestionHashes(ctx context.Context, req *proto.GetQuestionHashesRequest) (*proto.GetQuestionHashesResponse, error) {
	return c.questionSvc.GetQuestionHashes(ctx, req.QuestionIds)
}

func (c *Question) HandleGetQuestionIds(ctx context.Context, req *proto.GetQuestionIdsRequest) (*proto.GetQuestionIdsResponse, error) {
	return c.questionSvc.GetQuestionIds(ctx, req.QuestionHashes)
}

// HandleGetCodingQuestionInputDefinitions handles fetching input definitions for a coding question.
func (c *Question) HandleGetCodingQuestionInputDefinitions(ctx context.Context, req *proto.GetCodingQuestionInputDefinitionsRequest) (*proto.GetCodingQuestionInputDefinitionsResponse, error) {
	return c.questionSvc.GetCodingQuestionInputDefinitions(ctx, req.QuestionHash)
}
