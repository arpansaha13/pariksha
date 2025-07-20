package tests

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/metadata"

	"pariksha/common/pkg/types"
	"pariksha/paper/internal/models"
)

const defaultUserID types.UserID = 1

var defaultMetadata metadata.MD = metadata.MD{
	"user_id": []string{"1"},
}

func createTestPapers(t *testing.T, papers []models.Paper) []models.Paper {
	for i := range papers {
		if papers[i].Title == "" {
			papers[i].Title = fmt.Sprintf("Paper %d", i+1)
		}

		if papers[i].Hash == "" {
			papers[i].Hash = fmt.Sprintf("Paper Hash %d", i+1)
		}

		if papers[i].DurationMinutes == 0 {
			papers[i].DurationMinutes = 60
		}

		if papers[i].CreatedBy == 0 {
			papers[i].CreatedBy = defaultUserID
		}

		err := dbInst.Create(&papers[i]).Error
		require.NoError(t, err, "failed to create paper")

		perm := models.PaperPermission{
			PaperID: papers[i].ID,
			UserID:  papers[i].CreatedBy,
		}
		perm.SetWrite()

		err = dbInst.Create(&perm).Error
		require.NoError(t, err, "failed to assign paper permission")
	}
	return papers
}

func createDefaultTestPapers(t *testing.T, count int16) []models.Paper {
	require.Greater(t, count, int16(0))
	papers := make([]models.Paper, count)

	for i := range count {
		papers[i] = models.Paper{}
	}

	return createTestPapers(t, papers)
}

func createTestQuestions(t *testing.T, questions []models.PaperQuestion) []models.PaperQuestion {
	categoryOrderMap := make(map[int64]int16) // Tracks order per CategoryID

	for i := range questions {
		catID := int64(questions[i].CategoryID)

		// Assign Order per category
		categoryOrderMap[catID]++
		questions[i].Order = categoryOrderMap[catID]

		err := dbInst.Create(&questions[i]).Error
		require.NoError(t, err, "failed to create paper question")
	}
	return questions
}

func createTestCategories(t *testing.T, categories []models.PaperCategory) []models.PaperCategory {
	paperOrderMap := make(map[int64]int16) // Tracks order per CategoryID

	for i, c := range categories {
		paperID := int64(c.PaperID)

		// Assign Order per paper
		paperOrderMap[paperID]++
		categories[i].Order = paperOrderMap[paperID]

		err := dbInst.Create(&categories[i]).Error
		require.NoError(t, err, "failed to create paper category")
	}
	return categories
}
