package services

import (
	"context"
	"slices"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"

	"pariksha/common/pkg/constants"
	"pariksha/common/pkg/models"
	"pariksha/common/pkg/proto"
	"pariksha/common/pkg/types"
	"pariksha/common/pkg/utils"
	"pariksha/exam/internal/interceptors"
	"pariksha/exam/internal/interservice"
	"pariksha/exam/internal/repositories"
)

type Evaluation struct {
	answerRepo      *repositories.Answer
	participantRepo *repositories.Participant
	questionIntSvc  *interservice.Question
}

func NewEvaluation(
	answerRepo *repositories.Answer,
	participantRepo *repositories.Participant,
	questionIntSvc *interservice.Question,
) *Evaluation {
	return &Evaluation{
		answerRepo:      answerRepo,
		participantRepo: participantRepo,
		questionIntSvc:  questionIntSvc,
	}
}

// GetAnswerForEvaluation retrieves an answer for evaluation purposes.
func (s *Evaluation) GetAnswerForEvaluation(ctx context.Context, req *proto.ParticipantQuestionRequest) (*proto.AnswerMinimalResponse, error) {
	questionID, ok := interceptors.GetQuestionIDFromContext(ctx)
	if !ok {
		return nil, status.Error(codes.Internal, "question ID not found in context")
	}

	answer, err := s.answerRepo.GetAnswerForEvaluation(nil, types.ParticipantID(req.ParticipantId), types.QuestionID(questionID))
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return &proto.AnswerMinimalResponse{
				QuestionHash: req.QuestionHash,
			}, nil
		}
		return nil, status.Error(codes.Internal, "failed to fetch answer")
	}

	var answerBytes []byte
	if answer.Answer != nil {
		answerBytes = *answer.Answer
	}

	return &proto.AnswerMinimalResponse{
		AnswerId:     int64(answer.ID),
		Answer:       answerBytes,
		QuestionHash: req.QuestionHash,
	}, nil
}

// GetAnswerEvaluationData retrieves answer evaluation data.
func (s *Evaluation) GetAnswerEvaluationData(ctx context.Context, req *proto.ParticipantQuestionRequest) (*proto.GetAnswerEvaluationDataResponse, error) {
	questionID, ok := interceptors.GetQuestionIDFromContext(ctx)
	if !ok {
		return nil, status.Error(codes.Internal, "question ID not found in context")
	}

	participant, err := s.participantRepo.GetByID(nil, types.ParticipantID(req.ParticipantId))
	if err != nil {
		return nil, utils.HandleDBError(err, "participant not found")
	}

	if !slices.Contains([]int16{
		constants.PARTICIPANT_STATUS_ENDED,
		constants.PARTICIPANT_STATUS_EVALUATED,
	}, participant.Status) {
		return nil, status.Error(codes.FailedPrecondition, "cannot evaluate answers before exam completion")
	}

	answer, err := s.answerRepo.GetAnswerEvaluationData(nil, types.ParticipantID(req.ParticipantId), types.QuestionID(questionID))
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return &proto.GetAnswerEvaluationDataResponse{
				QuestionHash: req.QuestionHash,
			}, nil
		}
		return nil, utils.HandleDBError(err, "answer not found")
	}

	return &proto.GetAnswerEvaluationDataResponse{
		AnswerId:     int64(answer.ID),
		QuestionHash: req.QuestionHash,
		ScoreAwarded: int32(answer.ScoreAwarded),
	}, nil
}

// UpdateAnswerForEvaluation updates answer fields for evaluation.
func (s *Evaluation) UpdateAnswerForEvaluation(ctx context.Context, req *proto.UpdateAnswerRequest) (*proto.GetAnswerEvaluationDataResponse, error) {
	result, err := s.answerRepo.GetAnswerMaxScoreAndParticipantStatus(nil, types.AnswerID(req.AnswerId))
	if err != nil {
		return nil, utils.HandleDBError(err, "answer not found")
	}

	if result.ParticipantStatus != constants.PARTICIPANT_STATUS_ENDED {
		return nil, status.Error(codes.FailedPrecondition, "participant cannot be evaluated")
	}

	if req.NewScore != nil && req.GetNewScore() > int32(result.MaxScore) {
		return nil, status.Error(codes.InvalidArgument, "new score exceeds max score for the question")
	}

	var answer *models.Answer
	err = s.answerRepo.Transaction(func(tx *gorm.DB) error {
		var txAnswer *models.Answer
		txAnswer, err = utils.FindRecord[models.Answer](tx, req.AnswerId, "answer not found")
		if err != nil {
			return err
		}

		if req.NewScore != nil {
			if err := s.answerRepo.UpdateScoreAwarded(tx, txAnswer.ExamParticipantID, int32(txAnswer.ScoreAwarded), req.GetNewScore()); err != nil {
				return status.Error(codes.Internal, "failed to update exam participant")
			}
			txAnswer.ScoreAwarded = int16(req.GetNewScore())
		}

		if req.Evaluated != nil {
			txAnswer.Evaluated = *req.Evaluated
		}

		if err := s.answerRepo.UpdateAnswerForEvaluation(tx, txAnswer); err != nil {
			return err
		}
		answer = txAnswer
		return nil
	})
	if err != nil {
		return nil, err
	}

	questionHashes, err := s.questionIntSvc.GetQuestionHashesByIds(ctx, []types.QuestionID{answer.QuestionID})
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to fetch question hash")
	}

	return &proto.GetAnswerEvaluationDataResponse{
		AnswerId:     int64(answer.ID),
		QuestionHash: questionHashes[0],
		ScoreAwarded: int32(answer.ScoreAwarded),
	}, nil
}

// MarkParticipantAsEvaluated marks a participant as evaluated if all answers are evaluated.
func (s *Evaluation) MarkParticipantAsEvaluated(ctx context.Context, req *proto.ParticipantRequest) (*proto.EvaluationStatusResponse, error) {
	var unevaluatedCount int64

	err := s.answerRepo.Transaction(func(tx *gorm.DB) error {
		participant, err := s.participantRepo.GetByID(tx, types.ParticipantID(req.ParticipantId))
		if err != nil {
			return err
		}

		if participant.Status != constants.PARTICIPANT_STATUS_ENDED {
			return status.Error(codes.FailedPrecondition, "evaluation can only start if the exam has ended")
		}

		unevaluatedCount, err = s.answerRepo.CountUnevaluatedAnswers(tx, types.ParticipantID(req.ParticipantId))
		if err != nil {
			return status.Error(codes.Internal, "failed to count unevaluated answers")
		}

		if unevaluatedCount == 0 {
			if err := s.participantRepo.UpdateStatus(tx, participant, constants.PARTICIPANT_STATUS_EVALUATED); err != nil {
				return status.Error(codes.Internal, "failed to update exam participant status")
			}
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return &proto.EvaluationStatusResponse{
		UnevaluatedCount: int32(unevaluatedCount),
	}, nil
}
