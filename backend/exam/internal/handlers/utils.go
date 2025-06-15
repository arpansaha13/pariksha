package handlers

import (
	"encoding/json"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
	"gorm.io/gorm"

	"pariksha/common/pkg/constants"
	"pariksha/common/pkg/models"
	"pariksha/common/pkg/proto"
)

// validateExamStartTiming checks if the exam's `startsAt` constraints are valid
func validateExamStartTiming(startsAt time.Time) error {
	// Time input is not implemented in frontend yet
	// Compare dates only by truncating to start of day
	now := time.Now().Truncate(24 * time.Hour)

	if startsAt.Before(now) {
		return status.Error(codes.InvalidArgument, "start time cannot be in the past")
	}
	return nil
}

// validateExamEndTiming checks if the exam's `endsAt` constraints are valid
func validateExamEndTiming(startsAt time.Time, endsAt time.Time) error {
	if endsAt.Before(startsAt) || endsAt.Equal(startsAt) {
		return status.Error(codes.InvalidArgument, "end time must be after start time")
	}
	return nil
}

// validateExamDuration checks if the exam duration is within allowed limits
func validateExamDuration(durationMinutes int32) error {
	if durationMinutes <= 0 {
		return status.Error(codes.InvalidArgument, "duration minutes must be greater than zero")
	}
	if durationMinutes > int32(constants.MAX_EXAM_DURATION_MINUTES) {
		return status.Error(codes.InvalidArgument, "duration minutes cannot exceed maximum allowed duration")
	}
	return nil
}

// updateParticipantCounts updates the exam's participant counts and returns marshaled counts
func updateParticipantCounts(counts *models.ParticipantCount, fromStatus int16, toStatus int16) (json.RawMessage, error) {
	switch fromStatus {
	case constants.PARTICIPANT_STATUS_INVITED:
		counts.Invited--
	case constants.PARTICIPANT_STATUS_STARTED:
		counts.Started--
	}

	switch toStatus {
	case constants.PARTICIPANT_STATUS_INVITED:
		counts.Invited++
	case constants.PARTICIPANT_STATUS_STARTED:
		counts.Started++
	case constants.PARTICIPANT_STATUS_ENDED:
		counts.Ended++
	}

	marshaled, err := json.Marshal(counts)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to marshal counts")
	}

	return marshaled, nil
}

func examToProto(exam *models.Exam) (*proto.ExamResponse, error) {
	counts, err := exam.GetParticipantCounts()
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to parse participant counts")
	}

	return &proto.ExamResponse{
		ExamHash:           exam.Hash,
		Title:              exam.Title,
		StartsAt:           timestamppb.New(exam.StartsAt),
		EndsAt:             timestamppb.New(exam.EndsAt),
		CreatedBy:          int64(exam.CreatedBy),
		Type:               exam.Type,
		MaxCandidatesCount: exam.MaxCandidatesCount,
		MaxScore:           exam.MaxScore,
		DurationMinutes:    int32(exam.DurationMinutes),
		ParticipantCounts: &proto.ParticipantCount{
			Unattended: int32(counts.Unattended),
			Invited:    int32(counts.Invited),
			Started:    int32(counts.Started),
			Ended:      int32(counts.Ended),
		},
	}, nil
}

// validateAnswerJSON validates the answer JSON based on question type
func validateAnswerJSON(answerJSON []byte, questionType proto.QuestionType) error {
	if len(answerJSON) == 0 {
		return nil // Allow empty answers
	}

	switch questionType {
	case proto.QuestionType_MCQ:
		var mcqAnswer models.MCQAnswer
		if err := json.Unmarshal(answerJSON, &mcqAnswer); err != nil {
			return status.Error(codes.InvalidArgument, "invalid MCQ answer format")
		}
		// Empty object or nil optionIndex is invalid for MCQ
		if mcqAnswer.OptionIndex == nil {
			return status.Error(codes.InvalidArgument, "option index is required for MCQ answers")
		}
		// Validate that optionIndex is non-negative
		if *mcqAnswer.OptionIndex < 0 {
			return status.Error(codes.InvalidArgument, "option index cannot be negative")
		}
	case proto.QuestionType_SUBJECTIVE:
		var textAnswer models.SubjectiveAnswer
		if err := json.Unmarshal(answerJSON, &textAnswer); err != nil {
			return status.Error(codes.InvalidArgument, "invalid text answer format")
		}
	default:
		return status.Error(codes.InvalidArgument, "invalid question type")
	}
	return nil
}

// getAnswerForParticipant finds an answer for a participant and question
func getAnswerForParticipant(db *gorm.DB, participantID int64, questionID int64) (*models.Answer, error) {
	var answer models.Answer
	err := db.Where("exam_participant_id = ? AND question_id = ? AND answer IS NOT NULL",
		participantID, questionID).Take(&answer).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, status.Error(codes.Internal, constants.ErrInternalServer)
	}
	return &answer, nil
}

