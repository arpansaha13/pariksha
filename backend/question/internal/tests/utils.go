package tests

import (
	"testing"

	"github.com/stretchr/testify/require"

	"pariksha/common/pkg/types"
	"pariksha/common/pkg/utils/generate"
	"pariksha/question/internal/config/db"
	"pariksha/question/internal/models"
)

var questionIdToHashMap = make(map[types.QuestionID]string)

// createTestQuestions creates test questions in the database and manages their hashes
func createTestQuestions(t *testing.T, questions []models.Question) []models.Question {
	result := make([]models.Question, len(questions))
	for i := range questions {
		// Create question
		err := db.DB.Create(&questions[i]).Error
		require.NoError(t, err)

		// Check if hash exists for this ID, if not generate and store
		if hash, exists := questionIdToHashMap[questions[i].ID]; exists {
			questions[i].Hash = hash
		} else {
			questions[i].Hash = generate.HMACHash(int64(questions[i].ID))
			questionIdToHashMap[questions[i].ID] = questions[i].Hash
		}

		// Update hash in database
		err = db.DB.Model(&questions[i]).Update("hash", questions[i].Hash).Error
		require.NoError(t, err)

		result[i] = questions[i]
	}

	return result
}
