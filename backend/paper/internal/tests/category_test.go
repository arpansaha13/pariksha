package tests

import (
	"database/sql"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"

	"pariksha/common/pkg/constants"
	"pariksha/common/pkg/models"
	"pariksha/common/pkg/proto"
	"pariksha/common/pkg/types"
	"pariksha/common/pkg/utils/generate"
	"pariksha/common/pkg/utils/testrunner"
	"pariksha/paper/internal/config/db"
)

func TestGetPaperCategories(t *testing.T) {
	tests := []CategoryListCase{
		{
			BaseTestCase: BaseTestCase{
				name:         "Success - Get categories",
				userID:       userID,
				expectedCode: codes.OK,
			},
			setup: func(t *testing.T) (*models.Paper, []models.QuestionCategory) {
				paper := createTestPaper(t, userID)
				categories := createDefaultTestCategories(t, paper.ID, 3)
				return &paper, categories
			},
			validate: func(t *testing.T, categories []models.QuestionCategory, resp *proto.CategoryList) {
				assert.EqualValues(t, len(categories), len(resp.Categories))
				for i, c := range resp.Categories {
					assert.Equal(t, categories[i].Name, c.Name)
					assert.EqualValues(t, categories[i].Order, c.Order)
				}
			},
		},
		{
			BaseTestCase: BaseTestCase{
				name:         "No access to paper",
				userID:       userID,
				expectedCode: codes.PermissionDenied,
			},
			setup: func(t *testing.T) (*models.Paper, []models.QuestionCategory) {
				paper := createTestPaper(t, 2) // Owned by different user
				return &paper, nil
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clearTables(t)
			paper, categories := tt.setup(t)

			ctx := createContextWithUserID(tt.userID)
			testrunner.Runner(t, ctx, tt.expectedCode,
				&proto.PaperRequest{PaperHash: paper.PaperHash.Hash},
				client.GetPaperCategories,
				func(t *testing.T, resp *proto.CategoryList) {
					if tt.validate != nil {
						tt.validate(t, categories, resp)
					}
				})
		})
	}
}

