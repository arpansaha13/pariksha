package interceptors

import (
	"context"
	"log"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"

	"pariksha/common/pkg/models"
	"pariksha/common/pkg/proto"
	"pariksha/common/pkg/utils"
	"pariksha/exam/internal/repositories"
)

type contextKey string

const (
	examContextKey       contextKey = "exam"
	permissionContextKey contextKey = "permission"

	PERMISSION_DENIED_MESSAGE string = "no permission to perform this action"
	DATABASE_ERROR_MESSAGE    string = "database error"
)

var (
	requiresRead = map[string]struct{}{
		"/proto.Exam/GetExam":           {},
		"/proto.Exam/GetExamPermission": {},
	}

	// In case of LINK exam, allow access to these handlers even without a permission entry in db
	allowInLinkExam = map[string]struct{}{
		"/proto.Exam/GetExam":           {},
		"/proto.Exam/StartExam":         {},
		"/proto.Exam/GetExamPermission": {},
	}

	requiresWrite = map[string]struct{}{
		"/proto.Exam/UpdateExam":            {},
		"/proto.Exam/GetExamParticipants":   {},
		"/proto.Exam/AddExamParticipant":    {},
		"/proto.Exam/RemoveExamParticipant": {},
	}

	requiresParticipate = map[string]struct{}{
		"/proto.Exam/EndExam":            {},
		"/proto.Exam/UpsertAnswer":       {},
		"/proto.Exam/GetExamParticipant": {},
		"/proto.Exam/GetAnswerForExam":   {},
		"/proto.Exam/GetExamResults":     {},
	}

	requiresEvaluate = map[string]struct{}{
		"/proto.Exam/GetAnswerEvaluationData":    {},
		"/proto.Exam/UpdateAnswerForEvaluation":  {},
		"/proto.Exam/MarkParticipantAsEvaluated": {},
		"/proto.Exam/GetAnswerForEvaluation":     {},
		"/proto.Exam/GetParticipantById":         {},
	}

	handlerSpecificPermissionChecks = map[string]func(*models.ExamPermission) bool{
		"/proto.Exam/GetExamQuestions": func(p *models.ExamPermission) bool {
			return p.CanParticipate() || p.CanEvaluate()
		},
		"/proto.Exam/GetExamCategories": func(p *models.ExamPermission) bool {
			return p.CanParticipate() || p.CanEvaluate()
		},
		"/proto.Exam/GetParticipantAnswers": func(p *models.ExamPermission) bool {
			return p.CanParticipate() || p.CanEvaluate()
		},
	}
)

func shouldIntercept(methodName string) bool {
	_, hasRead := requiresRead[methodName]
	_, hasWrite := requiresWrite[methodName]
	_, hasEvaluate := requiresEvaluate[methodName]
	_, hasParticipate := requiresParticipate[methodName]
	_, allowsLink := allowInLinkExam[methodName]
	_, hasSpecificCheck := handlerSpecificPermissionChecks[methodName]

	return hasRead || hasWrite || hasEvaluate || hasParticipate || allowsLink || hasSpecificCheck
}

func checkPermissions(permission *models.ExamPermission, methodName string) error {
	if _, ok := requiresRead[methodName]; ok && !permission.CanRead() {
		return status.Error(codes.PermissionDenied, PERMISSION_DENIED_MESSAGE)
	}
	if _, ok := requiresWrite[methodName]; ok && !permission.CanWrite() {
		return status.Error(codes.PermissionDenied, PERMISSION_DENIED_MESSAGE)
	}
	if _, ok := requiresParticipate[methodName]; ok && !permission.CanParticipate() {
		return status.Error(codes.PermissionDenied, PERMISSION_DENIED_MESSAGE)
	}
	if _, ok := requiresEvaluate[methodName]; ok && !permission.CanEvaluate() {
		return status.Error(codes.PermissionDenied, PERMISSION_DENIED_MESSAGE)
	}
	return nil
}

func GeneralExamAuthInterceptor(
	examRepo *repositories.Exam,
	permissionRepo *repositories.Permission,
	dbInst *gorm.DB, // TODO: Temporary fix
) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		methodName := info.FullMethod
		if !shouldIntercept(methodName) {
			return handler(ctx, req)
		}

		userID, err := utils.GetUserIDFromMetadata(ctx)
		if err != nil {
			return nil, err
		}

		examHash, err := getExamHashFromRequest(ctx, req, dbInst)
		if err != nil {
			return nil, err
		}

		exam, err := examRepo.GetByHash(nil, *examHash)
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				return nil, status.Error(codes.NotFound, "exam not found")
			}
			return nil, status.Error(codes.Internal, DATABASE_ERROR_MESSAGE)
		}

		ctx = context.WithValue(ctx, examContextKey, exam)

		permission, err := permissionRepo.GetByExamHashAndUserId(nil, *examHash, userID)
		if err == gorm.ErrRecordNotFound {
			if _, ok := allowInLinkExam[methodName]; ok && exam.Type == proto.ExamType_LINK {
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

func getExamHashFromRequest(ctx context.Context, req any, dbInst *gorm.DB) (*string, error) {
	var examHash string

	switch r := req.(type) {
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
		examHash = r.(interface{ GetExamHash() string }).GetExamHash()

	case *proto.UpdateAnswerRequest:
		// Find exam ID using joins, selecting only exam_id
		err := dbInst.Model(&models.ExamParticipant{}).
			Select("exams.hash").
			Joins("INNER JOIN exams ON exams.id = exam_participants.exam_id").
			Joins("INNER JOIN answers ON answers.exam_participant_id = exam_participants.id").
			Where("answers.id = ?", r.AnswerId).
			Take(&examHash).Error
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

		if err := dbInst.Model(&models.ExamParticipant{}).
			Select("exams.hash").
			Joins("INNER JOIN exams ON exams.id = exam_participants.exam_id").
			Where("exam_participants.id = ?", participantID).
			Take(&examHash).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				return nil, status.Error(codes.NotFound, "participant not found")
			}
			return nil, status.Error(codes.Internal, DATABASE_ERROR_MESSAGE)
		}

	default:
		log.Printf("Unhandled exam request type: %T", req)
		return nil, status.Error(codes.Internal, "unhandled request type in exam access interceptor")
	}

	return &examHash, nil
}

// Getter function to safely access exam from context
func GetExamFromContext(ctx context.Context) (*models.Exam, bool) {
	exam, ok := ctx.Value(examContextKey).(*models.Exam)
	return exam, ok
}

// Getter function to safely access permission from context
func GetPermissionFromContext(ctx context.Context) (*models.ExamPermission, bool) {
	permission, ok := ctx.Value(permissionContextKey).(*models.ExamPermission)
	return permission, ok
}
