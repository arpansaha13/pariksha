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
func updateParticipantCounts(tx *gorm.DB, exam *models.Exam) error {
	counts, err := exam.GetParticipantCounts()
	if err != nil {
		return err
	}

	counts.Started--
	counts.Ended++

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

		now := time.Now()
		shouldEndExam := participant.Status == constants.PARTICIPANT_STATUS_STARTED &&
			participant.ScheduledEndTime.Valid &&
			participant.ScheduledEndTime.Time.Before(now)

		if shouldEndExam {
			err := db.DB.Transaction(func(tx *gorm.DB) error {
				participant.Status = constants.PARTICIPANT_STATUS_ENDED
				participant.EndedAt.Time = now
				participant.EndedAt.Valid = true

				// Update participant status
				if err := tx.Save(participant).Error; err != nil {
					return status.Error(codes.Internal, "failed to update participant status")
				}

				// Update exam participant counts
				if err := updateParticipantCounts(tx, exam); err != nil {
					return status.Error(codes.Internal, "failed to update participant counts")
				}

				return nil
			})

			if err != nil {
				return nil, err
			}
		}

		return handler(ctx, req)
	}
}
