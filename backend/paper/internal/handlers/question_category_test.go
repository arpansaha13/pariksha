package handlers

import (
	"database/sql"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"pariksha/common/pkg/constants"
	"pariksha/common/pkg/proto"
	"pariksha/paper/internal/config/db"
	"pariksha/paper/internal/models"
)

func TestGetPaperCategories(t *testing.T) {
	tests := []struct {
		name         string
		setup        func(t *testing.T) (*models.Paper, []models.QuestionCategory)
		userID       int32
		expectedCode codes.Code
		validate     func(t *testing.T, categories []models.QuestionCategory, resp *proto.CategoryList)
	}{
		{
			name: "Success - Get categories",
			setup: func(t *testing.T) (*models.Paper, []models.QuestionCategory) {
				paper := createTestPaper(t, int(userID))
				var defaultCategory models.QuestionCategory
				require.NoError(t, db.DB.Where("paper_id = ?", paper.ID).First(&defaultCategory).Error)

				categories := []models.QuestionCategory{
					defaultCategory,
					{
						PaperID: sql.NullInt64{Int64: int64(paper.ID), Valid: true},
						Name:    "Category 2",
						Order:   2,
					},
					{
						PaperID: sql.NullInt64{Int64: int64(paper.ID), Valid: true},
						Name:    "Category 3",
						Order:   3,
					},
				}
				// Skip first category as it's already created
				categoriesToBeCreated := categories[1:]
				require.NoError(t, db.DB.Create(&categoriesToBeCreated).Error)
				return &paper, categories
			},
			userID:       userID,
			expectedCode: codes.OK,
			validate: func(t *testing.T, categories []models.QuestionCategory, resp *proto.CategoryList) {
				assert.Equal(t, len(categories), len(resp.Categories))
				for i, c := range resp.Categories {
					assert.Equal(t, categories[i].Name, c.Name)
					assert.Equal(t, int32(categories[i].Order), c.Order)
				}
			},
		},
		{
			name: "No access to paper",
			setup: func(t *testing.T) (*models.Paper, []models.QuestionCategory) {
				paper := createTestPaper(t, 2) // Create paper owned by different user
				return &paper, nil
			},
			userID:       userID,
			expectedCode: codes.PermissionDenied,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clearTables(t)
			paper, categories := tt.setup(t)

			ctx := createContextWithUserID(tt.userID)
			resp, err := client.GetPaperCategories(ctx, &proto.PaperRequest{
				PaperId: int32(paper.ID),
			})

			if tt.expectedCode != codes.OK {
				assert.Equal(t, tt.expectedCode, status.Code(err))
				return
			}

			require.NoError(t, err)
			require.NotNil(t, resp)
			tt.validate(t, categories, resp)
		})
	}
}

func TestCreateCategory(t *testing.T) {
	tests := []struct {
		name         string
		setup        func(t *testing.T) *models.Paper
		userID       int32
		expectedCode codes.Code
		validate     func(t *testing.T, resp *proto.CategoryResponse)
	}{
		{
			name: "Success - Create category",
			setup: func(t *testing.T) *models.Paper {
				paper := createTestPaper(t, int(userID))
				return &paper
			},
			userID:       userID,
			expectedCode: codes.OK,
			validate: func(t *testing.T, resp *proto.CategoryResponse) {
				assert.NotZero(t, resp.Id)
				assert.Equal(t, "Category 2", resp.Name) // Second category since createTestPaper creates one
				assert.Equal(t, int32(2), resp.Order)
			},
		},
		{
			name: "Not paper owner",
			setup: func(t *testing.T) *models.Paper {
				paper := createTestPaper(t, 2) // Different user
				return &paper
			},
			userID:       userID,
			expectedCode: codes.PermissionDenied,
		},
		{
			name: "Paper not found",
			setup: func(t *testing.T) *models.Paper {
				return &models.Paper{ID: 999}
			},
			userID:       userID,
			expectedCode: codes.NotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clearTables(t)
			paper := tt.setup(t)

			ctx := createContextWithUserID(tt.userID)
			resp, err := client.CreateCategory(ctx, &proto.CreateCategoryRequest{
				PaperId: int32(paper.ID),
			})

			if tt.expectedCode != codes.OK {
				assert.Equal(t, tt.expectedCode, status.Code(err))
				return
			}

			require.NoError(t, err)
			require.NotNil(t, resp)
			tt.validate(t, resp)
		})
	}
}

