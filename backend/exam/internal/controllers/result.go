package controllers

import (
	"context"

	"pariksha/common/pkg/proto"
	"pariksha/exam/internal/services"
)

type Result struct {
	resultSvc *services.Result
}

func NewResult(s *services.Result) *Result {
	return &Result{
		resultSvc: s,
	}
}

func (c *Result) HandleGetExamResults(ctx context.Context, req *proto.ExamRequest) (*proto.ExamResultsResponse, error) {
	return c.resultSvc.GetExamResults(ctx, req)
}
