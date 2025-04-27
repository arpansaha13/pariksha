package interceptors

import (
	"context"
	"log"
	"strings"

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
	examContextKey            contextKey = "exam"
	participantContextKey     contextKey = "participant"
	PERMISSION_DENIED_MESSAGE string     = "no permission to perform this action"
	DATABASE_ERROR_MESSAGE    string     = "database error"
)

var (
	ownerMethods = []string{
		"/proto.ExamService/UpdateExam",
		"/proto.ExamService/GetExamParticipants",
		"/proto.ExamService/AddExamParticipant",
		"/proto.ExamService/RemoveExamParticipant",
		// "/proto.ExamService/GetParticipantAnswers",
		// "/proto.ExamService/UpdateAnswerForEvaluation",
		// "/proto.ExamService/MarkAsEvaluated",
	}

	notOwnerMethods = []string{
		"/proto.ExamService/StartExam",
		"/proto.ExamService/EndExam",
		"/proto.ExamService/UpsertAnswer",
		"/proto.ExamService/CheckExamParticipant",
		"/proto.ExamService/GetExamQuestions",
		"/proto.ExamService/GetExamCategories",
		"/proto.ExamService/UpsertAnswer",
	}

	participantMethods = []string{
		"/proto.ExamService/EndExam",
		"/proto.ExamService/UpsertAnswer",
		"/proto.ExamService/CheckExamParticipant",
		"/proto.ExamService/GetExamQuestions",
		"/proto.ExamService/GetExamCategories",
		"/proto.ExamService/GetExamParticipant",
		"/proto.ExamService/GetAnswer",
		"/proto.ExamService/UpsertAnswer",
	}

	defaultMethods = []string{
		"/proto.ExamService/GetExam",
		"/proto.ExamService/CheckExamAccess",
		"/proto.ExamService/GetAnswerById",
	}
)

func shouldIntercept(methodName string) bool {
	// Merge all method arrays once
	allMethods := make([]string, 0, len(ownerMethods)+len(notOwnerMethods)+len(participantMethods)+len(defaultMethods))
	allMethods = append(allMethods, ownerMethods...)
	allMethods = append(allMethods, notOwnerMethods...)
	allMethods = append(allMethods, participantMethods...)
	allMethods = append(allMethods, defaultMethods...)

	for _, method := range allMethods {
		if strings.HasSuffix(methodName, method) {
			return true
		}
	}
	return false
}

type AuthorizationRule interface {
	Authorize(ctx context.Context, methodName string, exam *models.Exam, userID int64) (context.Context, error)
	ShouldApply(methodName string) bool
	ShouldStopOnError() bool
	ShouldStopOnSuccess() bool
}

type OwnerRule struct{}

func (r *OwnerRule) Authorize(ctx context.Context, methodName string, exam *models.Exam, userID int64) (context.Context, error) {
	if exam.CreatedBy != userID {
		return ctx, status.Error(codes.PermissionDenied, PERMISSION_DENIED_MESSAGE)
	}
	return ctx, nil
}

func (r *OwnerRule) ShouldApply(methodName string) bool {
	for _, method := range ownerMethods {
		if strings.HasSuffix(methodName, method) {
			return true
		}
	}
	return false
}

func (r *OwnerRule) ShouldStopOnError() bool {
	return true
}

func (r *OwnerRule) ShouldStopOnSuccess() bool {
	return true
}

type NotOwnerRule struct{}

func (r *NotOwnerRule) Authorize(ctx context.Context, methodName string, exam *models.Exam, userID int64) (context.Context, error) {
	if exam.CreatedBy == userID {
		return ctx, status.Error(codes.PermissionDenied, PERMISSION_DENIED_MESSAGE)
	}
	return ctx, nil
}

func (r *NotOwnerRule) ShouldApply(methodName string) bool {
	for _, method := range notOwnerMethods {
		if strings.HasSuffix(methodName, method) {
			return true
		}
	}
	return false
}

func (r *NotOwnerRule) ShouldStopOnError() bool {
	return true
}

func (r *NotOwnerRule) ShouldStopOnSuccess() bool {
	return false
}

type ParticipantRule struct{}

func (r *ParticipantRule) Authorize(ctx context.Context, methodName string, exam *models.Exam, userID int64) (context.Context, error) {
	participant, err := fetchParticipant(exam.ID, userID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return ctx, status.Error(codes.PermissionDenied, PERMISSION_DENIED_MESSAGE)
		}
		return ctx, status.Error(codes.Internal, DATABASE_ERROR_MESSAGE)
	}
	return context.WithValue(ctx, participantContextKey, participant), nil
}

func (r *ParticipantRule) ShouldApply(methodName string) bool {
	for _, method := range participantMethods {
		if strings.HasSuffix(methodName, method) {
			return true
		}
	}
	return false
}

