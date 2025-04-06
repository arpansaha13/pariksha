package handlers

import (
	"context"
	"encoding/json"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
	"gorm.io/gorm"

	"pariksha/common/pkg/constants"
	"pariksha/common/pkg/models"
	"pariksha/common/pkg/proto"
	"pariksha/exam/internal/config/db"
)

func (s *ExamServer) GetExamParticipants(ctx context.Context, req *proto.ExamRequest) (*proto.ParticipantList, error) {
	var participants []models.ExamParticipant
	if err := db.DB.Where("exam_id = ?", req.ExamId).Find(&participants).Error; err != nil {
		return nil, status.Error(codes.Internal, "failed to fetch participants")
	}

	response := &proto.ParticipantList{
		Participants: make([]*proto.ParticipantResponse, len(participants)),
	}

	for i, p := range participants {
		response.Participants[i] = &proto.ParticipantResponse{
			Id:           int32(p.ID),
			UserId:       int32(p.UserID),
			Status:       int32(p.Status),
			ScoreAwarded: int32(p.ScoreAwarded),
		}
		if p.StartedAt.Valid {
			response.Participants[i].StartedAt = timestamppb.New(p.StartedAt.Time)
		}
		if p.EndedAt.Valid {
			response.Participants[i].EndedAt = timestamppb.New(p.EndedAt.Time)
		}
	}

	return response, nil
}

func (s *ExamServer) AddExamParticipant(ctx context.Context, req *proto.AddParticipantRequest) (*proto.ParticipantResponse, error) {
	var exam models.Exam
	if err := db.DB.Take(&exam, req.ExamId).Error; err != nil {
		return nil, status.Error(codes.NotFound, "exam not found")
	}

	if exam.Type == constants.EXAM_ACCESS_TYPE_LINK {
		return nil, status.Error(codes.InvalidArgument, "participants cannot be added in exams with access-type LINK")
	}

	counts, err := exam.GetParticipantCounts()
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to get participant counts")
	}

	currTotalParticipants := counts.Invited + counts.Started + counts.Ended
	if currTotalParticipants >= exam.MaxCandidatesCount {
		return nil, status.Error(codes.FailedPrecondition, "maximum participant limit reached for the exam")
	}

	participant := models.ExamParticipant{
		ExamID: int(req.ExamId),
		UserID: int(req.UserId),
	}

	if err := db.DB.Create(&participant).Error; err != nil {
		return nil, status.Error(codes.Internal, "failed to add participant")
	}

	counts.Invited++
	exam.ParticipantCounts, err = json.Marshal(counts)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to marshal counts")
	}

	if err := db.DB.Save(&exam).Error; err != nil {
		return nil, status.Error(codes.Internal, "failed to update exam")
	}

	return &proto.ParticipantResponse{
		Id:           int32(participant.ID),
		UserId:       int32(participant.UserID),
		Status:       int32(participant.Status),
		ScoreAwarded: int32(participant.ScoreAwarded),
	}, nil
}

func (s *ExamServer) RemoveExamParticipant(ctx context.Context, req *proto.RemoveParticipantRequest) (*proto.Empty, error) {
	err := db.DB.Transaction(func(tx *gorm.DB) error {
		var exam models.Exam
		if err := tx.Take(&exam, req.ExamId).Error; err != nil {
			return status.Error(codes.NotFound, "exam not found")
		}

		if exam.StartsAt.Before(time.Now()) {
			return status.Error(codes.FailedPrecondition, "cannot remove participant after exam has started")
		}

		counts, err := exam.GetParticipantCounts()
		if err != nil {
			return status.Error(codes.Internal, "failed to get participant counts")
		}

		var participant models.ExamParticipant
		if err := tx.Take(&participant, req.ParticipantId).Error; err != nil {
			return status.Error(codes.NotFound, "participant not found")
		}

		// Update counts based on participant's status
		switch participant.Status {
		case constants.PARTICIPANT_STATUS_INVITED:
			counts.Invited--
		case constants.PARTICIPANT_STATUS_STARTED:
			counts.Started--
		case constants.PARTICIPANT_STATUS_ENDED:
			counts.Ended--
		case constants.PARTICIPANT_STATUS_UNATTENDED:
			counts.Unattended--
		}

		exam.ParticipantCounts, err = json.Marshal(counts)
		if err != nil {
			return status.Error(codes.Internal, "failed to marshal counts")
		}

		if err := tx.Save(&exam).Error; err != nil {
			return err
		}
		if err := tx.Delete(&participant).Error; err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return &proto.Empty{}, nil
}
