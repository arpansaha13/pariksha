package controllers

import (
	"context"

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

func (c *Question) HandleDeleteQuestion(ctx context.Context, req *proto.QuestionRequest) (*proto.Empty, error) {
	if err := c.questionSvc.DeleteQuestion(ctx, req.QuestionHash); err != nil {
		return nil, err
	}
	return &proto.Empty{}, nil
}

// HandleReorderQuestions handles reordering questions in a category
func (c *Question) HandleReorderQuestions(ctx context.Context, req *proto.ReorderQuestionsRequest) (*proto.Empty, error) {
	if err := c.questionSvc.ReorderQuestions(ctx, req.CategoryId, req.QuestionHashes); err != nil {
		return nil, grpcerror.Internal(err, "failed to reorder questions")
	}
	return &proto.Empty{}, nil
}

// _____________________EXAM QUESTION HANDLERS_____________________

// HandleGetQuestionsByIds handles getting questions by their IDs for exam service
func (c *Question) HandleGetQuestionsByIds(ctx context.Context, req *proto.GetQuestionsByIdsRequest) (*proto.GetQuestionsByIdsResponse, error) {
	return c.questionSvc.GetQuestionsByIds(ctx, req.QuestionIds)
}

// HandleGetExamQuestion handles getting a question for exam taking
func (c *Question) HandleGetExamQuestion(ctx context.Context, req *proto.QuestionRequest) (*proto.QuestionResponse, error) {
	return c.questionSvc.GetExamQuestion(ctx, req.QuestionHash)
}

func (c *Question) HandleGetQuestionHashes(ctx context.Context, req *proto.GetQuestionHashesRequest) (*proto.GetQuestionHashesResponse, error) {
	return c.questionSvc.GetQuestionHashes(ctx, req.QuestionIds)
}

func (c *Question) HandleGetQuestionIds(ctx context.Context, req *proto.GetQuestionIdsRequest) (*proto.GetQuestionIdsResponse, error) {
	return c.questionSvc.GetQuestionIds(ctx, req.QuestionHashes)
}
