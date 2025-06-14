package handlers

import (
	"context"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
	"gorm.io/gorm"

	"pariksha/common/pkg/constants"
	"pariksha/common/pkg/models"
	"pariksha/common/pkg/proto"
	"pariksha/common/pkg/types"
	"pariksha/common/pkg/utils"
	"pariksha/exam/internal/config/db"
	"pariksha/exam/internal/interceptors"
)

func (s *ExamServer) GetExamParticipants(ctx context.Context, req *proto.ExamRequest) (*proto.ParticipantList, error) {
	examID, ok := interceptors.GetExamIDFromContext(ctx)
	if !ok {
		return nil, status.Error(codes.Internal, "exam ID not found in context")
	}

	var participants []models.ExamParticipant
	if err := db.DB.Where("exam_id = ?", examID).Find(&participants).Error; err != nil {
		return nil, status.Error(codes.Internal, "failed to fetch participants")
	}

	response := &proto.ParticipantList{
		Participants: make([]*proto.ParticipantResponse, len(participants)),
	}

	for i, p := range participants {
		response.Participants[i] = &proto.ParticipantResponse{
			ParticipantId: int64(p.ID),
			UserId:        int64(p.UserID),
			Status:        int32(p.Status),
			ScoreAwarded:  int32(p.ScoreAwarded),
		}
		if p.StartedAt.Valid {
			response.Participants[i].StartedAt = timestamppb.New(p.StartedAt.Time)
		}
		if p.EndedAt.Valid {
			response.Participants[i].EndedAt = timestamppb.New(p.EndedAt.Time)
		}
		if p.ScheduledEndTime.Valid {
			response.Participants[i].ScheduledEndTime = timestamppb.New(p.ScheduledEndTime.Time)
		}
	}

	return response, nil
}

func (s *ExamServer) AddExamParticipant(ctx context.Context, req *proto.AddParticipantRequest) (*proto.ParticipantResponse, error) {
	exam, ok := interceptors.GetExamFromContext(ctx)
	if !ok {
		return nil, status.Error(codes.Internal, "exam not found in context")
	}

	if exam.Type == constants.EXAM_ACCESS_TYPE_LINK {
		return nil, status.Error(codes.InvalidArgument, "participants cannot be added in exams with access-type LINK")
	}

	counts, err := exam.GetParticipantCounts()
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to get participant counts")
	}

	currTotalParticipants := int32(counts.Invited + counts.Started + counts.Ended)
	if currTotalParticipants >= exam.MaxCandidatesCount {
		return nil, status.Error(codes.FailedPrecondition, "maximum participant limit reached for the exam")
	}

	typedExamId := types.ExamID(exam.ID)
	typedUserId := types.UserID(req.UserId)

	participant := models.ExamParticipant{
		ExamID: typedExamId,
		UserID: typedUserId,
	}

	err = utils.TransactionHandler(db.DB, func(tx *gorm.DB) error {
		if err := tx.Create(&participant).Error; err != nil {
			return status.Error(codes.Internal, "failed to add participant")
		}

		permission := models.ExamPermission{
			ExamID: typedExamId,
			UserID: typedUserId,
		}
		permission.SetParticipate()

		if err := tx.Create(&permission).Error; err != nil {
			return status.Error(codes.Internal, "failed to create participant permissions")
		}

		exam.ParticipantCounts, err = updateParticipantCounts(&counts, 0, constants.PARTICIPANT_STATUS_INVITED)
		if err != nil {
			return err
		}

		return tx.Save(&exam).Error
	})

	if err != nil {
		return nil, status.Error(codes.Internal, "failed to add exam participant")
	}

	return &proto.ParticipantResponse{
		ParticipantId: int64(participant.ID),
		UserId:        int64(participant.UserID),
		Status:        int32(participant.Status),
		ScoreAwarded:  int32(participant.ScoreAwarded),
	}, nil
}

func (s *ExamServer) RemoveExamParticipant(ctx context.Context, req *proto.RemoveParticipantRequest) (*proto.Empty, error) {
	exam, ok := interceptors.GetExamFromContext(ctx)
	if !ok {
		return nil, status.Error(codes.Internal, "exam not found in context")
	}

	if exam.StartsAt.Before(time.Now()) {
		return nil, status.Error(codes.FailedPrecondition, "cannot remove participant after exam has started")
	}

	var transactionErr error
	err := utils.TransactionHandler(db.DB, func(tx *gorm.DB) error {
		counts, err := exam.GetParticipantCounts()
		if err != nil {
			return status.Error(codes.Internal, "failed to get participant counts")
		}

		var participant models.ExamParticipant
		if err := tx.Take(&participant, req.ParticipantId).Error; err != nil {
			transactionErr = status.Error(codes.NotFound, "participant not found")
			return err
		}

		// Update counts based on participant's status
		exam.ParticipantCounts, err = updateParticipantCounts(&counts, participant.Status, 0)
		if err != nil {
			return err
		}

		if err := tx.Save(&exam).Error; err != nil {
			return err
		}

		// Delete the participant's permissions
		if err := tx.Where("exam_id = ? AND user_id = ?", exam.ID, participant.UserID).Delete(&models.ExamPermission{}).Error; err != nil {
			return err
		}

		if err := tx.Delete(&participant).Error; err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		if transactionErr != nil {
			return nil, transactionErr
		}
		return nil, err
	}

	return &proto.Empty{}, nil
}

// GetExamParticipant returns participant data for the current user
func (s *ExamServer) GetExamParticipant(ctx context.Context, req *proto.GetExamParticipantRequest) (*proto.GetExamParticipantResponse, error) {
	userID, err := utils.GetUserIDFromMetadata(ctx)
	if err != nil {
		return nil, err
	}

	examID, ok := interceptors.GetExamIDFromContext(ctx)
	if !ok {
		return nil, status.Error(codes.Internal, "exam ID not found in context")
	}

	var participant models.ExamParticipant
	if err := db.DB.Where("exam_id = ? AND user_id = ?", examID, userID).Take(&participant).Error; err != nil {
		return nil, utils.HandleDBError(err, "participant not found")
	}

	response := &proto.GetExamParticipantResponse{
		ParticipantId: int64(participant.ID),
		ScoreAwarded:  participant.ScoreAwarded,
	}

	if participant.StartedAt.Valid {
		response.StartedAt = timestamppb.New(participant.StartedAt.Time)
	}

	if participant.ScheduledEndTime.Valid {
		response.ScheduledEndTime = timestamppb.New(participant.ScheduledEndTime.Time)
	}

	return response, nil
}

func (s *ExamServer) GetParticipantById(ctx context.Context, req *proto.ParticipantRequest) (*proto.ParticipantResponse, error) {
	var participant models.ExamParticipant
	if err := db.DB.Take(&participant, req.ParticipantId).Error; err != nil {
		return nil, utils.HandleDBError(err, "participant not found")
	}

	return &proto.ParticipantResponse{
		ParticipantId: int64(participant.ID),
		Status:        int32(participant.Status),
	}, nil
}
