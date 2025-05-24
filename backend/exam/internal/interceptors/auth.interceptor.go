package interceptors

import (
	"context"
	"log"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"

	"pariksha/common/pkg/constants"
	"pariksha/common/pkg/models"
	"pariksha/common/pkg/proto"
	"pariksha/common/pkg/utils"
	"pariksha/exam/internal/config/db"
)

type contextKey string

const (
	examContextKey       contextKey = "exam"
	permissionContextKey contextKey = "permission"

	PERMISSION_DENIED_MESSAGE string = "no permission to perform this action"
	DATABASE_ERROR_MESSAGE    string = "database error"
)

var (
	requiresRead = map[string]bool{
		"/proto.Exam/GetExam":           true,
		"/proto.Exam/GetExamPermission": true,
	}

	// In case of LINK exam, allow access to these handlers even without a permission entry in db
	allowInLinkExam = map[string]bool{
		"/proto.Exam/GetExam":           true,
		"/proto.Exam/StartExam":         true,
		"/proto.Exam/GetExamPermission": true,
	}

	requiresWrite = map[string]bool{
		"/proto.Exam/UpdateExam":            true,
		"/proto.Exam/GetExamParticipants":   true,
		"/proto.Exam/AddExamParticipant":    true,
		"/proto.Exam/RemoveExamParticipant": true,
	}

	requiresParticipate = map[string]bool{
		"/proto.Exam/EndExam":            true,
		"/proto.Exam/UpsertAnswer":       true,
		"/proto.Exam/GetExamParticipant": true,
		"/proto.Exam/GetAnswerForExam":   true,
		"/proto.Exam/GetExamResults":     true,
	}

	requiresEvaluate = map[string]bool{
		"/proto.Exam/GetAnswerEvaluationData":    true,
		"/proto.Exam/UpdateAnswerForEvaluation":  true,
		"/proto.Exam/MarkParticipantAsEvaluated": true,
		"/proto.Exam/GetAnswerForEvaluation":     true,
		"/proto.Exam/GetParticipantById":         true,
	}

	handlerSpecificPermissionChecks = map[string]func(*models.ExamPermissions) bool{
		"/proto.Exam/GetExamQuestions": func(p *models.ExamPermissions) bool {
			return p.CanParticipate() || p.CanEvaluate()
		},
		"/proto.Exam/GetExamCategories": func(p *models.ExamPermissions) bool {
			return p.CanParticipate() || p.CanEvaluate()
		},
		"/proto.Exam/GetParticipantAnswers": func(p *models.ExamPermissions) bool {
			return p.CanParticipate() || p.CanEvaluate()
		},
	}
)

func shouldIntercept(methodName string) bool {
	return requiresRead[methodName] ||
		requiresWrite[methodName] ||
		requiresEvaluate[methodName] ||
		requiresParticipate[methodName] ||
		allowInLinkExam[methodName] ||
		handlerSpecificPermissionChecks[methodName] != nil
}

func checkPermissions(permission *models.ExamPermissions, methodName string) error {
	if requiresRead[methodName] && !permission.CanRead() {
		return status.Error(codes.PermissionDenied, PERMISSION_DENIED_MESSAGE)
	}
	if requiresWrite[methodName] && !permission.CanWrite() {
		return status.Error(codes.PermissionDenied, PERMISSION_DENIED_MESSAGE)
	}
	if requiresParticipate[methodName] && !permission.CanParticipate() {
		return status.Error(codes.PermissionDenied, PERMISSION_DENIED_MESSAGE)
	}
	if requiresEvaluate[methodName] && !permission.CanEvaluate() {
		return status.Error(codes.PermissionDenied, PERMISSION_DENIED_MESSAGE)
	}
	return nil
}

func GeneralExamAuthInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		methodName := info.FullMethod
		if !shouldIntercept(methodName) {
			return handler(ctx, req)
		}

		userID, err := utils.GetUserIDFromMetadata(ctx)
		if err != nil {
			return nil, err
		}

		examID, err := getExamIdFromRequest(req)
		if err != nil {
			return nil, err
		}

		exam, err := fetchExam(*examID)
		if err != nil {
			return nil, err
		}
		ctx = context.WithValue(ctx, examContextKey, exam)

		permission, err := fetchExamPermission(*examID, userID)
		if err == gorm.ErrRecordNotFound {
			if exam.Type == constants.EXAM_ACCESS_TYPE_LINK && allowInLinkExam[methodName] {
				return handler(ctx, req)
			}
			return nil, status.Error(codes.PermissionDenied, "No permission to access this exam")
		}
		if err != nil {
			return nil, status.Error(codes.Internal, "failed to fetch permissions")
		}
		ctx = context.WithValue(ctx, permissionContextKey, permission)

		// Check if there's a handler-specific permission check for this method
		if permissionCheckFn, exists := handlerSpecificPermissionChecks[methodName]; exists {
			if !permissionCheckFn(permission) {
				return nil, status.Error(codes.PermissionDenied, PERMISSION_DENIED_MESSAGE)
			}
			// Fall back to standard permission checking
		} else if err := checkPermissions(permission, methodName); err != nil {
			return nil, err
		}

		return handler(ctx, req)
	}
}

func fetchExam(examID int64) (*models.Exam, error) {
	var exam models.Exam
	if err := db.DB.Where("id IN (?)", []int64{examID}).Take(&exam).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, status.Error(codes.NotFound, "exam not found")
		}
		return nil, status.Error(codes.Internal, DATABASE_ERROR_MESSAGE)
	}
	return &exam, nil
}

func fetchExamPermission(examID int64, userID int64) (*models.ExamPermissions, error) {
	var permission models.ExamPermissions
	err := db.DB.Where("exam_id = ? AND user_id = ?", examID, userID).Take(&permission).Error
	if err != nil {
		return nil, err
	}
	return &permission, nil
}

func getExamIdFromRequest(req any) (*int64, error) {
	var examID int64

	switch r := req.(type) {
	// Group all cases that directly use ExamId field
	case *proto.ExamRequest,
		*proto.UpdateExamRequest,
		*proto.StartExamRequest,
		*proto.EndExamRequest,
		*proto.AddParticipantRequest,
		*proto.RemoveParticipantRequest,
		*proto.UpsertAnswersRequest,
		*proto.CheckParticipantRequest,
		*proto.GetExamParticipantRequest,
		*proto.GetAnswerRequest:
		examID = r.(interface{ GetExamId() int64 }).GetExamId()

	case *proto.UpdateAnswerRequest:
		// Find exam ID using joins, selecting only exam_id
		err := db.DB.Model(&models.ExamParticipant{}).
			Select("exam_participants.exam_id").
			Joins("INNER JOIN answers ON answers.exam_participant_id = exam_participants.id").
			Where("answers.id = ?", r.AnswerId).
			Take(&examID).Error
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				return nil, status.Error(codes.NotFound, "answer not found")
			}
			return nil, status.Error(codes.Internal, DATABASE_ERROR_MESSAGE)
		}

	case *proto.ParticipantQuestionRequest, *proto.ParticipantRequest:
		var participantID int64
		switch v := r.(type) {
		case *proto.ParticipantQuestionRequest:
			participantID = v.ParticipantId
		case *proto.ParticipantRequest:
			participantID = v.ParticipantId
		}

		if err := db.DB.Model(&models.ExamParticipant{}).
			Select("exam_id").
			Where("id = ?", participantID).
			Take(&examID).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				return nil, status.Error(codes.NotFound, "participant not found")
			}
			return nil, status.Error(codes.Internal, DATABASE_ERROR_MESSAGE)
		}

	default:
		log.Printf("Unhandled exam request type: %T", req)
		return nil, status.Error(codes.Internal, "unhandled request type in exam access interceptor")
	}

	return &examID, nil
}

// Getter function to safely access exam from context
func GetExamFromContext(ctx context.Context) (*models.Exam, bool) {
	exam, ok := ctx.Value(examContextKey).(*models.Exam)
	return exam, ok
}

// Getter function to safely access permission from context
func GetPermissionFromContext(ctx context.Context) (*models.ExamPermissions, bool) {
	permission, ok := ctx.Value(permissionContextKey).(*models.ExamPermissions)
	return permission, ok
}
