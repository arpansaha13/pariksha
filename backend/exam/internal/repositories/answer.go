package repositories

import (
	"database/sql"
	"encoding/json"

	"gorm.io/gorm"

	"pariksha/common/pkg/models"
	"pariksha/common/pkg/types"
	"pariksha/common/pkg/utils"
)

type Answer struct {
	db *gorm.DB
}

func NewAnswer(db *gorm.DB) *Answer {
	return &Answer{db: db}
}

// getTx returns tx if not nil, otherwise returns r.db
func (r *Answer) getTx(tx *gorm.DB) *gorm.DB {
	if tx == nil {
		return r.db
	}
	return tx
}

func (r *Answer) Transaction(fc func(tx *gorm.DB) error, opts ...*sql.TxOptions) error {
	return utils.TransactionHandler(r.db, fc, opts...)
}

// QueryResult represents a joined answer and question row for participant answers.
type QueryResult struct {
	ID                types.AnswerID      `gorm:"primaryKey;type:bigint"`
	ExamParticipantID types.ParticipantID `gorm:"type:bigint"`
	Answer            *json.RawMessage    `gorm:"type:json"`

	QuestionID types.QuestionID `gorm:"type:bigint"`
	Order      int16            `gorm:"column:order"`
	CategoryID types.CategoryID `gorm:"column:category_id"`
	MaxScore   int16            `gorm:"column:max_score"`
}

// GetParticipantAnswers fetches answers and related question info for a participant and exam.
func (r *Answer) GetParticipantAnswers(tx *gorm.DB, examID types.ExamID, participantID types.ParticipantID) ([]QueryResult, error) {
	tx = r.getTx(tx)
	var results []QueryResult
	err := tx.Table("exam_questions").
		Select("exam_questions.question_id, exam_questions.order, exam_questions.category_id, exam_questions.max_score", "answers.id, answers.answer, answers.exam_participant_id").
		Joins("LEFT JOIN answers ON exam_questions.question_id = answers.question_id AND answers.exam_participant_id = ?", participantID).
		Where("exam_questions.exam_id = ?", examID).
		Find(&results).Error
	return results, err
}

// GetAnswerByParticipantAndQuestion fetches an answer by participant ID and question ID.
func (r *Answer) GetAnswerByParticipantAndQuestion(tx *gorm.DB, participantID types.ParticipantID, questionID types.QuestionID) (*models.Answer, error) {
	tx = r.getTx(tx)
	var answer models.Answer
	err := tx.Where("exam_participant_id = ? AND question_id = ? AND answer IS NOT NULL", participantID, questionID).Take(&answer).Error
	if err != nil {
		return nil, err
	}
	return &answer, nil
}

// UpsertAnswer creates or updates an answer for a participant and question.
func (r *Answer) UpsertAnswer(tx *gorm.DB, participantID types.ParticipantID, questionID types.QuestionID, answerContent *json.RawMessage) (*models.Answer, error) {
	tx = r.getTx(tx)
	var answer models.Answer
	err := tx.Where("exam_participant_id = ? AND question_id = ?", participantID, questionID).Take(&answer).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			answer = models.Answer{
				ExamParticipantID: participantID,
				QuestionID:        questionID,
				Answer:            answerContent,
			}
			if err := tx.Create(&answer).Error; err != nil {
				return nil, err
			}
			return &answer, nil
		}
		return nil, err
	}
	answer.Answer = answerContent
	if err := tx.Save(&answer).Error; err != nil {
		return nil, err
	}
	return &answer, nil
}

// GetAnswerForEvaluation fetches answer for evaluation by participant and question ID.
func (r *Answer) GetAnswerForEvaluation(tx *gorm.DB, participantID types.ParticipantID, questionID types.QuestionID) (*models.Answer, error) {
	tx = r.getTx(tx)
	var answer models.Answer
	err := tx.Model(&models.Answer{}).
		Select("id", "question_id", "answer").
		Where("exam_participant_id = ? AND question_id = ?", participantID, questionID).
		Take(&answer).Error
	if err != nil {
		return nil, err
	}
	return &answer, nil
}

// GetAnswerEvaluationData fetches answer evaluation data by participant and question ID.
func (r *Answer) GetAnswerEvaluationData(tx *gorm.DB, participantID types.ParticipantID, questionID types.QuestionID) (*models.Answer, error) {
	tx = r.getTx(tx)
	var answer models.Answer
	err := tx.Model(&models.Answer{}).
		Select("id", "question_id", "score_awarded").
		Where("exam_participant_id = ? AND question_id = ? AND answer IS NOT NULL", participantID, questionID).
		Take(&answer).Error
	if err != nil {
		return nil, err
	}
	return &answer, nil
}

// GetAnswerMaxScoreAndParticipantStatus fetches max score and participant status for an answer.
type AnswerMaxScoreStatus struct {
	MaxScore          int16
	ParticipantStatus int16 `gorm:"column:status"`
}

func (r *Answer) GetAnswerMaxScoreAndParticipantStatus(tx *gorm.DB, answerID types.AnswerID) (*AnswerMaxScoreStatus, error) {
	tx = r.getTx(tx)
	var result AnswerMaxScoreStatus
	err := tx.Table("answers").
		Joins("INNER JOIN exam_participants ON exam_participants.id = answers.exam_participant_id").
		Joins("INNER JOIN exam_questions ON exam_questions.exam_id = exam_participants.exam_id AND exam_questions.question_id = answers.question_id").
		Where("answers.id = ?", answerID).
		Select("exam_questions.max_score", "exam_participants.status").
		Take(&result).Error
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// UpdateAnswerForEvaluation updates answer fields for evaluation.
func (r *Answer) UpdateAnswerForEvaluation(tx *gorm.DB, answer *models.Answer) error {
	tx = r.getTx(tx)
	return tx.Save(answer).Error
}

// UpdateScoreAwarded updates the participant's score_awarded atomically.
func (r *Answer) UpdateScoreAwarded(tx *gorm.DB, participantID types.ParticipantID, oldScore, newScore int32) error {
	tx = r.getTx(tx)
	return tx.Exec(
		`UPDATE exam_participants
		SET score_awarded = score_awarded - ? + ?
		WHERE id = ?`,
		oldScore, newScore, participantID,
	).Error
}

// CountUnevaluatedAnswers counts unevaluated answers for a participant.
func (r *Answer) CountUnevaluatedAnswers(tx *gorm.DB, participantID types.ParticipantID) (int64, error) {
	tx = r.getTx(tx)
	var count int64
	err := tx.Model(&models.Answer{}).
		Where("exam_participant_id = ? AND evaluated = ? AND answer IS NOT NULL", participantID, false).
		Count(&count).Error
	return count, err
}

// GetAllByParticipantID fetches all answers for a given participant.
func (r *Answer) GetAllByParticipantID(tx *gorm.DB, participantID types.ParticipantID) ([]models.Answer, error) {
	tx = r.getTx(tx)
	var answers []models.Answer
	err := tx.Where("exam_participant_id = ?", participantID).Find(&answers).Error
	return answers, err
}
