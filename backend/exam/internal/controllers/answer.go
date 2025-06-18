package controllers

import (
	"context"

	"pariksha/common/pkg/proto"
	"pariksha/exam/internal/services"
)

type Answer struct {
	answerSvc *services.Answer
}

func NewAnswer(s *services.Answer) *Answer {
	return &Answer{
		answerSvc: s,
	}
}

func (c *Answer) HandleGetParticipantAnswers(ctx context.Context, req *proto.ParticipantRequest) (*proto.AnswerList, error) {
	return c.answerSvc.GetParticipantAnswers(ctx, req)
}

func (c *Answer) HandleGetAnswerForExam(ctx context.Context, req *proto.GetAnswerRequest) (*proto.AnswerMinimalResponse, error) {
	return c.answerSvc.GetAnswerForExam(ctx, req)
}

func (c *Answer) HandleUpsertAnswer(ctx context.Context, req *proto.UpsertAnswersRequest) (*proto.UpsertAnswersResponse, error) {
	return c.answerSvc.UpsertAnswer(ctx, req)
}
