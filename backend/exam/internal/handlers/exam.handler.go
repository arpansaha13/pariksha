package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"

	"pariksha/common/pkg/constants"
	"pariksha/common/pkg/models"
	"pariksha/common/pkg/proto"
	"pariksha/common/pkg/types"
	"pariksha/common/pkg/utils"
	"pariksha/exam/internal/config/db"
	"pariksha/exam/internal/services"
)

type ExamServer struct {
	proto.UnimplementedExamServiceServer
}

func (s *ExamServer) GetUserExams(ctx context.Context, _ *proto.Empty) (*proto.ExamList, error) {
	userID, err := utils.GetUserIDFromMetadata(ctx)
	if err != nil {
		return nil, err
	}

	var exams []models.Exam
	if err := db.DB.Where("created_by = ?", userID).Find(&exams).Error; err != nil {
		return nil, status.Error(codes.Internal, "failed to retrieve exams")
	}

	response := &proto.ExamList{
		Exams: make([]*proto.ExamResponse, len(exams)),
	}

	for i := range exams {
		response.Exams[i], err = createExamResponse(&exams[i])
		if err != nil {
			return nil, err
		}
	}

	return response, nil
}

func (s *ExamServer) CreateExam(ctx context.Context, req *proto.CreateExamRequest) (*proto.ExamResponse, error) {
	userID, err := utils.GetUserIDFromMetadata(ctx)
	if err != nil {
		return nil, err
	}

	startsAt := req.StartsAt.AsTime()
	endsAt := req.EndsAt.AsTime()

	// Time input is not implemented in frontend yet
	// Compare dates only by truncating to start of day
	now := time.Now().Truncate(24 * time.Hour)

	if startsAt.Before(now) {
		return nil, status.Error(codes.InvalidArgument, "start time cannot be in the past")
	}

	if endsAt.Before(startsAt) || endsAt.Equal(startsAt) {
		return nil, status.Error(codes.InvalidArgument, "end time must be after start time")
	}

	if req.MaxCandidatesCount <= 0 {
		return nil, status.Error(codes.InvalidArgument, "max candidates count must be greater than zero")
	}

	if req.DurationMinutes <= 0 {
		return nil, status.Error(codes.InvalidArgument, "duration minutes must be greater than zero")
	}

	exam := models.Exam{
		Title:              req.Title,
		StartsAt:           startsAt,
		EndsAt:             endsAt,
		CreatedBy:          userID,
		MaxCandidatesCount: req.MaxCandidatesCount,
		PaperID:            req.PaperId,
		DurationMinutes:    req.DurationMinutes,
	}

	// Only set Type if it's not LINK
	if req.Type != nil && req.GetType() != constants.EXAM_ACCESS_TYPE_LINK {
		if req.GetType() != constants.EXAM_ACCESS_TYPE_INVITE {
			return nil, status.Error(codes.InvalidArgument, "exam type must be either LINK or INVITE")
		}
		exam.Type = req.GetType()
	}

	if err := db.DB.Create(&exam).Error; err != nil {
		return nil, status.Error(codes.Internal, "failed to create exam")
	}

	services.PushToExamQueue(types.ExamQueuePayload{
		ExamID:  exam.ID,
		PaperID: exam.PaperID,
	})

	return createExamResponse(&exam)
}

