package handlers

import (
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	"pariksha/common/pkg/models"
	"pariksha/common/pkg/proto"
)

func createExamResponse(exam *models.Exam) (*proto.ExamResponse, error) {
	counts, err := exam.GetParticipantCounts()
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to parse participant counts")
	}

	return &proto.ExamResponse{
		Id:                 int32(exam.ID),
		Title:              exam.Title,
		StartsAt:           timestamppb.New(exam.StartsAt),
		EndsAt:             timestamppb.New(exam.EndsAt),
		CreatedBy:          int32(exam.CreatedBy),
		Type:               exam.Type,
		MaxCandidatesCount: int32(exam.MaxCandidatesCount),
		PaperId:            int32(exam.PaperID),
		ParticipantCounts: &proto.ParticipantCount{
			Unattended: int32(counts.Unattended),
			Invited:    int32(counts.Invited),
			Started:    int32(counts.Started),
			Ended:      int32(counts.Ended),
		},
	}, nil
}