func TestCreateCategory(t *testing.T) {
	tests := []CreateCategoryCase{
		{
			BaseTestCase: BaseTestCase{
				name:         "Success - Create category",
				userID:       userID,
				expectedCode: codes.OK,
			},
			setup: func(t *testing.T) *models.Paper {
				paper := createTestPaper(t, userID)
				return &paper
				return &paper
			},
			validate: func(t *testing.T, resp *proto.CategoryResponse) {
				assert.NotZero(t, resp.CategoryId)
				assert.Equal(t, "Category 1", resp.Name)
				assert.EqualValues(t, 1, resp.Order)
			},
		},
		{
			BaseTestCase: BaseTestCase{
				name:         "Not paper owner",
				userID:       userID,
				expectedCode: codes.PermissionDenied,
			},
			setup: func(t *testing.T) *models.Paper {
				paper := createTestPaper(t, 2) // Different user
				return &paper
			},
		},
		{
			BaseTestCase: BaseTestCase{
				name:         "Paper not found",
				userID:       userID,
				expectedCode: codes.NotFound,
			},
			setup: func(t *testing.T) *models.Paper {
				return &models.Paper{
					ID: 999,
					PaperHash: models.PaperHash{
						Hash: generate.HMACHash(999),
						ID:   999,
					},
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clearTables(t)
			paper := tt.setup(t)

			ctx := createContextWithUserID(tt.userID)
			testrunner.Runner(t, ctx, tt.expectedCode,
				&proto.CreateCategoryRequest{PaperHash: paper.PaperHash.Hash},
				client.CreateCategory,
				func(t *testing.T, resp *proto.CategoryResponse) {
					if tt.validate != nil {
						tt.validate(t, resp)
					}
				})
		})
	}
}

func TestUpdateCategory(t *testing.T) {
	tests := []UpdateCategoryCase{
		{
			BaseTestCase: BaseTestCase{
				name:         "Success - Update unlocked category",
				userID:       userID,
				expectedCode: codes.OK,
			},
			setup: func(t *testing.T) *models.QuestionCategory {
				paper := createTestPaper(t, userID)
				category := createDefaultTestCategory(t, paper.ID)
				return &category
			},
			newName: "Updated Name",
			validate: func(t *testing.T, category *models.QuestionCategory) {
				var updated models.QuestionCategory
				require.NoError(t, db.DB.First(&updated, category.ID).Error)
				assert.Equal(t, "Updated Name", updated.Name)
				assert.Equal(t, category.ID, updated.ID) // Same record was updated
			},
		},
		{
			BaseTestCase: BaseTestCase{
				name:         "Success - Update locked category creates new record",
				userID:       userID,
				expectedCode: codes.OK,
			},
			setup: func(t *testing.T) *models.QuestionCategory {
				paper := createTestPaper(t, userID)
				categories := createTestCategories(t, []models.QuestionCategory{
					{
						PaperID: sql.NullInt64{Int64: int64(paper.ID), Valid: true},
						Locked:  true,
					},
				})
				return &categories[0]
			},
			newName: "Updated Name",
			validate: func(t *testing.T, category *models.QuestionCategory) {
				// Original category should be unlinked
				var original models.QuestionCategory
				require.NoError(t, db.DB.First(&original, category.ID).Error)
				assert.Equal(t, "Category 1", original.Name)
				assert.True(t, original.Locked)
				assert.Zero(t, original.PaperID) // Unlinked

				// New category should be created
				var newCategory models.QuestionCategory
				require.NoError(t, db.DB.Where("paper_id = ? AND name = ?", category.PaperID, "Updated Name").First(&newCategory).Error)
				assert.NotEqualValues(t, category.ID, newCategory.ID) // Different record
				assert.Equal(t, "Updated Name", newCategory.Name)
				assert.Equal(t, category.Order, newCategory.Order)
				assert.False(t, newCategory.Locked)
			},
		},
		{
			BaseTestCase: BaseTestCase{
				name:         "Success - Update locked category with questions",
				userID:       userID,
				expectedCode: codes.OK,
			},
			setup: func(t *testing.T) *models.QuestionCategory {
				paper := createTestPaper(t, userID)
				categories := createTestCategories(t, []models.QuestionCategory{
					{
						PaperID: sql.NullInt64{Int64: int64(paper.ID), Valid: true},
						Locked:  true,
					},
				})
				category := &categories[0]

				createTestQuestions(t, []models.Question{
					{
						PaperID:    sql.NullInt64{Int64: int64(paper.ID), Valid: true},
						CategoryID: category.ID,
						Type:       constants.QUESTION_TYPE_MCQ,
						Question:   json.RawMessage(`{"statement":"Test Question 1","options":["A","B","C"]}`),
						Locked:     true,
					},
					{
						PaperID:    sql.NullInt64{Int64: int64(paper.ID), Valid: true},
						CategoryID: category.ID,
						Type:       constants.QUESTION_TYPE_MCQ,
						Question:   json.RawMessage(`{"statement":"Test Question 2","options":["A","B","C"]}`),
						Locked:     true,
					},
				})

				return category
			},
			newName: "Updated Name",
			validate: func(t *testing.T, category *models.QuestionCategory) {
				// Original category should be unlinked from paper
				var original models.QuestionCategory
				require.NoError(t, db.DB.First(&original, category.ID).Error)
				assert.Equal(t, "Category 1", original.Name)
				assert.True(t, original.Locked)
				assert.False(t, original.PaperID.Valid) // Should be unlinked

				// New category should be created
				var newCategory models.QuestionCategory
				require.NoError(t, db.DB.Where("paper_id = ? AND name = ?", category.PaperID, "Updated Name").First(&newCategory).Error)
				assert.NotEqualValues(t, category.ID, newCategory.ID)
				assert.Equal(t, "Updated Name", newCategory.Name)
				assert.Equal(t, category.Order, newCategory.Order)
				assert.False(t, newCategory.Locked)

				// Questions should be moved to new category
				var questions []models.Question
				require.NoError(t, db.DB.Where("category_id = ?", newCategory.ID).Find(&questions).Error)
				assert.EqualValues(t, 2, len(questions))
				for _, q := range questions {
					assert.Equal(t, newCategory.ID, q.CategoryID)
					assert.True(t, q.PaperID.Valid)
					assert.Equal(t, category.PaperID.Int64, q.PaperID.Int64)
				}

				// No questions should remain in old category
				count := int64(0)
				require.NoError(t, db.DB.Model(&models.Question{}).Where("category_id = ?", category.ID).Count(&count).Error)
				assert.EqualValues(t, 0, count)
			},
		},
		{
			BaseTestCase: BaseTestCase{
				name:         "Success - Update category",
				userID:       userID,
				expectedCode: codes.OK,
			},
			setup: func(t *testing.T) *models.QuestionCategory {
				paper := createTestPaper(t, userID)
				category := createDefaultTestCategory(t, paper.ID)
				return &category
			},
			newName: "Updated Name",
			validate: func(t *testing.T, category *models.QuestionCategory) {
				var updated models.QuestionCategory
				require.NoError(t, db.DB.First(&updated, category.ID).Error)
				assert.Equal(t, "Updated Name", updated.Name)
			},
		},
		{
			BaseTestCase: BaseTestCase{
				name:         "Not paper owner",
				userID:       userID,
				expectedCode: codes.PermissionDenied,
			},
			setup: func(t *testing.T) *models.QuestionCategory {
				paper := createTestPaper(t, 2) // Different user
				category := createDefaultTestCategory(t, paper.ID)
				return &category
			},
			newName: "Updated Name",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clearTables(t)
			category := tt.setup(t)

			ctx := createContextWithUserID(tt.userID)
			testrunner.Runner(t, ctx, tt.expectedCode,
				&proto.UpdateCategoryRequest{
					CategoryId: int64(category.ID),
					Name:       tt.newName,
				},
				client.UpdateCategory,
				func(t *testing.T, resp *proto.Empty) {
					if tt.validate != nil {
						tt.validate(t, category)
					}
				})
		})
	}
}