// validateExamState checks common exam state constraints
func validateExamState(exam *models.Exam, now time.Time) error {
	if exam.StartsAt.After(now) {
		return status.Error(codes.FailedPrecondition, "exam has not started yet")
	}
	if exam.EndsAt.Before(now) {
		return status.Error(codes.FailedPrecondition, "exam has ended")
	}
	return nil
}

// handleParticipantUpdate performs the following steps to update a participant's status:
//
// 1. Gets current participant counts for the exam
//
// 2. Updates the counts based on the status change (fromStatus -> toStatus)
//
// 3. Saves the updated counts to the exam record
//
// 4. Updates the participant's status to toStatus
//
// 5. Saves the participant record
//
// All operations are performed within the provided transaction
func handleParticipantUpdate(tx *gorm.DB, exam *models.Exam, participant *models.ExamParticipant, fromStatus, toStatus int16) error {
	counts, err := exam.GetParticipantCounts()
	if err != nil {
		return status.Error(codes.Internal, "failed to get participant counts")
	}

	exam.ParticipantCounts, err = updateParticipantCounts(&counts, fromStatus, toStatus)
	if err != nil {
		return err
	}

	if err := tx.Save(exam).Error; err != nil {
		return status.Error(codes.Internal, "failed to update exam")
	}

	participant.Status = toStatus
	if err := tx.Save(participant).Error; err != nil {
		return status.Error(codes.Internal, "failed to update participant")
	}

	return nil
}

// examUpdateCtx holds the context for an exam update operation
type examUpdateCtx struct {
	exam      *models.Exam
	req       *proto.UpdateExamRequest
	now       time.Time
	isUpdated bool
}

// validateExamUpdate performs all validation checks for exam updates
func validateExamUpdate(ctx *examUpdateCtx) error {
	// After exam has ended, only title updates are allowed
	if ctx.now.After(ctx.exam.EndsAt) {
		if ctx.req.Title == nil || *ctx.req.Title == ctx.exam.Title {
			return status.Error(codes.FailedPrecondition, "cannot update exam after it has ended")
		}
		return nil
	}

	// After exam has started, certain fields cannot be updated
	if ctx.now.After(ctx.exam.StartsAt) {
		if ctx.req.StartsAt != nil {
			return status.Error(codes.FailedPrecondition, "cannot update start time after exam has started")
		}
		if ctx.req.Type != nil {
			return status.Error(codes.FailedPrecondition, "cannot update type after exam has started")
		}
		if ctx.req.DurationMinutes != nil {
			return status.Error(codes.FailedPrecondition, "cannot update duration after exam has started")
		}
	}

	return nil
}

// updateExamFields applies the update request to exam fields
func updateExamFields(ctx *examUpdateCtx) error {
	// After exam has ended, only update title
	if ctx.now.After(ctx.exam.EndsAt) {
		if ctx.req.Title != nil && *ctx.req.Title != ctx.exam.Title {
			ctx.exam.Title = *ctx.req.Title
			ctx.isUpdated = true
		}
		return nil
	}

	// Update basic fields
	if ctx.req.Title != nil && *ctx.req.Title != ctx.exam.Title {
		ctx.exam.Title = *ctx.req.Title
		ctx.isUpdated = true
	}

	// Update time fields
	if err := updateExamTimeFields(ctx); err != nil {
		return err
	}

	// Update exam settings
	if err := updateExamSettings(ctx); err != nil {
		return err
	}

	return nil
}

// updateExamTimeFields updates exam time-related fields
func updateExamTimeFields(ctx *examUpdateCtx) error {
	if ctx.req.StartsAt != nil {
		startsAt := ctx.req.StartsAt.AsTime()
		if err := validateExamStartTiming(startsAt); err != nil {
			return err
		}
		ctx.exam.StartsAt = startsAt
		ctx.isUpdated = true
	}

	if ctx.req.EndsAt != nil {
		endsAt := ctx.req.EndsAt.AsTime()
		if err := validateExamEndTiming(ctx.exam.StartsAt, endsAt); err != nil {
			return err
		}
		ctx.exam.EndsAt = endsAt
		ctx.isUpdated = true
	}

	return nil
}

// updateExamSettings updates exam configuration settings
func updateExamSettings(ctx *examUpdateCtx) error {
	if ctx.req.Type != nil {
		ctx.exam.Type = ctx.req.GetType()
		ctx.isUpdated = true
	}

	if ctx.req.DurationMinutes != nil {
		if err := validateExamDuration(ctx.req.GetDurationMinutes()); err != nil {
			return err
		}
		ctx.exam.DurationMinutes = int16(ctx.req.GetDurationMinutes())
		ctx.isUpdated = true
	}

	return nil
}
