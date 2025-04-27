package handlers

import (
	"context"
	"database/sql"
	"strconv"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"

	"pariksha/common/pkg/constants"
	"pariksha/common/pkg/models"
	"pariksha/common/pkg/proto"
	"pariksha/exam/internal/config/db"
	"pariksha/exam/internal/interceptors"
)

func (s *ExamServer) GetParticipantAnswers(ctx context.Context, req *proto.ParticipantRequest) (*proto.AnswerList, error) {
	var answers []models.Answer
	if err := db.DB.Where("exam_participant_id = ?", req.ParticipantId).Find(&answers).Error; err != nil {
		return nil, status.Error(codes.NotFound, "answers not found")
	}

	if len(answers) == 0 {
		return nil, status.Error(codes.NotFound, "answers not found")
	}

	response := &proto.AnswerList{
		Answers: make([]*proto.AnswerResponse, len(answers)),
	}

	for i, answer := range answers {
		response.Answers[i] = &proto.AnswerResponse{
			Id:                answer.ID,
			ExamParticipantId: answer.ExamParticipantID,
			QuestionId:        answer.QuestionID,
			Answer:            answer.Answer,
			Comments:          answer.Comments.String,
			ScoreAwarded:      int32(answer.ScoreAwarded),
		}
	}

	return response, nil
}

// GetAnswer finds an answer using participant ID and question ID and returns minimal info
func (s *ExamServer) GetAnswer(ctx context.Context, req *proto.GetAnswerRequest) (*proto.GetAnswerResponse, error) {
	var answer models.Answer
	if err := db.DB.Where("exam_participant_id = ? AND question_id = ?", req.ParticipantId, req.QuestionId).Take(&answer).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, status.Error(codes.NotFound, "answer not found")
		}
		return nil, status.Error(codes.Internal, "database error")
	}

	return &proto.GetAnswerResponse{
		Id:     answer.ID,
		Answer: answer.Answer,
	}, nil
}

// GetAnswerById finds an answer using its ID
func (s *ExamServer) GetAnswerById(ctx context.Context, req *proto.GetAnswerByIdRequest) (*proto.AnswerResponse, error) {
	var answer models.Answer
	if err := db.DB.Take(&answer, req.AnswerId).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, status.Error(codes.NotFound, "answer not found")
		}
		return nil, status.Error(codes.Internal, "database error")
	}

	return &proto.AnswerResponse{
		Id:                answer.ID,
		ExamParticipantId: answer.ExamParticipantID,
		QuestionId:        answer.QuestionID,
		Answer:            answer.Answer,
		Comments:          answer.Comments.String,
		ScoreAwarded:      int32(answer.ScoreAwarded),
	}, nil
}

func (s *ExamServer) UpsertAnswer(ctx context.Context, req *proto.UpsertAnswersRequest) (*proto.UpsertAnswersResponse, error) {
	participant, ok := interceptors.GetParticipantFromContext(ctx)
	if !ok {
		return nil, status.Error(codes.Internal, "participant not found in context")
	}

	if participant.Status != constants.PARTICIPANT_STATUS_STARTED {
		return nil, status.Error(codes.FailedPrecondition, "participant has not started the exam")
	}

	if participant.Status == constants.PARTICIPANT_STATUS_ENDED {
		return nil, status.Error(codes.FailedPrecondition, "participant has ended the exam")
	}

	// Get question type from metadata
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Error(codes.Internal, "missing metadata")
	}

	questionTypes := md.Get("question_type")
	if len(questionTypes) == 0 {
		return nil, status.Error(codes.Internal, "missing question type in metadata")
	}
	questionType := questionTypes[0]

	// Validate answer JSON based on question type
	if err := validateAnswerJSON(req.Answer.Answer, questionType); err != nil {
		return nil, err
	}

	var answer models.Answer
	if err := db.DB.Where("exam_participant_id = ? AND question_id = ?", participant.ID, req.Answer.QuestionId).Take(&answer).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			// Create new answer
			answer = models.Answer{
				ExamParticipantID: participant.ID,
				QuestionID:        req.Answer.QuestionId,
				Answer:            req.Answer.Answer,
			}
			if err := db.DB.Create(&answer).Error; err != nil {
				return nil, status.Error(codes.Internal, "failed to create answer")
			}
		} else {
			return nil, status.Error(codes.Internal, "database error")
		}
	} else {
		// Update existing answer
		answer.Answer = req.Answer.Answer
		if err := db.DB.Save(&answer).Error; err != nil {
			return nil, status.Error(codes.Internal, "failed to update answer")
		}
	}

	return &proto.UpsertAnswersResponse{
		AnswerId: answer.ID,
	}, nil
}

