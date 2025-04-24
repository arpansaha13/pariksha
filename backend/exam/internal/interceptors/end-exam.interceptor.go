package interceptors

import (
	"context"
	"strings"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"

	"pariksha/common/pkg/constants"
	"pariksha/common/pkg/models"
	"pariksha/exam/internal/config/db"
)

func endExamShouldIntercept(methodName string) bool {
	methodsToIntercept := []string{
		"/proto.ExamService/EndExam",
		"/proto.ExamService/UpsertAnswer",
	}

	for _, method := range methodsToIntercept {
		if strings.HasSuffix(methodName, method) {
			return true
		}
	}
	return false
}

// updateParticipantCounts updates the participant counts JSON in exam
func updateParticipantCounts(tx *gorm.DB, exam *models.Exam, oldStatus, newStatus int) error {
	counts, err := exam.GetParticipantCounts()
	if err != nil {
		return err
	}

	// Decrement old status count
	switch oldStatus {
	case constants.PARTICIPANT_STATUS_STARTED:
		counts.Started--
	}

	// Increment new status count
	switch newStatus {
	case constants.PARTICIPANT_STATUS_ENDED:
		counts.Ended++
	}

	// Update the exam's participant counts
	return tx.Model(exam).Update("participant_counts", counts).Error
}

// EndExamInterceptor is a fallback for the delayed-job for ending exam if the examWorker or examQueue fails.
// EndExamInterceptor should come after ExamAccessInterceptor. It uses participant data from context
func EndExamInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		methodName := info.FullMethod
		if !endExamShouldIntercept(methodName) {
			return handler(ctx, req)
		}

		participant, ok := GetParticipantFromContext(ctx)
		if !ok {
			return nil, status.Error(codes.Internal, "participant not found in context")
		}

		exam, ok := GetExamFromContext(ctx)
		if !ok {
			return nil, status.Error(codes.Internal, "exam not found in context")
		}

		shouldEndExam := participant.Status == constants.PARTICIPANT_STATUS_STARTED &&
			participant.ScheduledEndTime.Valid &&
			participant.ScheduledEndTime.Time.Before(time.Now())

		if shouldEndExam {
			// Start transaction
			tx := db.DB.Begin()
			if tx.Error != nil {
				return nil, status.Error(codes.Internal, "failed to start transaction")
			}

			oldStatus := participant.Status
			participant.Status = constants.PARTICIPANT_STATUS_ENDED
			participant.EndedAt.Time = time.Now()
			participant.EndedAt.Valid = true

			// Update participant status
			if err := tx.Save(participant).Error; err != nil {
				tx.Rollback()
				return nil, status.Error(codes.Internal, "failed to update participant status")
			}

			// Update exam participant counts
			if err := updateParticipantCounts(tx, exam, oldStatus, participant.Status); err != nil {
				tx.Rollback()
				return nil, status.Error(codes.Internal, "failed to update participant counts")
			}

			// Commit transaction
			if err := tx.Commit().Error; err != nil {
				tx.Rollback()
				return nil, status.Error(codes.Internal, "failed to commit transaction")
			}
		}

		return handler(ctx, req)
	}
}