func (r *ParticipantRule) ShouldStopOnError() bool {
	return true
}

func (r *ParticipantRule) ShouldStopOnSuccess() bool {
	return true
}

type DefaultRule struct{}

func (r *DefaultRule) Authorize(ctx context.Context, methodName string, exam *models.Exam, userID int64) (context.Context, error) {
	// Allow owner
	if exam.CreatedBy == userID {
		return ctx, nil
	}

	participant, err := fetchParticipant(exam.ID, userID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			if exam.Type != constants.EXAM_ACCESS_TYPE_LINK {
				return ctx, status.Error(codes.PermissionDenied, PERMISSION_DENIED_MESSAGE)
			}
			return ctx, nil
		}
		return ctx, status.Error(codes.Internal, DATABASE_ERROR_MESSAGE)
	}
	return context.WithValue(ctx, participantContextKey, participant), nil
}

func (r *DefaultRule) ShouldApply(methodName string) bool {
	return true
}

func (r *DefaultRule) ShouldStopOnError() bool {
	return true
}

func (r *DefaultRule) ShouldStopOnSuccess() bool {
	return true
}

type AuthorizationChain struct {
	rules []AuthorizationRule
}

func NewAuthorizationChain() *AuthorizationChain {
	return &AuthorizationChain{
		rules: []AuthorizationRule{
			&OwnerRule{},
			&NotOwnerRule{},
			&ParticipantRule{},
			&DefaultRule{},
		},
	}
}

func (c *AuthorizationChain) Authorize(ctx context.Context, methodName string, exam *models.Exam, userID int64) (context.Context, error) {
	for _, rule := range c.rules {
		if rule.ShouldApply(methodName) {
			newCtx, err := rule.Authorize(ctx, methodName, exam, userID)
			if err != nil {
				if rule.ShouldStopOnError() {
					return ctx, err
				}
				continue
			}
			ctx = newCtx
			if rule.ShouldStopOnSuccess() {
				return ctx, nil
			}
		}
	}
	return ctx, nil
}

func ExamAccessInterceptor() grpc.UnaryServerInterceptor {
	authChain := NewAuthorizationChain()

	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		methodName := info.FullMethod
		if !shouldIntercept(methodName) {
			return handler(ctx, req)
		}

		userID, err := utils.GetUserIDFromMetadata(ctx)
		if err != nil {
			return nil, err
		}

		var examID int64
		switch r := req.(type) {
		case *proto.ExamRequest:
			examID = r.ExamId
		case *proto.UpdateExamRequest:
			examID = r.ExamId
		case *proto.StartExamRequest:
			examID = r.ExamId
		case *proto.EndExamRequest:
			examID = r.ExamId
		case *proto.AddParticipantRequest:
			examID = r.ExamId
		case *proto.RemoveParticipantRequest:
			examID = r.ExamId
		case *proto.UpsertAnswersRequest:
			examID = r.ExamId
		case *proto.CheckParticipantRequest:
			examID = r.ExamId
		case *proto.GetExamParticipantRequest:
			examID = r.ExamId
		case *proto.GetAnswerRequest:
			examID = r.ExamId
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
		default:
			log.Printf("Unhandled exam request type: %T", req)
			return nil, status.Error(codes.Internal, "unhandled request type in exam access interceptor")
		}

		exam, err := fetchExam(examID)
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				return nil, status.Error(codes.NotFound, "exam not found")
			}
			return nil, status.Error(codes.Internal, DATABASE_ERROR_MESSAGE)
		}

		ctx = context.WithValue(ctx, examContextKey, exam)
		ctx, err = authChain.Authorize(ctx, methodName, exam, userID)
		if err != nil {
			return nil, err
		}

		return handler(ctx, req)
	}
}

func fetchExam(examID int64) (*models.Exam, error) {
	var exam models.Exam
	err := db.DB.Take(&exam, examID).Error
	if err != nil {
		return nil, err
	}
	return &exam, nil
}

func fetchParticipant(examID int64, userID int64) (*models.ExamParticipant, error) {
	var participant models.ExamParticipant
	err := db.DB.Where("exam_id = ? AND user_id = ?", examID, userID).Take(&participant).Error
	if err != nil {
		return nil, err
	}
	return &participant, nil
}

// Getter function to safely access exam from context
func GetExamFromContext(ctx context.Context) (*models.Exam, bool) {
	exam, ok := ctx.Value(examContextKey).(*models.Exam)
	return exam, ok
}

// Getter function to safely access exam from context
func GetParticipantFromContext(ctx context.Context) (*models.ExamParticipant, bool) {
	participant, ok := ctx.Value(participantContextKey).(*models.ExamParticipant)
	return participant, ok
}