func (s *ExamServer) UpdateExam(ctx context.Context, req *proto.UpdateExamRequest) (*proto.ExamResponse, error) {
	userID, err := utils.GetUserIDFromMetadata(ctx)
	if err != nil {
		return nil, err
	}

	var exam models.Exam
	if err := db.DB.Take(&exam, req.ExamId).Error; err != nil {
		return nil, status.Error(codes.NotFound, "exam not found")
	}

	// Verify ownership
	if exam.CreatedBy != userID {
		return nil, status.Error(codes.PermissionDenied, "not authorized to update exam")
	}

	isUpdated := false
	now := time.Now()

	if req.Title != nil && *req.Title != exam.Title {
		exam.Title = *req.Title
		isUpdated = true
	}

	if now.After(exam.EndsAt) {
		return nil, status.Error(codes.FailedPrecondition, "cannot update exam after it has ended")
	}

	if req.StartsAt != nil {
		startsAt := req.StartsAt.AsTime()
		if now.After(exam.StartsAt) {
			return nil, status.Error(codes.FailedPrecondition, "cannot update start time after exam has started")
		}
		if startsAt.Before(now) {
			return nil, status.Error(codes.InvalidArgument, "start time cannot be in the past")
		}
		exam.StartsAt = startsAt
		isUpdated = true
	}

	if req.EndsAt != nil {
		endsAt := req.EndsAt.AsTime()
		if endsAt.Before(exam.StartsAt) || endsAt.Equal(exam.StartsAt) {
			return nil, status.Error(codes.InvalidArgument, "end time cannot be before or equal to start time")
		}
		exam.EndsAt = endsAt
		isUpdated = true
	}

	if req.Type != nil {
		if now.After(exam.StartsAt) {
			return nil, status.Error(codes.FailedPrecondition, "cannot update type after exam has started")
		}
		exam.Type = *req.Type
		isUpdated = true
	}

	if req.DurationMinutes != nil && req.GetDurationMinutes() > 0 {
		exam.DurationMinutes = req.GetDurationMinutes()
		isUpdated = true
	}

	if isUpdated {
		if err := db.DB.Save(&exam).Error; err != nil {
			return nil, status.Error(codes.Internal, "failed to update exam")
		}
	}

	return createExamResponse(&exam)
}

