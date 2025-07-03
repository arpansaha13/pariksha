package services

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"gorm.io/gorm"

	"pariksha/common/pkg/constants"
	"pariksha/common/pkg/models"
	"pariksha/common/pkg/proto"
	"pariksha/common/pkg/types"
	"pariksha/common/pkg/utils/grpcerror"
	"pariksha/paper/internal/interceptors"
	"pariksha/paper/internal/repositories"
	paperUtils "pariksha/paper/internal/utils"
)

type TestCase struct {
	questionRepo *repositories.Question
	testCaseRepo *repositories.TestCase
}

// NewTestCase creates a new test case service instance
func NewTestCase(questionRepo *repositories.Question, testCaseRepo *repositories.TestCase) *TestCase {
	return &TestCase{
		testCaseRepo: testCaseRepo,
		questionRepo: questionRepo,
	}
}

// UpsertTestCases handles bulk creation and updates of test cases
func (s *TestCase) UpsertTestCases(ctx context.Context, req *proto.UpsertTestCasesRequest) (*emptypb.Empty, error) {
	question, err := interceptors.GetQuestionFromContext(ctx)
	if err != nil {
		return nil, err
	}

	if question.Type != proto.QuestionType_CODING {
		return nil, status.Error(codes.InvalidArgument, "test cases can only be added to coding questions")
	}

	if len(req.TestCases) > int(constants.MAX_CODING_TEST_CASES_COUNT) {
		return nil, status.Error(codes.InvalidArgument,
			fmt.Sprintf("Number of test cases cannot be more than %d", constants.MAX_CODING_TEST_CASES_COUNT))
	}

	err = s.testCaseRepo.Transaction(func(tx *gorm.DB) error {
		inputDefinitionsLength, err := s.questionRepo.GetInputDefinitionsLength(tx, question.ID)
		if err != nil {
			return grpcerror.Internal(err, "could not get input_definitions length")
		}

		existingTestCases, err := s.testCaseRepo.GetUnscopedTestCasesByQuestionId(tx, question.ID)
		if err != nil {
			return grpcerror.Internal(err, "could not get unscoped test cases")
		}

		var toCreate, toUpdate []models.TestCase
		for idx, tc := range req.TestCases {
			content := models.TestCaseContent{
				Inputs: tc.Inputs,
				Output: tc.Output,
			}

			if tc.Explanation != nil && strings.TrimSpace(*tc.Explanation) != "" {
				content.Explanation = tc.Explanation
			}

			err := content.Validate(inputDefinitionsLength)
			if err != nil {
				return err
			}

			contentBytes, err := json.Marshal(content)
			if err != nil {
				return status.Error(codes.Internal, "failed to marshal test case content")
			}

			order := int16(idx + 1)

			// Hash the incoming content for comparison
			dataHash := paperUtils.GenerateDataHash(content, tc.Hidden)

			// Check if there's an existing test case (including soft-deleted) at this order
			var existingAtOrder *models.TestCase
			for _, existing := range existingTestCases {
				if existing.Order == order {
					existingAtOrder = &existing
					break
				}
			}

			if existingAtOrder != nil {
				// If test case exists (even if soft-deleted), update it
				toUpdate = append(toUpdate, models.TestCase{
					ID:         existingAtOrder.ID,
					QuestionID: question.ID,
					Order:      order,
					Content:    contentBytes,
					DataHash:   dataHash,
					Hidden:     tc.Hidden,
					DeletedAt:  gorm.DeletedAt{},
				})
			} else {
				toCreate = append(toCreate, models.TestCase{
					QuestionID: question.ID,
					Order:      order,
					Content:    contentBytes,
					DataHash:   dataHash,
					Hidden:     tc.Hidden,
				})
			}
		}

		// Delete extra test cases if request list is shorter
		if len(req.TestCases) < len(existingTestCases) {
			var idsToDelete []types.TestCaseID
			for _, tc := range existingTestCases[len(req.TestCases):] {
				if !tc.DeletedAt.Valid { // Only delete if not already soft-deleted
					idsToDelete = append(idsToDelete, tc.ID)
				}
			}
			if len(idsToDelete) > 0 {
				if err := s.testCaseRepo.DeleteByIds(tx, idsToDelete); err != nil {
					return status.Error(codes.Internal, "failed to delete extra test cases")
				}
			}
		}

		// Bulk create new test cases
		if len(toCreate) > 0 {
			if err := s.testCaseRepo.Create(tx, &toCreate); err != nil {
				return err
			}
		}

		// Bulk update existing test cases (includes reviving soft-deleted ones)
		for _, tc := range toUpdate {
			if err := s.testCaseRepo.UpdateUnscopedTestCase(tx, tc); err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}