func TestReorderCategories(t *testing.T) {
	tests := []ReorderCategoriesCase{
		{
			BaseTestCase: BaseTestCase{
				name:         "Success - Reorder categories",
				userID:       userID,
				expectedCode: codes.OK,
			},
			setup: func(t *testing.T) (*models.Paper, []models.QuestionCategory) {
				paper := createTestPaper(t, userID)
				categories := createDefaultTestCategories(t, paper.ID, 2)
				return &paper, categories
			},
			validate: func(t *testing.T, categories []models.QuestionCategory) {
				var updated []models.QuestionCategory
				require.NoError(t, db.DB.Where("paper_id = ?", categories[0].PaperID).
					Order("\"order\"").Find(&updated).Error)

				assert.Equal(t, len(categories), len(updated))
				assert.Equal(t, categories[1].ID, updated[0].ID) // Category 2 should be first
				assert.Equal(t, categories[0].ID, updated[1].ID) // Default category should be second
				assert.EqualValues(t, 1, updated[0].Order)
				assert.EqualValues(t, 2, updated[1].Order)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clearTables(t)
			paper, categories := tt.setup(t)

			ctx := createContextWithUserID(tt.userID)
			testrunner.Runner(t, ctx, tt.expectedCode,
				&proto.ReorderCategoriesRequest{
					PaperHash:   paper.PaperHash.Hash,
					CategoryIds: []int64{int64(categories[1].ID), int64(categories[0].ID)}, // Reverse order
				},
				client.ReorderCategories,
				func(t *testing.T, resp *proto.Empty) {
					if tt.validate != nil {
						tt.validate(t, categories)
					}
				})
		})
	}
}