func TestUpdateCategory(t *testing.T) {
	tests := []struct {
		name         string
		setup        func(t *testing.T) *models.QuestionCategory
		userID       int32
		newName      string
		expectedCode codes.Code
		validate     func(t *testing.T, category *models.QuestionCategory)
	}{
		{
			name: "Success - Update unlocked category",
			setup: func(t *testing.T) *models.QuestionCategory {
				paper := createTestPaper(t, int(userID))
				category := models.QuestionCategory{
					PaperID: sql.NullInt64{Int64: int64(paper.ID), Valid: true},
					Name:    "Original Name",
					Order:   1,
					Locked:  false,
				}
				require.NoError(t, db.DB.Create(&category).Error)
				return &category
			},
			userID:       userID,
			newName:      "Updated Name",
			expectedCode: codes.OK,
			validate: func(t *testing.T, category *models.QuestionCategory) {
				var updated models.QuestionCategory
				require.NoError(t, db.DB.First(&updated, category.ID).Error)
				assert.Equal(t, "Updated Name", updated.Name)
				assert.Equal(t, category.ID, updated.ID) // Same record was updated
			},
		},
		{
			name: "Success - Update locked category creates new record",
			setup: func(t *testing.T) *models.QuestionCategory {
				paper := createTestPaper(t, int(userID))
				category := models.QuestionCategory{
					PaperID: sql.NullInt64{Int64: int64(paper.ID), Valid: true},
					Name:    "Original Name",
					Order:   1,
					Locked:  true,
				}
				require.NoError(t, db.DB.Create(&category).Error)
				return &category
			},
			userID:       userID,
			newName:      "Updated Name",
			expectedCode: codes.OK,
			validate: func(t *testing.T, category *models.QuestionCategory) {
				// Original category should be unlinked
				var original models.QuestionCategory
				require.NoError(t, db.DB.First(&original, category.ID).Error)
				assert.Equal(t, "Original Name", original.Name)
				assert.True(t, original.Locked)
				assert.Zero(t, original.PaperID) // Unlinked

				// New category should be created
				var newCategory models.QuestionCategory
				require.NoError(t, db.DB.Where("paper_id = ? AND name = ?", category.PaperID, "Updated Name").First(&newCategory).Error)
				assert.NotEqual(t, category.ID, newCategory.ID) // Different record
				assert.Equal(t, "Updated Name", newCategory.Name)
				assert.Equal(t, category.Order, newCategory.Order)
				assert.False(t, newCategory.Locked)
			},
		},
		{
			name: "Success - Update category",
			setup: func(t *testing.T) *models.QuestionCategory {
				paper := createTestPaper(t, int(userID))
				category := models.QuestionCategory{
					PaperID: sql.NullInt64{Int64: int64(paper.ID), Valid: true},
					Name:    "Original Name",
					Order:   1,
				}
				require.NoError(t, db.DB.Create(&category).Error)
				return &category
			},
			userID:       userID,
			newName:      "Updated Name",
			expectedCode: codes.OK,
			validate: func(t *testing.T, category *models.QuestionCategory) {
				var updated models.QuestionCategory
				require.NoError(t, db.DB.First(&updated, category.ID).Error)
				assert.Equal(t, "Updated Name", updated.Name)
			},
		},
		{
			name: "Not paper owner",
			setup: func(t *testing.T) *models.QuestionCategory {
				paper := createTestPaper(t, 2) // Different user
				category := models.QuestionCategory{
					PaperID: sql.NullInt64{Int64: int64(paper.ID), Valid: true},
					Name:    "Original Name",
					Order:   1,
				}
				require.NoError(t, db.DB.Create(&category).Error)
				return &category
			},
			userID:       userID,
			newName:      "Updated Name",
			expectedCode: codes.PermissionDenied,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clearTables(t)
			category := tt.setup(t)

			ctx := createContextWithUserID(tt.userID)
			_, err := client.UpdateCategory(ctx, &proto.UpdateCategoryRequest{
				CategoryId: int32(category.ID),
				Name:       tt.newName,
			})

			if tt.expectedCode != codes.OK {
				assert.Equal(t, tt.expectedCode, status.Code(err))
				return
			}

			require.NoError(t, err)
			tt.validate(t, category)
		})
	}
}

func TestReorderCategories(t *testing.T) {
	tests := []struct {
		name         string
		setup        func(t *testing.T) (*models.Paper, []models.QuestionCategory)
		userID       int32
		expectedCode codes.Code
		validate     func(t *testing.T, categories []models.QuestionCategory)
	}{
		{
			name: "Success - Reorder categories",
			setup: func(t *testing.T) (*models.Paper, []models.QuestionCategory) {
				paper := createTestPaper(t, int(userID))
				var defaultCategory models.QuestionCategory
				require.NoError(t, db.DB.Where("paper_id = ?", paper.ID).First(&defaultCategory).Error)

				additionalCategory := models.QuestionCategory{
					PaperID: sql.NullInt64{Int64: int64(paper.ID), Valid: true},
					Name:    "Category 2",
					Order:   2,
				}
				require.NoError(t, db.DB.Create(&additionalCategory).Error)

				categories := []models.QuestionCategory{defaultCategory, additionalCategory}
				return &paper, categories
			},
			userID:       userID,
			expectedCode: codes.OK,
			validate: func(t *testing.T, categories []models.QuestionCategory) {
				var updated []models.QuestionCategory
				require.NoError(t, db.DB.Where("paper_id = ?", categories[0].PaperID).
					Order("\"order\"").Find(&updated).Error)

				assert.Equal(t, len(categories), len(updated))
				assert.Equal(t, categories[1].ID, updated[0].ID) // Category 2 should be first
				assert.Equal(t, categories[0].ID, updated[1].ID) // Default category should be second
				assert.Equal(t, 1, updated[0].Order)
				assert.Equal(t, 2, updated[1].Order)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clearTables(t)
			paper, categories := tt.setup(t)

			ctx := createContextWithUserID(tt.userID)
			_, err := client.ReorderCategories(ctx, &proto.ReorderCategoriesRequest{
				PaperId:     int32(paper.ID),
				CategoryIds: []int32{int32(categories[1].ID), int32(categories[0].ID)}, // Reverse order
			})

			if tt.expectedCode != codes.OK {
				assert.Equal(t, tt.expectedCode, status.Code(err))
				return
			}

			require.NoError(t, err)
			tt.validate(t, categories)
		})
	}
}