func (s *ExamServer) StartExam(ctx context.Context, req *proto.StartExamRequest) (*proto.Empty, error) {
	userID, err := utils.GetUserIDFromMetadata(ctx)
	if err != nil {
		return nil, err
	}

	err = db.DB.Transaction(func(tx *gorm.DB) error {
		var exam models.Exam
		if err := tx.Take(&exam, req.ExamId).Error; err != nil {
			return status.Error(codes.NotFound, "exam not found")
		}

		now := time.Now()

		// Check exam timing constraints
		if exam.StartsAt.After(now) {
			return status.Error(codes.FailedPrecondition, "exam has not started yet")
		}
		if exam.EndsAt.Before(now) {
			return status.Error(codes.FailedPrecondition, "exam has ended")
		}

		var participant models.ExamParticipant
		participantErr := tx.Where("exam_id = ? AND user_id = ?", req.ExamId, userID).Take(&participant).Error

		if participantErr == gorm.ErrRecordNotFound {
			if exam.Type != constants.EXAM_ACCESS_TYPE_LINK {
				return status.Error(codes.PermissionDenied, "participant is not invited")
			}
			// Create participant with started status for LINK type exams
			scheduledEndTime := now.Add(time.Duration(exam.DurationMinutes) * time.Minute)
			participant = models.ExamParticipant{
				ExamID:           req.ExamId,
				UserID:           userID,
				Status:           constants.PARTICIPANT_STATUS_STARTED,
				StartedAt:        sql.NullTime{Time: now, Valid: true},
				ScheduledEndTime: sql.NullTime{Time: scheduledEndTime, Valid: true},
			}
			if err := tx.Create(&participant).Error; err != nil {
				return status.Error(codes.Internal, "failed to create participant")
			}
		} else if participantErr != nil {
			return status.Error(codes.Internal, "database error")
		} else if participant.Status != constants.PARTICIPANT_STATUS_INVITED {
			return status.Error(codes.FailedPrecondition, "participant has already started the exam")
		} else {
			// Update existing participant
			scheduledEndTime := now.Add(time.Duration(exam.DurationMinutes) * time.Minute)
			participant.Status = constants.PARTICIPANT_STATUS_STARTED
			participant.StartedAt = sql.NullTime{Time: now, Valid: true}
			participant.ScheduledEndTime = sql.NullTime{Time: scheduledEndTime, Valid: true}
			if err := tx.Save(&participant).Error; err != nil {
				return status.Error(codes.Internal, "failed to update participant")
			}
		}

		counts, err := exam.GetParticipantCounts()
		if err != nil {
			return status.Error(codes.Internal, "failed to get participant counts")
		}

		if participantErr == gorm.ErrRecordNotFound {
			counts.Started++
		} else {
			counts.Invited--
			counts.Started++
		}

		exam.ParticipantCounts, err = json.Marshal(counts)
		if err != nil {
			return status.Error(codes.Internal, "failed to marshal counts")
		}

		if err := tx.Save(&exam).Error; err != nil {
			return status.Error(codes.Internal, "failed to update exam")
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return &proto.Empty{}, nil
}

func (s *ExamServer) EndExam(ctx context.Context, req *proto.EndExamRequest) (*proto.Empty, error) {
	userID, err := utils.GetUserIDFromMetadata(ctx)
	if err != nil {
		return nil, err
	}

	err = db.DB.Transaction(func(tx *gorm.DB) error {
		var participant models.ExamParticipant
		if err := tx.Where("exam_id = ? AND user_id = ?", req.ExamId, userID).Take(&participant).Error; err != nil {
			return status.Error(codes.NotFound, "participant not found")
		}

		var exam models.Exam
		if err := tx.Take(&exam, req.ExamId).Error; err != nil {
			return status.Error(codes.NotFound, "exam not found")
		}

		if time.Now().Before(exam.StartsAt) {
			return status.Error(codes.FailedPrecondition, "exam has not started yet")
		}

		counts, err := exam.GetParticipantCounts()
		if err != nil {
			return status.Error(codes.Internal, "failed to parse participant counts")
		}

		// Update counts
		if participant.Status == constants.PARTICIPANT_STATUS_STARTED {
			counts.Started--
			counts.Ended++
		}

		exam.ParticipantCounts, err = json.Marshal(counts)
		if err != nil {
			return status.Error(codes.Internal, "failed to marshal counts")
		}

		participant.Status = constants.PARTICIPANT_STATUS_ENDED
		participant.EndedAt = sql.NullTime{Time: time.Now(), Valid: true}

		if err := tx.Save(&exam).Error; err != nil {
			return status.Error(codes.Internal, "failed to update exam")
		}
		if err := tx.Save(&participant).Error; err != nil {
			return status.Error(codes.Internal, "failed to update participant")
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return &proto.Empty{}, nil
}

func (s *ExamServer) GetExam(ctx context.Context, req *proto.ExamRequest) (*proto.ExamResponse, error) {
	userID, err := utils.GetUserIDFromMetadata(ctx)
	if err != nil {
		return nil, err
	}

	var exam models.Exam
	if err := db.DB.Take(&exam, req.ExamId).Error; err != nil {
		return nil, status.Error(codes.NotFound, "exam not found")
	}

	// Check if user owns the exam
	if exam.CreatedBy == userID {
		return createExamResponse(&exam)
	}

	// If exam type is LINK, everyone has access
	if exam.Type == constants.EXAM_ACCESS_TYPE_LINK {
		return createExamResponse(&exam)
	}

	// For INVITE type exams, check if user is a participant
	var participant models.ExamParticipant
	err = db.DB.Where("exam_id = ? AND user_id = ?", req.ExamId, userID).Take(&participant).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, status.Error(codes.PermissionDenied, "not authorized to view exam")
		}
		return nil, status.Error(codes.Internal, "database error")
	}

	return createExamResponse(&exam)
}

func (s *ExamServer) CheckExamAccess(ctx context.Context, req *proto.ExamRequest) (*proto.ExamAccessResponse, error) {
	userID, err := utils.GetUserIDFromMetadata(ctx)
	if err != nil {
		return nil, err
	}

	var exam models.Exam
	if err := db.DB.Take(&exam, req.ExamId).Error; err != nil {
		return nil, status.Error(codes.NotFound, "exam not found")
	}

	// Check if user owns the exam
	if exam.CreatedBy == userID {
		return &proto.ExamAccessResponse{
			AccessType: proto.ExamAccessType_OWNER,
		}, nil
	}

	// If exam type is LINK, everyone has access as participant
	if exam.Type == constants.EXAM_ACCESS_TYPE_LINK {
		return &proto.ExamAccessResponse{
			AccessType: proto.ExamAccessType_PARTICIPANT,
		}, nil
	}

	// For INVITE type exams, check if user is a participant
	var participant models.ExamParticipant
	err = db.DB.Where("exam_id = ? AND user_id = ?", req.ExamId, userID).Take(&participant).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, status.Error(codes.PermissionDenied, "not authorized to access exam")
		}
		return nil, status.Error(codes.Internal, "database error")
	}

	return &proto.ExamAccessResponse{
		AccessType: proto.ExamAccessType_PARTICIPANT,
	}, nil
}
