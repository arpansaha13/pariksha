package tests

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/types/known/emptypb"

	"pariksha/common/pkg/constants"
	"pariksha/common/pkg/proto"
	"pariksha/common/pkg/test"
	"pariksha/common/pkg/types"
	"pariksha/paper/internal/config/env"
	"pariksha/paper/internal/models"
)

func TestGetPaperCategories(t *testing.T) {
	type SetupReturn struct {
		Paper      models.Paper
		Categories []models.PaperCategory
	}

	testCases := []test.TestCase[*proto.PaperRequest, *proto.PaperCategoryList, *SetupReturn]{
		{
			Name:     "Get multiple categories for paper",
			Metadata: defaultMetadata,
			Setup: func(t *testing.T) *SetupReturn {
				papers := createDefaultTestPapers(t, 1)

				categories := createTestCategories(t, []models.PaperCategory{
					{
						PaperID:    papers[0].ID,
						CategoryID: 1,
					},
					{
						PaperID:    papers[0].ID,
						CategoryID: 2,
					},
				})

				return &SetupReturn{
					Paper:      papers[0],
					Categories: categories,
				}
			},
			GetRequest: func(setupData *SetupReturn) *proto.PaperRequest {
				return &proto.PaperRequest{
					PaperHash: setupData.Paper.Hash,
				}
			},
			ExpectedCode: codes.OK,
			Validate: func(t *testing.T, resp *proto.PaperCategoryList, setupData *SetupReturn) {
				require.Equal(t, len(setupData.Categories), len(resp.Categories))
				for i, cat := range resp.Categories {
					assert.Equal(t, int64(setupData.Categories[i].CategoryID), cat.Id)
					assert.Equal(t, int32(setupData.Categories[i].Order), cat.Order)
				}
			},
		},
		{
			Name:     "Paper with no categories",
			Metadata: defaultMetadata,
			Setup: func(t *testing.T) *SetupReturn {
				papers := createDefaultTestPapers(t, 1)

				return &SetupReturn{
					Paper:      papers[0],
					Categories: []models.PaperCategory{},
				}
			},
			GetRequest: func(setupData *SetupReturn) *proto.PaperRequest {
				return &proto.PaperRequest{
					PaperHash: setupData.Paper.Hash,
				}
			},
			ExpectedCode: codes.OK,
			Validate: func(t *testing.T, resp *proto.PaperCategoryList, setupData *SetupReturn) {
				assert.Empty(t, resp.Categories)
			},
		},
		{
			Name:     "Non-existent paper hash",
			Metadata: defaultMetadata,
			GetRequest: func(setupData *SetupReturn) *proto.PaperRequest {
				return &proto.PaperRequest{
					PaperHash: "nonexistent",
				}
			},
			ExpectedCode: codes.PermissionDenied,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			clearTables(t)
			test.Runner(t, tc, client.GetPaperCategories)
		})
	}
}

func TestGetPaperCategoriesMeta(t *testing.T) {
	md := metadata.MD{
		constants.X_EXAM_API_TOKEN: []string{env.EXAM_API_TOKEN},
	}

	type SetupReturn struct {
		Paper      models.Paper
		Categories []models.PaperCategory
	}

	testCases := []test.TestCase[*proto.PaperRequest, *proto.PaperCategoriesMeta, *SetupReturn]{
		{
			Name:     "Get categories metadata for paper",
			Metadata: md,
			Setup: func(t *testing.T) *SetupReturn {
				papers := createDefaultTestPapers(t, 1)

				categories := createTestCategories(t, []models.PaperCategory{
					{
						PaperID:    papers[0].ID,
						CategoryID: types.CategoryID(1),
						Order:      1,
					},
					{
						PaperID:    papers[0].ID,
						CategoryID: types.CategoryID(2),
						Order:      2,
					},
				})

				return &SetupReturn{
					Paper:      papers[0],
					Categories: categories,
				}
			},
			GetRequest: func(setupData *SetupReturn) *proto.PaperRequest {
				return &proto.PaperRequest{
					PaperHash: setupData.Paper.Hash,
				}
			},
			ExpectedCode: codes.OK,
			Validate: func(t *testing.T, resp *proto.PaperCategoriesMeta, setupData *SetupReturn) {
				require.Equal(t, len(setupData.Categories), len(resp.Categories))
				for i, cat := range resp.Categories {
					assert.Equal(t, int64(setupData.Categories[i].CategoryID), cat.Id)
					assert.Equal(t, int32(setupData.Categories[i].Order), cat.Order)
				}
			},
		},
		{
			Name:     "Non-existent paper hash",
			Metadata: md,
			GetRequest: func(setupData *SetupReturn) *proto.PaperRequest {
				return &proto.PaperRequest{
					PaperHash: "nonexistent",
				}
			},
			ExpectedCode: codes.NotFound,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			clearTables(t)
			test.Runner(t, tc, client.GetPaperCategoriesMeta)
		})
	}
}

