package controllers

import (
	"context"

	"google.golang.org/protobuf/types/known/emptypb"

	"pariksha/common/pkg/proto"
	"pariksha/exam/internal/services"
)

type Exam struct {
	examSvc *services.Exam
}

func NewExam(s *services.Exam) *Exam {
	return &Exam{
		examSvc: s,
	}
}

// HandleGetUserExams handles retrieving all exams for a user
func (c *Exam) HandleGetUserExams(ctx context.Context, req *emptypb.Empty) (*proto.ExamList, error) {
	return c.examSvc.GetUserExams(ctx)
}

// HandleCreateExam handles creating a new exam
func (c *Exam) HandleCreateExam(ctx context.Context, req *proto.CreateExamRequest) (*proto.ExamResponse, error) {
	return c.examSvc.CreateExam(ctx, req)
}

// HandleUpdateExam handles updating an exam's details
func (c *Exam) HandleUpdateExam(ctx context.Context, req *proto.UpdateExamRequest) (*proto.ExamResponse, error) {
	return c.examSvc.UpdateExam(ctx, req)
}

// HandleStartExam handles starting an exam for a participant
func (c *Exam) HandleStartExam(ctx context.Context, req *proto.StartExamRequest) (*emptypb.Empty, error) {
	return c.examSvc.StartExam(ctx, req)
}

// HandleEndExam handles ending an exam for a participant
func (c *Exam) HandleEndExam(ctx context.Context, req *proto.EndExamRequest) (*emptypb.Empty, error) {
	return c.examSvc.EndExam(ctx, req)
}

// HandleGetExam handles retrieving exam details
func (c *Exam) HandleGetExam(ctx context.Context, req *proto.ExamRequest) (*proto.ExamResponse, error) {
	return c.examSvc.GetExam(ctx, req)
}

// HandleGetExamPermission handles retrieving exam permissions
func (c *Exam) HandleGetExamPermission(ctx context.Context, req *proto.ExamRequest) (*proto.ExamPermissionResponse, error) {
	return c.examSvc.GetExamPermission(ctx, req)
}

// HandleDeleteExams handles deleting exams
func (c *Exam) HandleDeleteExams(ctx context.Context, req *proto.DeleteExamsRequest) (*emptypb.Empty, error) {
	return c.examSvc.DeleteExams(ctx, req)
}
