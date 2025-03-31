package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"strings"
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
	if err := db.DB.Preload("User").Where("exam_id = ?", req.ExamId).Find(&participants).Error; err != nil {
		return nil, status.Error(codes.Internal, "failed to fetch participants")
	}

	response := &proto.ParticipantList{
		Participants: make([]*proto.ParticipantResponse, len(participants)),
	}

	for i, p := range participants {
		response.Participants[i] = &proto.ParticipantResponse{
			Id:           int32(p.ID),
			UserId:       int32(p.UserID),
			FirstName:    p.User.FirstName.String,
			LastName:     p.User.LastName.String,
			Email:        p.User.Email,
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

func (s *ExamServer) AddExamParticipants(ctx context.Context, req *proto.AddParticipantsRequest) (*proto.AddParticipantsResponse, error) {
	var exam models.Exam
	if err := db.DB.Take(&exam, req.ExamId).Error; err != nil {
		return nil, status.Error(codes.NotFound, "exam not found")
	}

	if exam.Type == constants.EXAM_TYPE_OPEN {
		return nil, status.Error(codes.InvalidArgument, "participants cannot be added in OPEN exams")
	}

	counts, err := exam.GetParticipantCounts()
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to get participant counts")
	}

	var examParticipants []models.ExamParticipant
	addedCount := 0
	omittedCount := 0
	maxLimitReached := false

	for _, p := range req.Participants {
		currTotalParticipants := counts.Invited + counts.Started + counts.Ended
		if currTotalParticipants == exam.MaxCandidatesCount {
			maxLimitReached = true
			omittedCount++
			continue
		}

		var userID int32 = p.UserId
		if userID == 0 && p.Email != "" {
			// Create unverified user
			username := strings.Split(p.Email, "@")[0]
			user := models.User{
				Email:    p.Email,
				Username: username,
			}

			if p.FirstName != "" {
				user.FirstName = sql.NullString{String: p.FirstName, Valid: true}
			}
			if p.LastName != "" {
				user.LastName = sql.NullString{String: p.LastName, Valid: true}
			}

			if err := db.DB.Create(&user).Error; err != nil {
				return nil, status.Error(codes.Internal, "failed to create user")
			}
			userID = int32(user.ID)
		}

		participant := models.ExamParticipant{
			ExamID: int(req.ExamId),
			UserID: int(userID),
		}

		examParticipants = append(examParticipants, participant)
		counts.Invited++
		addedCount++
	}

	if len(examParticipants) > 0 {
		if err := db.DB.Create(&examParticipants).Error; err != nil {
			return nil, status.Error(codes.Internal, "failed to add participants")
		}

		exam.ParticipantCounts, err = json.Marshal(counts)
		if err != nil {
			return nil, status.Error(codes.Internal, "failed to marshal counts")
		}

		if err := db.DB.Save(&exam).Error; err != nil {
			return nil, status.Error(codes.Internal, "failed to update exam")
		}
	}

	response := &proto.AddParticipantsResponse{
		AddedCount:   int32(addedCount),
		OmittedCount: int32(omittedCount),
	}
	if maxLimitReached {
		maxLimitReason := "Maximum participant limit reached for the exam"
		response.MaxLimitReason = &maxLimitReason
	}

	return response, nil
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