func TestReorderCategories(t *testing.T) {
	type SetupReturn struct {
		Paper      models.Paper
		Categories []models.PaperCategory
	}

	testCases := []test.TestCase[*proto.ReorderPaperCategoriesRequest, *emptypb.Empty, *SetupReturn]{
		{
			Name:     "Reorder multiple categories",
			Metadata: defaultMetadata,
			Setup: func(t *testing.T) *SetupReturn {
				papers := createDefaultTestPapers(t, 1)

				categories := createTestCategories(t, []models.PaperCategory{
					{
						PaperID:    papers[0].ID,
						CategoryID: 1,
					},
					{
						PaperID:    papers[0].ID,
						CategoryID: 2,
					},
					{
						PaperID:    papers[0].ID,
						CategoryID: 3,
					},
				})

				return &SetupReturn{
					Paper:      papers[0],
					Categories: categories,
				}
			},
			GetRequest: func(setupData *SetupReturn) *proto.ReorderPaperCategoriesRequest {
				// Reverse the order of categories
				categoryIDs := make([]int64, len(setupData.Categories))
				for i, cat := range setupData.Categories {
					categoryIDs[len(setupData.Categories)-1-i] = int64(cat.CategoryID)
				}
				return &proto.ReorderPaperCategoriesRequest{
					PaperHash:   setupData.Paper.Hash,
					CategoryIds: categoryIDs,
				}
			},
			ExpectedCode: codes.OK,
			Validate: func(t *testing.T, resp *emptypb.Empty, setupData *SetupReturn) {
				var categories []models.PaperCategory
				err := dbInst.Where("paper_id = ?", setupData.Paper.ID).
					Order("\"order\" asc").
					Find(&categories).Error
				require.NoError(t, err)
				require.Len(t, categories, len(setupData.Categories))

				// Verify reversed order
				for i, cat := range categories {
					expectedID := setupData.Categories[len(setupData.Categories)-1-i].CategoryID
					assert.Equal(t, expectedID, cat.CategoryID)
					assert.Equal(t, int16(i+1), cat.Order)
				}
			},
		},
		{
			Name:     "Invalid category IDs length",
			Metadata: defaultMetadata,
			Setup: func(t *testing.T) *SetupReturn {
				papers := createDefaultTestPapers(t, 1)

				categories := createTestCategories(t, []models.PaperCategory{
					{
						PaperID:    papers[0].ID,
						CategoryID: 1,
					},
				})

				return &SetupReturn{
					Paper:      papers[0],
					Categories: categories,
				}
			},
			GetRequest: func(setupData *SetupReturn) *proto.ReorderPaperCategoriesRequest {
				return &proto.ReorderPaperCategoriesRequest{
					PaperHash:   setupData.Paper.Hash,
					CategoryIds: []int64{1, 999},
				}
			},
			ExpectedCode: codes.InvalidArgument,
		},
		{
			Name:     "Non-existent category IDs",
			Metadata: defaultMetadata,
			Setup: func(t *testing.T) *SetupReturn {
				papers := createDefaultTestPapers(t, 1)

				categories := createTestCategories(t, []models.PaperCategory{
					{
						PaperID:    papers[0].ID,
						CategoryID: 1,
					},
					{
						PaperID:    papers[0].ID,
						CategoryID: 2,
					},
				})

				return &SetupReturn{
					Paper:      papers[0],
					Categories: categories,
				}
			},
			GetRequest: func(setupData *SetupReturn) *proto.ReorderPaperCategoriesRequest {
				return &proto.ReorderPaperCategoriesRequest{
					PaperHash:   setupData.Paper.Hash,
					CategoryIds: []int64{1, 999},
				}
			},
			ExpectedCode: codes.NotFound,
		},
		{
			Name:     "Non-existent paper hash",
			Metadata: defaultMetadata,
			GetRequest: func(setupData *SetupReturn) *proto.ReorderPaperCategoriesRequest {
				return &proto.ReorderPaperCategoriesRequest{
					PaperHash:   "nonexistent",
					CategoryIds: []int64{1, 2, 3},
				}
			},
			ExpectedCode: codes.PermissionDenied,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			clearTables(t)
			test.Runner(t, tc, client.ReorderPaperCategories)
		})
	}
}

