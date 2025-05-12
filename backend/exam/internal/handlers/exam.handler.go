package handlers

import (
	"context"
	"database/sql"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"

	"pariksha/common/pkg/constants"
	"pariksha/common/pkg/models"
	"pariksha/common/pkg/proto"
	"pariksha/common/pkg/structs"
	"pariksha/common/pkg/utils"
	"pariksha/common/pkg/utils/ptr"
	"pariksha/exam/internal/config/db"
	"pariksha/exam/internal/interceptors"
	"pariksha/exam/internal/services"
)

type ExamServer struct {
	proto.UnimplementedExamServiceServer
}

// GetUserExams retrieves all exams created by the authenticated user
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
		response.Exams[i], err = examToProto(&exams[i])
		if err != nil {
			return nil, err
		}
	}

	return response, nil
}

// CreateExam creates a new exam with the specified configuration and validates time constraints
func (s *ExamServer) CreateExam(ctx context.Context, req *proto.CreateExamRequest) (*proto.ExamResponse, error) {
	userID, err := utils.GetUserIDFromMetadata(ctx)
	if err != nil {
		return nil, err
	}

	startsAt := req.StartsAt.AsTime()
	endsAt := req.EndsAt.AsTime()

	if err := validateExamStartTiming(startsAt); err != nil {
		return nil, err
	}

	if err := validateExamEndTiming(startsAt, endsAt); err != nil {
		return nil, err
	}

	if req.MaxCandidatesCount <= 0 {
		return nil, status.Error(codes.InvalidArgument, "max candidates count must be greater than zero")
	}

	if err := validateExamDuration(req.DurationMinutes); err != nil {
		return nil, err
	}

	exam := models.Exam{
		Title:              req.Title,
		StartsAt:           startsAt,
		EndsAt:             endsAt,
		CreatedBy:          userID,
		MaxCandidatesCount: req.MaxCandidatesCount,
		PaperID:            req.PaperId,
		DurationMinutes:    int16(req.DurationMinutes),
	}

	// Only set Type if it's not LINK - will use database default
	if req.Type != nil && req.GetType() != constants.EXAM_ACCESS_TYPE_LINK {
		if req.GetType() != constants.EXAM_ACCESS_TYPE_INVITE {
			return nil, status.Error(codes.InvalidArgument, "exam type must be either LINK or INVITE")
		}
		exam.Type = req.GetType()
	}

	err = utils.TransactionHandler(db.DB, func(tx *gorm.DB) error {
		if err := tx.Create(&exam).Error; err != nil {
			return status.Error(codes.Internal, "failed to create exam")
		}

		permission := models.ExamPermissions{
			ExamID: exam.ID,
			UserID: userID,
		}
		permission.SetWrite()
		permission.SetEvaluate()

		if err := tx.Create(&permission).Error; err != nil {
			return status.Error(codes.Internal, "failed to create exam permissions")
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	services.EnqueuePrepareQuestons(structs.PrepareQuestionsPayload{
		ExamID:  exam.ID,
		PaperID: exam.PaperID,
	})

	return examToProto(&exam)
}

// UpdateExam modifies an existing exam's details while enforcing time-based restrictions
func (s *ExamServer) UpdateExam(ctx context.Context, req *proto.UpdateExamRequest) (*proto.ExamResponse, error) {
	exam, ok := interceptors.GetExamFromContext(ctx)
	if !ok {
		return nil, status.Error(codes.Internal, "exam not found in context")
	}

	isUpdated := false
	now := time.Now()

	if now.After(exam.EndsAt) {
		// Allow only title updates after exam has ended
		if req.Title != nil && *req.Title != exam.Title {
			exam.Title = *req.Title
			if err := db.DB.Save(&exam).Error; err != nil {
				return nil, status.Error(codes.Internal, "failed to update exam")
			}
			return examToProto(exam)
		}
		return nil, status.Error(codes.FailedPrecondition, "cannot update exam after it has ended")
	}

	if req.Title != nil && *req.Title != exam.Title {
		exam.Title = *req.Title
		isUpdated = true
	}

	if req.StartsAt != nil && now.After(exam.StartsAt) {
		return nil, status.Error(codes.FailedPrecondition, "cannot update start time after exam has started")
	}

	if req.StartsAt != nil {
		startsAt := req.StartsAt.AsTime()
		if err := validateExamStartTiming(startsAt); err != nil {
			return nil, err
		}
		exam.StartsAt = startsAt
		isUpdated = true
	}

	if req.EndsAt != nil {
		endsAt := req.EndsAt.AsTime()
		if err := validateExamEndTiming(exam.StartsAt, endsAt); err != nil {
			return nil, err
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

	if req.DurationMinutes != nil {
		if err := validateExamDuration(req.GetDurationMinutes()); err != nil {
			return nil, err
		}
		exam.DurationMinutes = int16(req.GetDurationMinutes())
		isUpdated = true
	}

	if isUpdated {
		if err := db.DB.Save(&exam).Error; err != nil {
			return nil, status.Error(codes.Internal, "failed to update exam")
		}
	}

	return examToProto(exam)
}

// StartExam initiates an exam for a participant and updates participation statistics
func (s *ExamServer) StartExam(ctx context.Context, req *proto.StartExamRequest) (*proto.Empty, error) {
	userID, err := utils.GetUserIDFromMetadata(ctx)
	if err != nil {
		return nil, err
	}

	exam, ok := interceptors.GetExamFromContext(ctx)
	if !ok {
		return nil, status.Error(codes.Internal, "exam not found in context")
	}

	now := time.Now()
	if err := validateExamState(exam, now); err != nil {
		return nil, err
	}

	participant := &models.ExamParticipant{}
	err = utils.TransactionHandler(db.DB, func(tx *gorm.DB) error {
		err := tx.Where("exam_id = ? AND user_id = ?", req.ExamId, userID).Take(participant).Error
		participantExists := err == nil

		if !participantExists {
			if exam.Type != constants.EXAM_ACCESS_TYPE_LINK {
				return status.Error(codes.PermissionDenied, "participant is not invited")
			}
			// Create participant with started status for LINK type exams
			scheduledEndTime := now.Add(time.Duration(exam.DurationMinutes) * time.Minute)
			participant = &models.ExamParticipant{
				ExamID:           req.ExamId,
				UserID:           userID,
				Status:           constants.PARTICIPANT_STATUS_STARTED,
				StartedAt:        sql.NullTime{Time: now, Valid: true},
				ScheduledEndTime: sql.NullTime{Time: scheduledEndTime, Valid: true},
			}
			if err := tx.Create(&participant).Error; err != nil {
				return status.Error(codes.Internal, "failed to create participant")
			}

			// Create permissions for the new participant
			permission := models.ExamPermissions{
				ExamID: req.ExamId,
				UserID: userID,
			}
			permission.SetParticipate()
			if err := tx.Create(&permission).Error; err != nil {
				return status.Error(codes.Internal, "failed to create participant permissions")
			}
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

		if !participantExists {
			exam.ParticipantCounts, err = updateParticipantCounts(&counts, 0, constants.PARTICIPANT_STATUS_STARTED)
		} else {
			exam.ParticipantCounts, err = updateParticipantCounts(&counts, constants.PARTICIPANT_STATUS_INVITED, constants.PARTICIPANT_STATUS_STARTED)
		}
		if err != nil {
			return err
		}

		if err := tx.Save(&exam).Error; err != nil {
			return status.Error(codes.Internal, "failed to update exam")
		}

		// After successfully creating/updating participant
		// Add delayed task for auto-ending exam
		autoEndPayload := structs.AutoEndExamPayload{
			ExamID:        req.ExamId,
			ParticipantID: participant.ID,
		}
		services.EnqueueAutoEndExam(autoEndPayload, participant.ScheduledEndTime.Time)

		return nil
	})

	if err != nil {
		return nil, err
	}

	return &proto.Empty{}, nil
}

// EndExam marks a participant's exam as complete and updates participation statistics
func (s *ExamServer) EndExam(ctx context.Context, req *proto.EndExamRequest) (*proto.Empty, error) {
	exam, ok := interceptors.GetExamFromContext(ctx)
	if !ok {
		return nil, status.Error(codes.Internal, "exam not found in context")
	}

	// Should be added to context by EndExamInterceptor
	participant, ok := interceptors.GetParticipantFromContext(ctx)
	if !ok {
		return nil, status.Error(codes.Internal, "participant not found in context")
	}

	// Return success if exam is already ended
	if participant.Status == constants.PARTICIPANT_STATUS_ENDED {
		return &proto.Empty{}, nil
	}

	err := utils.TransactionHandler(db.DB, func(tx *gorm.DB) error {
		if time.Now().Before(exam.StartsAt) {
			return status.Error(codes.FailedPrecondition, "exam has not started yet")
		}

		participant.EndedAt = sql.NullTime{Time: time.Now(), Valid: true}
		return handleParticipantUpdate(tx, exam, participant,
			participant.Status, constants.PARTICIPANT_STATUS_ENDED)
	})

	if err != nil {
		return nil, err
	}

	return &proto.Empty{}, nil
}

// GetExam retrieves detailed information about a specific exam
func (s *ExamServer) GetExam(ctx context.Context, req *proto.ExamRequest) (*proto.ExamResponse, error) {
	exam, ok := interceptors.GetExamFromContext(ctx)
	if !ok {
		return nil, status.Error(codes.Internal, "exam not found in context")
	}

	return examToProto(exam)
}

// GetExamQuestions retrieves all questions associated with an exam
func (s *ExamServer) GetExamQuestions(ctx context.Context, req *proto.ExamRequest) (*proto.ExamQuestionsResponse, error) {
	var examQuestions []models.ExamQuestion
	if err := db.DB.Model(&models.ExamQuestion{}).
		Select("question_id", "category_id", "order", "max_score").
		Where("exam_id = ?", req.ExamId).
		Find(&examQuestions).Error; err != nil {
		return nil, status.Error(codes.Internal, "failed to fetch questions")
	}

	questions := make([]*proto.ExamQuestion, len(examQuestions))
	for i, eq := range examQuestions {
		questions[i] = &proto.ExamQuestion{
			QuestionId: eq.QuestionID,
			CategoryId: eq.CategoryID,
			Order:      int32(eq.Order),
			MaxScore:   int32(eq.MaxScore),
		}
	}

	return &proto.ExamQuestionsResponse{
		Questions: questions,
	}, nil
}

// GetExamCategories retrieves all category IDs associated with an exam
func (s *ExamServer) GetExamCategories(ctx context.Context, req *proto.ExamRequest) (*proto.ExamCategoriesResponse, error) {
	var examCategories []models.ExamCategory
	if err := db.DB.Model(&models.ExamCategory{}).
		Select("category_id", "order").
		Where("exam_id = ?", req.ExamId).
		Find(&examCategories).Error; err != nil {
		return nil, status.Error(codes.Internal, "failed to fetch categories")
	}

	categories := make([]*proto.ExamCategory, len(examCategories))
	for i, ec := range examCategories {
		categories[i] = &proto.ExamCategory{
			CategoryId: ec.CategoryID,
			Order:      int32(ec.Order),
		}
	}

	return &proto.ExamCategoriesResponse{
		Categories: categories,
	}, nil
}

func (s *ExamServer) GetExamPermission(ctx context.Context, req *proto.ExamRequest) (*proto.ExamPermissionResponse, error) {
	userID, err := utils.GetUserIDFromMetadata(ctx)
	if err != nil {
		return nil, err
	}

	exam, ok := interceptors.GetExamFromContext(ctx)
	if !ok {
		return nil, status.Error(codes.Internal, "exam not found in context")
	}

	permission, ok := interceptors.GetPermissionFromContext(ctx)
	// For LINK type exams, grant PARTICIPATE permission if no permission exists
	if !ok {
		if exam.Type == constants.EXAM_ACCESS_TYPE_LINK {
			return &proto.ExamPermissionResponse{
				CanRead:           true,
				CanWrite:          false,
				CanParticipate:    true,
				CanEvaluate:       false,
				ParticipantStatus: ptr.Int32(int32(constants.PARTICIPANT_STATUS_INVITED)),
			}, nil
		}
		return &proto.ExamPermissionResponse{}, nil
	}

	response := &proto.ExamPermissionResponse{
		CanRead:        permission.CanRead(),
		CanWrite:       permission.CanWrite(),
		CanParticipate: permission.CanParticipate(),
		CanEvaluate:    permission.CanEvaluate(),
	}

	// Check if user is a participant and add status if found
	if permission.CanParticipate() {
		var participant models.ExamParticipant
		err = db.DB.Where("exam_id = ? AND user_id = ?", req.ExamId, userID).Take(&participant).Error
		if err == nil {
			response.ParticipantStatus = ptr.Int32(int32(participant.Status))
		}
	}

	return response, nil
}

// DeleteExams handles the batch deletion of exams and their associated permissions
func (s *ExamServer) DeleteExams(ctx context.Context, req *proto.DeleteExamsRequest) (*proto.Empty, error) {
	err := utils.TransactionHandler(db.DB, func(tx *gorm.DB) error {
		if err := tx.Where("id IN ?", req.ExamIds).Delete(&models.Exam{}).Error; err != nil {
			return status.Error(codes.Internal, "failed to delete exams")
		}

		if err := tx.Where("exam_id IN ?", req.ExamIds).Delete(&models.ExamPermissions{}).Error; err != nil {
			return status.Error(codes.Internal, "failed to delete exam permissions")
		}

		if err := services.EnqueuePostDeleteExamsCleanup(req.ExamIds); err != nil {
			return status.Error(codes.Internal, "failed to enqueue post-delete-exam cleanup task")
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return &proto.Empty{}, nil
}
