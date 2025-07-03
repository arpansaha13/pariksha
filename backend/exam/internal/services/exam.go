package services

import (
	"context"
	"database/sql"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"gorm.io/gorm"

	"pariksha/common/pkg/constants"
	"pariksha/common/pkg/models"
	"pariksha/common/pkg/proto"
	"pariksha/common/pkg/structs"
	"pariksha/common/pkg/types"
	"pariksha/common/pkg/utils"
	"pariksha/common/pkg/utils/generate"
	"pariksha/common/pkg/utils/ptr"
	"pariksha/exam/internal/config/db"
	"pariksha/exam/internal/interceptors"
	"pariksha/exam/internal/interservice"
	"pariksha/exam/internal/repositories"
)

type Exam struct {
	examRepo        *repositories.Exam
	participantRepo *repositories.Participant
	permissionRepo  *repositories.Permission
}

func NewExam(examRepo *repositories.Exam, participantRepo *repositories.Participant, permissionRepo *repositories.Permission) *Exam {
	return &Exam{
		examRepo:        examRepo,
		participantRepo: participantRepo,
		permissionRepo:  permissionRepo,
	}
}

// GetUserExams retrieves all exams created by or participated in by the authenticated user
func (s *Exam) GetUserExams(ctx context.Context) (*proto.ExamList, error) {
	userID, err := utils.GetUserIDFromMetadata(ctx)
	if err != nil {
		return nil, err
	}

	exams, err := s.examRepo.GetByUserID(nil, userID)
	if err != nil {
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
func (s *Exam) CreateExam(ctx context.Context, req *proto.CreateExamRequest) (*proto.ExamResponse, error) {
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

	var exam models.Exam
	err = s.examRepo.Transaction(func(tx *gorm.DB) error {
		exam = models.Exam{
			Title:              req.Title,
			StartsAt:           startsAt,
			EndsAt:             endsAt,
			CreatedBy:          userID,
			MaxCandidatesCount: req.MaxCandidatesCount,
			PaperHash:          req.PaperHash,
			DurationMinutes:    int16(req.DurationMinutes),
		}

		// Only set Type if it's not LINK - will use database default
		if req.Type != nil && req.GetType() != constants.EXAM_ACCESS_TYPE_LINK {
			if req.GetType() != constants.EXAM_ACCESS_TYPE_INVITE {
				return status.Error(codes.InvalidArgument, "exam type must be either LINK or INVITE")
			}
			exam.Type = req.GetType()
		}

		if err := s.examRepo.Create(tx, &exam); err != nil {
			return status.Error(codes.Internal, "failed to create exam")
		}

		// Generate and store hash
		exam.Hash = generate.HMACHash(int64(exam.ID))
		if err := s.examRepo.UpdateHash(tx, &exam); err != nil {
			return status.Error(codes.Internal, "failed to store exam hash")
		}

		ownerPerm := repositories.PermissionFlags{
			Write:    true,
			Evaluate: true,
		}
		if err := s.permissionRepo.Create(tx, exam.ID, userID, &ownerPerm); err != nil {
			return status.Error(codes.Internal, "failed to create exam permissions")
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	interservice.EnqueuePrepareQuestions(structs.PrepareQuestionsPayload{
		ExamID:    exam.ID,
		PaperHash: exam.PaperHash,
	})

	return examToProto(&exam)
}

// UpdateExam modifies an existing exam's details while enforcing time-based restrictions
func (s *Exam) UpdateExam(ctx context.Context, req *proto.UpdateExamRequest) (*proto.ExamResponse, error) {
	exam, ok := interceptors.GetExamFromContext(ctx)
	if !ok {
		return nil, status.Error(codes.Internal, "exam not found in context")
	}

	updateCtx := &examUpdateCtx{
		exam:      exam,
		req:       req,
		now:       time.Now(),
		isUpdated: false,
	}

	// Validate update request
	if err := validateExamUpdate(updateCtx); err != nil {
		return nil, err
	}

	// Apply updates
	if err := updateExamFields(updateCtx); err != nil {
		return nil, err
	}

	// Save if any fields were updated
	if updateCtx.isUpdated {
		if err := s.examRepo.Save(nil, exam); err != nil {
			return nil, status.Error(codes.Internal, "failed to update exam")
		}
	}

	return examToProto(exam)
}

// StartExam initiates an exam for a participant and updates participation statistics
func (s *Exam) StartExam(ctx context.Context, _ *proto.StartExamRequest) (*emptypb.Empty, error) {
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

	typedExamId := types.ExamID(exam.ID)
	var participant *models.ExamParticipant
	err = s.examRepo.Transaction(func(tx *gorm.DB) error {
		participant, err = s.participantRepo.GetByExamAndUser(tx, typedExamId, userID)
		participantExists := err == nil

		if !participantExists {
			if exam.Type != constants.EXAM_ACCESS_TYPE_LINK {
				return status.Error(codes.PermissionDenied, "participant is not invited")
			}
			// Create participant with started status for LINK type exams
			scheduledEndTime := now.Add(time.Duration(exam.DurationMinutes) * time.Minute)
			participant = &models.ExamParticipant{
				ExamID:           typedExamId,
				UserID:           userID,
				Status:           constants.PARTICIPANT_STATUS_STARTED,
				StartedAt:        sql.NullTime{Time: now, Valid: true},
				ScheduledEndTime: sql.NullTime{Time: scheduledEndTime, Valid: true},
			}
			if err := s.participantRepo.Create(tx, participant); err != nil {
				return status.Error(codes.Internal, "failed to create participant")
			}

			// Create permissions for the new participant
			participantPerm := repositories.PermissionFlags{Participate: true}
			if err := s.permissionRepo.Create(tx, typedExamId, userID, &participantPerm); err != nil {
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
			if err := s.participantRepo.Save(tx, participant); err != nil {
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

		if err := s.examRepo.Save(tx, exam); err != nil {
			return status.Error(codes.Internal, "failed to update exam")
		}

		// After successfully creating/updating participant
		// Add delayed task for auto-ending exam
		autoEndPayload := structs.AutoEndExamPayload{
			ExamID:        typedExamId,
			ParticipantID: participant.ID,
		}
		interservice.EnqueueAutoEndExam(autoEndPayload, participant.ScheduledEndTime.Time)

		return nil
	})

	if err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

// EndExam marks a participant's exam as complete and updates participation statistics
func (s *Exam) EndExam(ctx context.Context, req *proto.EndExamRequest) (*emptypb.Empty, error) {
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
		return &emptypb.Empty{}, nil
	}

	err := s.examRepo.Transaction(func(tx *gorm.DB) error {
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

	return &emptypb.Empty{}, nil
}

// GetExam retrieves detailed information about a specific exam
func (s *Exam) GetExam(ctx context.Context, req *proto.ExamRequest) (*proto.ExamResponse, error) {
	exam, ok := interceptors.GetExamFromContext(ctx)
	if !ok {
		return nil, status.Error(codes.Internal, "exam not found in context")
	}

	return examToProto(exam)
}

func (s *Exam) GetExamPermission(ctx context.Context, req *proto.ExamRequest) (*proto.ExamPermissionResponse, error) {
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
		err = db.DB.Where("exam_id = ? AND user_id = ?", exam.ID, userID).Take(&participant).Error
		if err == nil {
			response.ParticipantStatus = ptr.Int32(int32(participant.Status))
		}
	}

	return response, nil
}

// DeleteExams handles the batch deletion of exams and their associated permissions
func (s *Exam) DeleteExams(ctx context.Context, req *proto.DeleteExamsRequest) (*emptypb.Empty, error) {
	err := s.examRepo.Transaction(func(tx *gorm.DB) error {
		examIDs, err := s.examRepo.DeleteByHashes(tx, req.ExamHashes)
		if err != nil {
			return status.Error(codes.Internal, "failed to fetch/delete exam IDs")
		}

		if err := s.permissionRepo.DeleteByExamIDs(tx, examIDs); err != nil {
			return status.Error(codes.Internal, "failed to delete exam permissions")
		}

		if err := interservice.EnqueuePostDeleteExamsCleanup(examIDs); err != nil {
			return status.Error(codes.Internal, "failed to enqueue post-delete-exam cleanup task")
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}