func TestCreateCategory(t *testing.T) {
	type SetupReturn struct {
		Paper models.Paper
	}

	testCases := []test.TestCase[*proto.PaperRequest, *proto.PaperCategoryResponse, *SetupReturn]{
		{
			Name:     "Create category for paper",
			Metadata: defaultMetadata,
			Setup: func(t *testing.T) *SetupReturn {
				papers := createDefaultTestPapers(t, 1)

				// Create an initial category to test order increment
				createTestCategories(t, []models.PaperCategory{
					{
						PaperID:    papers[0].ID,
						CategoryID: 1,
					},
				})

				return &SetupReturn{Paper: papers[0]}
			},
			GetRequest: func(setupData *SetupReturn) *proto.PaperRequest {
				return &proto.PaperRequest{
					PaperHash: setupData.Paper.Hash,
				}
			},
			ExpectedCode: codes.OK,
			Validate: func(t *testing.T, resp *proto.PaperCategoryResponse, setupData *SetupReturn) {
				// Verify the response
				assert.NotZero(t, resp.Id)
				assert.NotEmpty(t, resp.Name)
				assert.Equal(t, int32(2), resp.Order) // Should be second category

				// Verify database entry
				var category models.PaperCategory
				err := dbInst.Where("paper_id = ? AND category_id = ?", setupData.Paper.ID, resp.Id).
					Take(&category).Error
				require.NoError(t, err)
				assert.Equal(t, int16(2), category.Order)
			},
		},
		{
			Name:     "Non-existent paper hash",
			Metadata: defaultMetadata,
			GetRequest: func(setupData *SetupReturn) *proto.PaperRequest {
				return &proto.PaperRequest{
					PaperHash: "nonexistent",
				}
			},
			ExpectedCode: codes.PermissionDenied,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			clearTables(t)
			test.Runner(t, tc, client.CreatePaperCategory)
		})
	}
}

func TestUpdateCategory(t *testing.T) {
	type SetupReturn struct {
		Paper    models.Paper
		Category models.PaperCategory
	}

	testCases := []test.TestCase[*proto.UpdatePaperCategoryRequest, *emptypb.Empty, *SetupReturn]{
		{
			Name:     "Update category name",
			Metadata: defaultMetadata,
			Setup: func(t *testing.T) *SetupReturn {
				papers := createDefaultTestPapers(t, 1)

				categories := createTestCategories(t, []models.PaperCategory{
					{
						PaperID:    papers[0].ID,
						CategoryID: 1,
					},
				})

				return &SetupReturn{
					Paper:    papers[0],
					Category: categories[0],
				}
			},
			GetRequest: func(setupData *SetupReturn) *proto.UpdatePaperCategoryRequest {
				return &proto.UpdatePaperCategoryRequest{
					PaperHash:  setupData.Paper.Hash,
					CategoryId: int64(setupData.Category.CategoryID),
					Name:       "Updated Category Name",
				}
			},
			ExpectedCode: codes.OK,
		},
		{
			Name:     "Non-existent paper hash",
			Metadata: defaultMetadata,
			Setup: func(t *testing.T) *SetupReturn {
				papers := createDefaultTestPapers(t, 1)

				categories := createTestCategories(t, []models.PaperCategory{
					{
						PaperID:    papers[0].ID,
						CategoryID: 1,
					},
				})

				return &SetupReturn{
					Paper:    papers[0],
					Category: categories[0],
				}
			},
			GetRequest: func(setupData *SetupReturn) *proto.UpdatePaperCategoryRequest {
				return &proto.UpdatePaperCategoryRequest{
					PaperHash:  "nonexistent",
					CategoryId: int64(setupData.Category.CategoryID),
					Name:       "Updated Name",
				}
			},
			ExpectedCode: codes.PermissionDenied,
		},
		{
			Name:     "Non-existent category ID",
			Metadata: defaultMetadata,
			Setup: func(t *testing.T) *SetupReturn {
				papers := createDefaultTestPapers(t, 1)

				categories := createTestCategories(t, []models.PaperCategory{
					{
						PaperID:    papers[0].ID,
						CategoryID: 1,
					},
				})

				return &SetupReturn{
					Paper:    papers[0],
					Category: categories[0],
				}
			},
			GetRequest: func(setupData *SetupReturn) *proto.UpdatePaperCategoryRequest {
				return &proto.UpdatePaperCategoryRequest{
					PaperHash:  setupData.Paper.Hash,
					CategoryId: 999, // Non-existent category ID
					Name:       "Updated Name",
				}
			},
			ExpectedCode: codes.NotFound,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			clearTables(t)
			test.Runner(t, tc, client.UpdatePaperCategory)
		})
	}
}

