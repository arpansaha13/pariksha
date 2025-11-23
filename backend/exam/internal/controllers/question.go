package controllers

import (
	"context"

	"pariksha/common/pkg/proto"
	"pariksha/exam/internal/services"
)

type Question struct {
	questionSvc *services.Question
}

func NewQuestion(s *services.Question) *Question {
	return &Question{
		questionSvc: s,
	}
}

// HandleGetExamQuestions handles retrieving exam questions
func (c *Question) HandleGetExamQuestions(ctx context.Context, req *proto.ExamRequest) (*proto.ExamQuestionsResponse, error) {
	return c.questionSvc.GetExamQuestions(ctx, req)
}

// HandleGetExamQuestion handles retrieving a single exam question
func (c *Question) HandleGetExamQuestion(ctx context.Context, req *proto.ExamQuestionRequest) (*proto.ExamQuestionResponse, error) {
	return c.questionSvc.GetExamQuestion(ctx, req)
}
