package controllers

import (
	"context"

	"pariksha/common/pkg/proto"
	"pariksha/exam/internal/services"
)

type Evaluation struct {
	evaluationSvc *services.Evaluation
}

func NewEvaluation(s *services.Evaluation) *Evaluation {
	return &Evaluation{
		evaluationSvc: s,
	}
}

func (c *Evaluation) HandleGetAnswerForEvaluation(ctx context.Context, req *proto.ParticipantQuestionRequest) (*proto.AnswerMinimalResponse, error) {
	return c.evaluationSvc.GetAnswerForEvaluation(ctx, req)
}

func (c *Evaluation) HandleGetAnswerEvaluationData(ctx context.Context, req *proto.ParticipantQuestionRequest) (*proto.GetAnswerEvaluationDataResponse, error) {
	return c.evaluationSvc.GetAnswerEvaluationData(ctx, req)
}

func (c *Evaluation) HandleUpdateAnswerForEvaluation(ctx context.Context, req *proto.UpdateAnswerRequest) (*proto.GetAnswerEvaluationDataResponse, error) {
	return c.evaluationSvc.UpdateAnswerForEvaluation(ctx, req)
}

func (c *Evaluation) HandleMarkParticipantAsEvaluated(ctx context.Context, req *proto.ParticipantRequest) (*proto.EvaluationStatusResponse, error) {
	return c.evaluationSvc.MarkParticipantAsEvaluated(ctx, req)
}