func TestDeleteCategory(t *testing.T) {
	tests := []DeleteCategoryCase{
		{
			BaseTestCase: BaseTestCase{
				name:         "Success - Delete category when multiple exist",
				userID:       userID,
				expectedCode: codes.OK,
			},
			setup: func(t *testing.T) *models.QuestionCategory {
				paper := createTestPaper(t, userID)
				categories := createDefaultTestCategories(t, paper.ID, 2)
				return &categories[0]
			},
			validate: func(t *testing.T, categoryID types.CategoryID) {
				var category models.QuestionCategory
				err := db.DB.Take(&category, categoryID).Error
				assert.Error(t, err)
			},
		},
		{
			BaseTestCase: BaseTestCase{
				name:         "Fail - Cannot delete last category",
				userID:       userID,
				expectedCode: codes.FailedPrecondition,
			},
			setup: func(t *testing.T) *models.QuestionCategory {
				paper := createTestPaper(t, userID)
				category := createDefaultTestCategory(t, paper.ID)
				return &category
			},
			validate: func(t *testing.T, categoryID types.CategoryID) {
				var category models.QuestionCategory
				err := db.DB.First(&category, categoryID).Error
				assert.NoError(t, err)
			},
		},
		{
			BaseTestCase: BaseTestCase{
				name:         "Not paper owner",
				userID:       userID,
				expectedCode: codes.PermissionDenied,
			},
			setup: func(t *testing.T) *models.QuestionCategory {
				paper := createTestPaper(t, 2) // Different user
				category := createDefaultTestCategory(t, paper.ID)
				return &category
			},
		},
		{
			BaseTestCase: BaseTestCase{
				name:         "Success - Delete locked category unlinks it",
				userID:       userID,
				expectedCode: codes.OK,
			},
			setup: func(t *testing.T) *models.QuestionCategory {

				paper := createTestPaper(t, userID)
				categories := createTestCategories(t, []models.QuestionCategory{
					{
						PaperID: sql.NullInt64{Int64: int64(paper.ID), Valid: true},
					},
					{
						PaperID: sql.NullInt64{Int64: int64(paper.ID), Valid: true},
						Locked:  true,
					},
				})
				category := &categories[1]

				// Create test question
				createTestQuestions(t, []models.Question{
					{
						PaperID:    sql.NullInt64{Int64: int64(paper.ID), Valid: true},
						CategoryID: category.ID,
						Type:       constants.QUESTION_TYPE_MCQ,
						Question:   json.RawMessage(`{"statement":"Test MCQ","options":["A","B","C"]}`),
						Locked:     true,
					},
				})

				return category
			},
			validate: func(t *testing.T, categoryID types.CategoryID) {
				// Verify category still exists but is unlinked
				var category models.QuestionCategory
				require.NoError(t, db.DB.First(&category, categoryID).Error)
				assert.True(t, category.Locked)
				assert.Zero(t, category.PaperID) // Should be unlinked

				// Verify question is also unlinked but not deleted
				var question models.Question
				require.NoError(t, db.DB.Where("category_id = ?", categoryID).First(&question).Error)
				assert.True(t, question.Locked)
				assert.False(t, question.PaperID.Valid) // Should be unlinked
			},
		},
		{
			BaseTestCase: BaseTestCase{
				name:         "Success - Delete category with mix of locked and unlocked questions",
				userID:       userID,
				expectedCode: codes.OK,
			},
			setup: func(t *testing.T) *models.QuestionCategory {
				paper := createTestPaper(t, userID)

				// Note: A locked category may have both locked and unlocked questions.
				// But an unlocked category can only have unlocked questions.
				categories := createTestCategories(t, []models.QuestionCategory{
					{
						PaperID: sql.NullInt64{Int64: int64(paper.ID), Valid: true},
					},
					{
						PaperID: sql.NullInt64{Int64: int64(paper.ID), Valid: true},
						Locked:  true,
					},
				})
				category := categories[1]

				// Add locked and unlocked questions
				questions := []models.Question{
					{
						PaperID:    sql.NullInt64{Int64: int64(paper.ID), Valid: true},
						CategoryID: category.ID,
						Type:       constants.QUESTION_TYPE_MCQ,
						Question:   json.RawMessage(`{"statement":"Locked MCQ","options":["A","B"]}`),
						MaxScore:   5,
						Locked:     true,
					},
					{
						PaperID:    sql.NullInt64{Int64: int64(paper.ID), Valid: true},
						CategoryID: category.ID,
						Type:       constants.QUESTION_TYPE_MCQ,
						Question:   json.RawMessage(`{"statement":"Unlocked MCQ","options":["A","B"]}`),
						MaxScore:   5,
						Locked:     false,
					},
				}
				require.NoError(t, db.DB.Create(&questions).Error)

				return &category
			},
			validate: func(t *testing.T, categoryID types.CategoryID) {
				// Verify category still exists but is unlinked
				var category models.QuestionCategory
				require.NoError(t, db.DB.First(&category, categoryID).Error)
				assert.True(t, category.Locked)
				assert.Zero(t, category.PaperID) // Should be unlinked

				// Verify locked question is unlinked but exists
				var lockedQuestion models.Question
				require.NoError(t, db.DB.Where("category_id = ? AND locked = true", categoryID).First(&lockedQuestion).Error)
				assert.Zero(t, lockedQuestion.PaperID) // Should be unlinked

				// Verify unlocked question is deleted
				var unlockedQuestion models.Question
				err := db.DB.Where("category_id = ? AND locked = false", categoryID).First(&unlockedQuestion).Error
				assert.Error(t, err) // Should be deleted
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clearTables(t)
			category := tt.setup(t)

			ctx := createContextWithUserID(tt.userID)
			testrunner.Runner(t, ctx, tt.expectedCode,
				&proto.CategoryRequest{CategoryId: int64(category.ID)},
				client.DeleteCategory,
				func(t *testing.T, resp *proto.Empty) {
					if tt.validate != nil {
						tt.validate(t, category.ID)
					}
				})
		})
	}
}