func TestDeleteCategory(t *testing.T) {
	type SetupReturn struct {
		Paper      models.Paper
		Categories []models.PaperCategory
		Questions  []models.PaperQuestion
	}

	testCases := []test.TestCase[*proto.PaperCategoryRequest, *emptypb.Empty, *SetupReturn]{
		{
			Name:     "Delete category with questions",
			Metadata: defaultMetadata,
			Setup: func(t *testing.T) *SetupReturn {
				papers := createDefaultTestPapers(t, 1)

				categories := createTestCategories(t, []models.PaperCategory{
					{
						PaperID:    papers[0].ID,
						CategoryID: 1,
					},
					{
						PaperID:    papers[0].ID,
						CategoryID: 2,
					},
				})

				questions := createTestQuestions(t, []models.PaperQuestion{
					{
						PaperID:    papers[0].ID,
						QuestionID: 1,
						CategoryID: categories[0].CategoryID,
						MaxScore:   10,
					},
					{
						PaperID:    papers[0].ID,
						QuestionID: 2,
						CategoryID: categories[0].CategoryID,
						MaxScore:   15,
					},
				})

				return &SetupReturn{
					Paper:      papers[0],
					Categories: categories,
					Questions:  questions,
				}
			},
			GetRequest: func(setupData *SetupReturn) *proto.PaperCategoryRequest {
				return &proto.PaperCategoryRequest{
					PaperHash:  setupData.Paper.Hash,
					CategoryId: int64(setupData.Categories[0].CategoryID),
				}
			},
			ExpectedCode: codes.OK,
			Validate: func(t *testing.T, resp *emptypb.Empty, setupData *SetupReturn) {
				// Verify category is deleted
				var count int64
				err := dbInst.Model(&models.PaperCategory{}).
					Where("paper_id = ? AND category_id = ?", setupData.Paper.ID, setupData.Categories[0].CategoryID).
					Count(&count).Error
				require.NoError(t, err)
				assert.Equal(t, int64(0), count)

				// Verify questions in category are deleted
				err = dbInst.Model(&models.PaperQuestion{}).
					Where("paper_id = ? AND category_id = ?", setupData.Paper.ID, setupData.Categories[0].CategoryID).
					Count(&count).Error
				require.NoError(t, err)
				assert.Equal(t, int64(0), count)
			},
		},
		{
			Name:     "Cannot delete last category",
			Metadata: defaultMetadata,
			Setup: func(t *testing.T) *SetupReturn {
				papers := createDefaultTestPapers(t, 1)

				categories := createTestCategories(t, []models.PaperCategory{
					{
						PaperID:    papers[0].ID,
						CategoryID: 1,
					},
				})

				return &SetupReturn{
					Paper:      papers[0],
					Categories: categories,
				}
			},
			GetRequest: func(setupData *SetupReturn) *proto.PaperCategoryRequest {
				return &proto.PaperCategoryRequest{
					PaperHash:  setupData.Paper.Hash,
					CategoryId: int64(setupData.Categories[0].CategoryID),
				}
			},
			ExpectedCode: codes.FailedPrecondition,
		},
		{
			Name:     "Non-existent paper hash",
			Metadata: defaultMetadata,
			GetRequest: func(setupData *SetupReturn) *proto.PaperCategoryRequest {
				return &proto.PaperCategoryRequest{
					PaperHash:  "nonexistent",
					CategoryId: 1,
				}
			},
			ExpectedCode: codes.PermissionDenied,
		},
		{
			Name:     "Non-existent category ID",
			Metadata: defaultMetadata,
			Setup: func(t *testing.T) *SetupReturn {
				papers := createDefaultTestPapers(t, 1)

				categories := createTestCategories(t, []models.PaperCategory{
					{
						PaperID:    papers[0].ID,
						CategoryID: 1,
					},
					{
						PaperID:    papers[0].ID,
						CategoryID: 2,
					},
				})

				return &SetupReturn{
					Paper:      papers[0],
					Categories: categories,
				}
			},
			GetRequest: func(setupData *SetupReturn) *proto.PaperCategoryRequest {
				return &proto.PaperCategoryRequest{
					PaperHash:  setupData.Paper.Hash,
					CategoryId: 999, // Non-existent category
				}
			},
			ExpectedCode: codes.NotFound,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			clearTables(t)
			test.Runner(t, tc, client.DeletePaperCategory)
		})
	}
}