func TestDeleteCategory(t *testing.T) {
	tests := []struct {
		name         string
		setup        func(t *testing.T) *models.QuestionCategory
		userID       int32
		expectedCode codes.Code
		validate     func(t *testing.T, categoryID int)
	}{
		{
			name: "Success - Delete category when multiple exist",
			setup: func(t *testing.T) *models.QuestionCategory {
				paper := createTestPaper(t, int(userID))
				// createTestPaper already creates one category
				category := models.QuestionCategory{
					PaperID: sql.NullInt64{Int64: int64(paper.ID), Valid: true},
					Name:    "Test Category",
					Order:   2,
				}
				require.NoError(t, db.DB.Create(&category).Error)
				return &category
			},
			userID:       userID,
			expectedCode: codes.OK,
			validate: func(t *testing.T, categoryID int) {
				var category models.QuestionCategory
				err := db.DB.First(&category, categoryID).Error
				assert.Error(t, err)
			},
		},
		{
			name: "Fail - Cannot delete last category",
			setup: func(t *testing.T) *models.QuestionCategory {
				paper := createTestPaper(t, int(userID))
				var category models.QuestionCategory
				require.NoError(t, db.DB.Where("paper_id = ?", paper.ID).First(&category).Error)
				return &category
			},
			userID:       userID,
			expectedCode: codes.FailedPrecondition,
			validate: func(t *testing.T, categoryID int) {
				var category models.QuestionCategory
				err := db.DB.First(&category, categoryID).Error
				assert.NoError(t, err)
			},
		},
		{
			name: "Not paper owner",
			setup: func(t *testing.T) *models.QuestionCategory {
				paper := createTestPaper(t, 2) // Different user
				category := models.QuestionCategory{
					PaperID: sql.NullInt64{Int64: int64(paper.ID), Valid: true},
					Name:    "Test Category",
					Order:   2,
				}
				require.NoError(t, db.DB.Create(&category).Error)
				return &category
			},
			userID:       userID,
			expectedCode: codes.PermissionDenied,
		},
		{
			name: "Success - Delete locked category unlinks it",
			setup: func(t *testing.T) *models.QuestionCategory {
				paper := createTestPaper(t, int(userID))
				// Default category from createTestPaper
				var defaultCategory models.QuestionCategory
				require.NoError(t, db.DB.Where("paper_id = ?", paper.ID).First(&defaultCategory).Error)

				// Create locked category
				category := models.QuestionCategory{
					PaperID: sql.NullInt64{Int64: int64(paper.ID), Valid: true},
					Name:    "Test Category",
					Order:   2,
					Locked:  true,
				}
				require.NoError(t, db.DB.Create(&category).Error)

				// Add a locked question to this category
				question := models.Question{
					PaperID:    sql.NullInt64{Int64: int64(paper.ID), Valid: true},
					CategoryID: category.ID,
					Type:       constants.QUESTION_TYPE_MCQ,
					Question:   json.RawMessage(`{"statement":"Test MCQ","options":["A","B"]}`),
					MaxScore:   5,
					Locked:     true,
				}
				require.NoError(t, db.DB.Create(&question).Error)

				return &category
			},
			userID:       userID,
			expectedCode: codes.OK,
			validate: func(t *testing.T, categoryID int) {
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
			name: "Success - Delete category with mix of locked and unlocked questions",
			setup: func(t *testing.T) *models.QuestionCategory {
				paper := createTestPaper(t, int(userID))
				// Default category from createTestPaper
				var defaultCategory models.QuestionCategory
				require.NoError(t, db.DB.Where("paper_id = ?", paper.ID).First(&defaultCategory).Error)

				// Create category
				// Note: A locked category may have both locked and unlocked questions.
				// But an unlocked category can only have unlocked questions.
				category := models.QuestionCategory{
					PaperID: sql.NullInt64{Int64: int64(paper.ID), Valid: true},
					Name:    "Test Category",
					Order:   2,
					Locked:  true,
				}
				require.NoError(t, db.DB.Create(&category).Error)

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
			userID:       userID,
			expectedCode: codes.OK,
			validate: func(t *testing.T, categoryID int) {
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
			_, err := client.DeleteCategory(ctx, &proto.CategoryRequest{
				CategoryId: int32(category.ID),
			})

			if tt.expectedCode != codes.OK {
				assert.Equal(t, tt.expectedCode, status.Code(err))
				return
			}

			require.NoError(t, err)
			tt.validate(t, category.ID)
		})
	}
}
