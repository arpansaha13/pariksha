package handlers

import (
	"encoding/json"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

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

// updateParticipantCounts updates the exam's participant counts and returns marshaled counts
func updateParticipantCounts(counts *models.ParticipantCount, fromStatus int, toStatus int) (json.RawMessage, error) {
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

func createExamResponse(exam *models.Exam) (*proto.ExamResponse, error) {
	counts, err := exam.GetParticipantCounts()
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to parse participant counts")
	}

	return &proto.ExamResponse{
		Id:                 exam.ID,
		Title:              exam.Title,
		StartsAt:           timestamppb.New(exam.StartsAt),
		EndsAt:             timestamppb.New(exam.EndsAt),
		CreatedBy:          exam.CreatedBy,
		Type:               exam.Type,
		MaxCandidatesCount: exam.MaxCandidatesCount,
		DurationMinutes:    exam.DurationMinutes,
		PaperId:            exam.PaperID,
		ParticipantCounts: &proto.ParticipantCount{
			Unattended: int32(counts.Unattended),
			Invited:    int32(counts.Invited),
			Started:    int32(counts.Started),
			Ended:      int32(counts.Ended),
		},
	}, nil
}

// validateAnswerJSON validates the answer JSON based on question type
func validateAnswerJSON(answerJSON []byte, questionType string) error {
	switch questionType {
	case constants.QUESTION_TYPE_MCQ:
		var mcqAnswer models.MCQAnswer
		if err := json.Unmarshal(answerJSON, &mcqAnswer); err != nil {
			return status.Error(codes.InvalidArgument, "invalid MCQ answer format")
		}
		// Validate that optionIndex is non-negative
		if mcqAnswer.OptionIndex < 0 {
			return status.Error(codes.InvalidArgument, "option index cannot be negative")
		}
	case constants.QUESTION_TYPE_SHORT, constants.QUESTION_TYPE_LONG:
		var textAnswer models.GeneralAnswer
		if err := json.Unmarshal(answerJSON, &textAnswer); err != nil {
			return status.Error(codes.InvalidArgument, "invalid text answer format")
		}
		// Validate that text is not empty
		if textAnswer.Text == "" {
			return status.Error(codes.InvalidArgument, "answer text cannot be empty")
		}
	default:
		return status.Error(codes.InvalidArgument, "invalid question type")
	}
	return nil
}
