package repositories

import (
	"database/sql"
	"encoding/json"

	"gorm.io/gorm"

	"pariksha/common/pkg/proto"
	"pariksha/common/pkg/types"
	"pariksha/common/pkg/utils"
	"pariksha/question/internal/models"
	"pariksha/question/internal/structs"
)

type Question struct {
	db *gorm.DB
}

func NewQuestion(db *gorm.DB) *Question {
	return &Question{db: db}
}

// getTx returns tx if not nil, otherwise returns r.db
func (r *Question) getTx(tx *gorm.DB) *gorm.DB {
	if tx == nil {
		return r.db
	}
	return tx
}

func (r *Question) Transaction(fc func(tx *gorm.DB) error, opts ...*sql.TxOptions) error {
	return utils.TransactionHandler(r.db, fc, opts...)
}

// Create inserts a new question model into the database
func (r *Question) Create(question *models.Question, tx *gorm.DB) (*models.Question, error) {
	db := r.getTx(tx)

	if err := db.Create(question).Error; err != nil {
		return nil, err
	}

	return question, nil
}

// GetQuestionsByIDs fetches full question records by IDs
func (r *Question) GetQuestionsByIDs(tx *gorm.DB, ids []types.QuestionID) ([]models.Question, error) {
	tx = r.getTx(tx)

	if len(ids) == 0 {
		return []models.Question{}, nil
	}

	var questions []models.Question
	err := tx.Where("id IN ?", ids).Find(&questions).Error
	if err != nil {
		return nil, err
	}

	return questions, nil
}

// GetQuestionsByHashes fetches full question records by hashes
func (r *Question) GetQuestionsByHashes(tx *gorm.DB, hashes []string) ([]models.Question, error) {
	tx = r.getTx(tx)

	var questions []models.Question
	err := tx.Where("hash IN ?", hashes).Find(&questions).Error
	if err != nil {
		return nil, err
	}

	return questions, nil
}

func (r *Question) GetQuestionsMetaByIDs(tx *gorm.DB, ids []types.QuestionID) ([]models.Question, error) {
	tx = r.getTx(tx)

	if len(ids) == 0 {
		return []models.Question{}, nil
	}

	var questions []models.Question
	err := tx.Select("id, hash, question, type").Where("id IN ?", ids).Find(&questions).Error
	if err != nil {
		return nil, err
	}

	return questions, nil
}

func (r *Question) GetQuestionMetaByHashes(tx *gorm.DB, hashes []string) ([]models.Question, error) {
	tx = r.getTx(tx)

	var questions []models.Question
	err := tx.Select("id, hash, question, type").Where("hash IN ?", hashes).Find(&questions).Error
	if err != nil {
		return nil, err
	}

	return questions, nil
}

func (r *Question) GetQuestionIDsByHashes(tx *gorm.DB, hashes []string) ([]models.Question, error) {
	tx = r.getTx(tx)

	var questions []models.Question
	err := tx.Select("id, hash").Where("hash IN ?", hashes).Find(&questions).Error
	if err != nil {
		return nil, err
	}

	return questions, nil
}

func (r *Question) GetQuestionHashesByIDs(tx *gorm.DB, ids []types.QuestionID) ([]models.Question, error) {
	tx = r.getTx(tx)

	if len(ids) == 0 {
		return []models.Question{}, nil
	}

	var questions []models.Question
	err := tx.Select("id, hash").Where("id IN ?", ids).Find(&questions).Error
	if err != nil {
		return nil, err
	}

	return questions, nil
}

// UpdateQuestionHash updates question hash
func (r *Question) UpdateQuestionHash(tx *gorm.DB, questionID types.QuestionID, hash string) error {
	tx = r.getTx(tx)
	return tx.Model(&models.Question{}).Where("id = ?", questionID).Update("hash", hash).Error
}

// UpdatePaperIndegree updates paper_indegree by incrementing or decrementing it for a list of question IDs
func (r *Question) UpdatePaperIndegree(questionIDs []int64, increment bool, tx *gorm.DB) error {
	db := r.getTx(tx)

	op := "-"
	if increment {
		op = "+"
	}
	if err := db.Model(&models.Question{}).
		Where("id IN ?", questionIDs).
		UpdateColumn("paper_indegree", gorm.Expr("paper_indegree "+op+" ?", 1)).
		Error; err != nil {
		return err
	}
	return nil
}

// UpdateExamIndegree updates exam_indegree by incrementing or decrementing it for a list of question IDs
func (r *Question) UpdateExamIndegree(questionIDs []int64, increment bool, tx *gorm.DB) error {
	db := r.getTx(tx)

	op := "-"
	if increment {
		op = "+"
	}

	if err := db.Model(&models.Question{}).
		Where("id IN ?", questionIDs).
		UpdateColumn("exam_indegree", gorm.Expr("exam_indegree "+op+" ?", 1)).
		Error; err != nil {
		return err
	}
	return nil
}

// GetInputDefinitionsByHash fetches input definitions for a coding question by hash.
func (r *Question) GetInputDefinitionsByHash(tx *gorm.DB, questionHash string) ([]structs.InputDefinition, error) {
	tx = r.getTx(tx)

	type QueryResult struct {
		InputDefs []byte             `gorm:"column:input_defs"`
		Type      proto.QuestionType `gorm:"column:type"`
	}
	var queryRes QueryResult
	if err := tx.Model(&models.Question{}).
		Select("type, question->>'input_definitions' as input_defs").
		Where("hash = ?", questionHash).
		Take(&queryRes).Error; err != nil {
		return nil, err
	}

	// Only allow coding questions
	if queryRes.Type != proto.QuestionType_CODING {
		return nil, sql.ErrNoRows
	}

	var inputDefs []structs.InputDefinition
	if err := json.Unmarshal(queryRes.InputDefs, &inputDefs); err != nil {
		return nil, err
	}
	return inputDefs, nil
}

// GetQuestionsByHashes fetches full question records by hashes
func (r *Question) GetQuestionTypeByHash(tx *gorm.DB, hash string) (*models.Question, error) {
	tx = r.getTx(tx)

	var question models.Question
	err := tx.Select("id, type").Where("hash = ?", hash).Find(&question).Error
	if err != nil {
		return nil, err
	}

	return &question, nil
}

// GetInputDefinitionsLength retrieves the length of InputDefinitions array
func (r *Question) GetInputDefinitionsLength(tx *gorm.DB, questionID types.QuestionID) (int, error) {
	tx = r.getTx(tx)

	var length int
	err := tx.Raw(`
			SELECT jsonb_array_length(
					(Question->>'input_definitions')::jsonb
			) FROM questions WHERE id = ?
    `, questionID).Scan(&length).Error
	return length, err
}

// GetQuestionByHash fetches a question by its hash
func (r *Question) GetQuestionByHash(tx *gorm.DB, hash string) (*models.Question, error) {
	questions, err := r.GetQuestionsByHashes(tx, []string{hash})
	if err != nil {
		return nil, err
	}
	if len(questions) == 0 {
		return nil, gorm.ErrRecordNotFound
	}
	return &questions[0], err
}

// UpdateQuestion updates an existing question's fields
func (r *Question) UpdateQuestion(tx *gorm.DB, questionID types.QuestionID, updates map[string]any) error {
	tx = r.getTx(tx)
	return tx.Model(&models.Question{}).Where("id = ?", questionID).Updates(updates).Error
}
