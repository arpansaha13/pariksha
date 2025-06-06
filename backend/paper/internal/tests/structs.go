package tests

import (
	"testing"

	"google.golang.org/grpc/codes"

	"pariksha/common/pkg/models"
	"pariksha/common/pkg/proto"
	"pariksha/common/pkg/types"
)

// BaseTestCase contains common fields used across all test cases
type BaseTestCase struct {
	name         string
	userID       types.UserID
	expectedCode codes.Code
}

type GetPaperTestCase struct {
	BaseTestCase
	setup    func(t *testing.T) *models.Paper
	validate func(t *testing.T, resp *proto.PaperResponse)
}

type ListPapersTestCase struct {
	BaseTestCase
	setup    func(t *testing.T)
	validate func(t *testing.T, resp *proto.PaperList)
}

type DeleteCategoryCase struct {
	BaseTestCase
	setup    func(t *testing.T) *models.QuestionCategory
	validate func(t *testing.T, categoryID types.CategoryID)
}

type UpdatePaperCase struct {
	BaseTestCase
	setup    func(t *testing.T) *models.Paper
	request  *proto.UpdatePaperRequest
	validate func(t *testing.T, paper *models.Paper)
}

type QuestionListCase struct {
	BaseTestCase
	setup    func(t *testing.T) (*models.Paper, []models.Question)
	validate func(t *testing.T, resp *proto.QuestionList)
}

type GetPaperQuestionCase struct {
	BaseTestCase
	setup    func(t *testing.T) (*models.Paper, *models.Question)
	validate func(t *testing.T, resp *proto.QuestionResponse)
}

type PaperPermissionsCase struct {
	BaseTestCase
	setup    func(t *testing.T) *models.Paper
	validate func(t *testing.T, resp *proto.PaperPermissionsResponse)
}

type CreateQuestionCase struct {
	BaseTestCase
	setup    func(t *testing.T) (*models.Paper, *models.QuestionCategory)
	request  *proto.CreateQuestionRequest
	validate func(t *testing.T, paper *models.Paper, resp *proto.CreateQuestionResponse)
}

type UpdateQuestionCase struct {
	BaseTestCase
	setup    func(t *testing.T) (*models.Paper, *models.Question)
	request  *proto.UpdateQuestionRequest
	validate func(t *testing.T, paper *models.Paper, question *models.Question)
}

type DeleteQuestionCase struct {
	BaseTestCase
	setup    func(t *testing.T) (*models.Paper, *models.Question)
	validate func(t *testing.T, paper *models.Paper, questionID types.QuestionID)
}

type ReorderQuestionsCase struct {
	BaseTestCase
	setup    func(t *testing.T) (*models.Paper, *models.QuestionCategory, []models.Question)
	request  *proto.ReorderQuestionsRequest
	validate func(t *testing.T, questions []models.Question)
}

type CategoryListCase struct {
	BaseTestCase
	setup    func(t *testing.T) (*models.Paper, []models.QuestionCategory)
	validate func(t *testing.T, categories []models.QuestionCategory, resp *proto.CategoryList)
}

type CreateCategoryCase struct {
	BaseTestCase
	setup    func(t *testing.T) *models.Paper
	validate func(t *testing.T, resp *proto.CategoryResponse)
}

type UpdateCategoryCase struct {
	BaseTestCase
	setup    func(t *testing.T) *models.QuestionCategory
	newName  string
	validate func(t *testing.T, category *models.QuestionCategory)
}

type ReorderCategoriesCase struct {
	BaseTestCase
	setup    func(t *testing.T) (*models.Paper, []models.QuestionCategory)
	validate func(t *testing.T, categories []models.QuestionCategory)
}