func (s *ExamServer) UpdateAnswerForEvaluation(ctx context.Context, req *proto.UpdateAnswerRequest) (*proto.Empty, error) {
	// Get max score from metadata
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Error(codes.Internal, "missing metadata")
	}

	scores := md.Get("question_score")
	if len(scores) == 0 {
		return nil, status.Error(codes.Internal, "missing question score")
	}

	maxScore, err := strconv.Atoi(scores[0])
	if err != nil {
		return nil, status.Error(codes.Internal, "invalid question score")
	}

	if req.NewScore != nil && req.GetNewScore() > int32(maxScore) {
		return nil, status.Error(codes.InvalidArgument, "new score exceeds max score for the question")
	}

	err = db.DB.Transaction(func(tx *gorm.DB) error {
		var answer models.Answer
		if err := tx.Take(&answer, req.AnswerId).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				return status.Error(codes.NotFound, "answer not found")
			}
			return status.Error(codes.Internal, "database error")
		}

		isUpdated := false

		if req.NewScore != nil {
			var examParticipant models.ExamParticipant
			if err := tx.Take(&examParticipant, answer.ExamParticipantID).Error; err != nil {
				return status.Error(codes.NotFound, "exam participant not found")
			}

			examParticipant.ScoreAwarded = examParticipant.ScoreAwarded - answer.ScoreAwarded + int(*req.NewScore)
			answer.ScoreAwarded = int(*req.NewScore)

			if err := tx.Save(&examParticipant).Error; err != nil {
				return status.Error(codes.Internal, "failed to update exam participant")
			}

			isUpdated = true
		}

		if req.Evaluated != nil {
			answer.Evaluated = *req.Evaluated
			isUpdated = true
		}

		if req.Comments != nil {
			answer.Comments = sql.NullString{String: *req.Comments, Valid: true}
			isUpdated = true
		}

		if isUpdated {
			if err := tx.Save(&answer).Error; err != nil {
				return status.Error(codes.Internal, "failed to update answer")
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return &proto.Empty{}, nil
}

func (s *ExamServer) MarkAsEvaluated(ctx context.Context, req *proto.ParticipantRequest) (*proto.EvaluationStatusResponse, error) {
	err := db.DB.Transaction(func(tx *gorm.DB) error {
		var examParticipant models.ExamParticipant
		if err := tx.Take(&examParticipant, req.ParticipantId).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				return status.Error(codes.NotFound, "exam participant not found")
			}
			return status.Error(codes.Internal, "database error")
		}

		if examParticipant.Status != constants.PARTICIPANT_STATUS_ENDED {
			return status.Error(codes.FailedPrecondition, "evaluation can only start if the exam has ended")
		}

		var unevaluatedCount int64
		if err := tx.Model(&models.Answer{}).Where("exam_participant_id = ? AND evaluated = ?", req.ParticipantId, false).Count(&unevaluatedCount).Error; err != nil {
			return status.Error(codes.Internal, "failed to count unevaluated answers")
		}

		if unevaluatedCount == 0 {
			examParticipant.Status = constants.PARTICIPANT_STATUS_EVALUATED
			if err := tx.Save(&examParticipant).Error; err != nil {
				return status.Error(codes.Internal, "failed to update exam participant status")
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	// Get final unevaluated count
	var unevaluatedCount int64
	if err := db.DB.Model(&models.Answer{}).Where("exam_participant_id = ? AND evaluated = ?", req.ParticipantId, false).Count(&unevaluatedCount).Error; err != nil {
		return nil, status.Error(codes.Internal, "failed to count unevaluated answers")
	}

	return &proto.EvaluationStatusResponse{
		UnevaluatedCount: int32(unevaluatedCount),
	}, nil
}
