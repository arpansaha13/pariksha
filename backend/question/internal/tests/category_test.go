package tests

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"

	"pariksha/common/pkg/proto"
	"pariksha/common/pkg/test"
	"pariksha/question/internal/config/db"
	"pariksha/question/internal/models"
)

const (
	TEST_CATEGORY_NAME     string = "Data Structures"
	TEST_NEW_CATEGORY_NAME string = "New Name"
)

func TestCreateCategory(t *testing.T) {
	testCases := []test.TestCase[*proto.CreateCategoryRequest, *proto.CategoryResponse, map[string]any]{
		{
			Name: "Valid category name",
			GetRequest: func(setupData map[string]any) *proto.CreateCategoryRequest {
				return &proto.CreateCategoryRequest{
					Name: TEST_CATEGORY_NAME,
				}
			},
			ExpectedCode: codes.OK,
			Validate: func(t *testing.T, resp *proto.CategoryResponse, setupData map[string]any) {
				assert.Greater(t, resp.Id, int64(0))
				assert.Equal(t, TEST_CATEGORY_NAME, resp.Name)

				// Verify in database
				var category models.Category
				err := db.DB.Take(&category, resp.Id).Error
				require.NoError(t, err)
				assert.Equal(t, TEST_CATEGORY_NAME, category.Name)
				assert.Equal(t, int32(0), category.PaperIndegree)
				assert.Equal(t, int32(0), category.ExamIndegree)
			},
		},
		{
			Name: "Empty category name",
			GetRequest: func(setupData map[string]any) *proto.CreateCategoryRequest {
				return &proto.CreateCategoryRequest{
					Name: "",
				}
			},
			ExpectedCode: codes.InvalidArgument,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			clearTables(t)
			test.Runner(t, tc, client.CreateCategory)
		})
	}
}

func TestUpdateCategoryName(t *testing.T) {
	type SetupReturn struct {
		Category models.Category
	}

	testCases := []test.TestCase[*proto.UpdateCategoryRequest, *proto.UpdateCategoryResponse, *SetupReturn]{
		{
			Name: "Update existing category",
			Setup: func(t *testing.T) *SetupReturn {
				category := models.Category{
					Name:          "Old Name",
					PaperIndegree: 1,
					ExamIndegree:  0,
				}
				err := db.DB.Create(&category).Error
				require.NoError(t, err)
				return &SetupReturn{Category: category}
			},
			GetRequest: func(setupData *SetupReturn) *proto.UpdateCategoryRequest {
				return &proto.UpdateCategoryRequest{
					Id:   int64(setupData.Category.ID),
					Name: TEST_NEW_CATEGORY_NAME,
				}
			},
			ExpectedCode: codes.OK,
			Validate: func(t *testing.T, resp *proto.UpdateCategoryResponse, setupData *SetupReturn) {
				assert.Equal(t, int64(setupData.Category.ID), resp.Id)

				// Verify in database
				var updatedCategory models.Category
				err := db.DB.Take(&updatedCategory, setupData.Category.ID).Error
				require.NoError(t, err)
				assert.Equal(t, TEST_NEW_CATEGORY_NAME, updatedCategory.Name)
				// Other fields should remain unchanged
				assert.Equal(t, setupData.Category.PaperIndegree, updatedCategory.PaperIndegree)
				assert.Equal(t, setupData.Category.ExamIndegree, updatedCategory.ExamIndegree)
			},
		},
		{
			Name: "Update non-existent category",
			GetRequest: func(setupData *SetupReturn) *proto.UpdateCategoryRequest {
				return &proto.UpdateCategoryRequest{
					Id:   99999,
					Name: TEST_NEW_CATEGORY_NAME,
				}
			},
			ExpectedCode: codes.NotFound,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			clearTables(t)
			test.Runner(t, tc, client.UpdateCategoryName)
		})
	}
}

func TestGetCategoriesByIds(t *testing.T) {
	type SetupReturn struct {
		Categories []models.Category
	}

	testCases := []test.TestCase[*proto.CategoryIdsRequest, *proto.GetCategoriesResponse, *SetupReturn]{
		{
			Name: "Get multiple categories",
			Setup: func(t *testing.T) *SetupReturn {
				categories := []models.Category{
					{
						Name:          "Category 1",
						PaperIndegree: 1,
						ExamIndegree:  0,
					},
					{
						Name:          "Category 2",
						PaperIndegree: 2,
						ExamIndegree:  1,
					},
				}
				err := db.DB.Create(&categories).Error
				require.NoError(t, err)
				return &SetupReturn{Categories: categories}
			},
			GetRequest: func(setupData *SetupReturn) *proto.CategoryIdsRequest {
				ids := make([]int64, len(setupData.Categories))
				for i, cat := range setupData.Categories {
					ids[i] = int64(cat.ID)
				}
				return &proto.CategoryIdsRequest{
					Ids: ids,
				}
			},
			ExpectedCode: codes.OK,
			Validate: func(t *testing.T, resp *proto.GetCategoriesResponse, setupData *SetupReturn) {
				require.Equal(t, len(setupData.Categories), len(resp.Categories))
				for i, cat := range resp.Categories {
					assert.Equal(t, int64(setupData.Categories[i].ID), cat.Id)
					assert.Equal(t, setupData.Categories[i].Name, cat.Name)
				}
			},
		},
		{
			Name: "Empty IDs list",
			GetRequest: func(setupData *SetupReturn) *proto.CategoryIdsRequest {
				return &proto.CategoryIdsRequest{
					Ids: []int64{},
				}
			},
			ExpectedCode: codes.OK,
			Validate: func(t *testing.T, resp *proto.GetCategoriesResponse, setupData *SetupReturn) {
				assert.Empty(t, resp.Categories)
			},
		},
		{
			Name: "Non-existent IDs",
			GetRequest: func(setupData *SetupReturn) *proto.CategoryIdsRequest {
				return &proto.CategoryIdsRequest{
					Ids: []int64{99999, 99998},
				}
			},
			ExpectedCode: codes.OK,
			Validate: func(t *testing.T, resp *proto.GetCategoriesResponse, setupData *SetupReturn) {
				assert.Empty(t, resp.Categories)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			clearTables(t)
			test.Runner(t, tc, client.GetCategoriesByIds)
		})
	}
}
