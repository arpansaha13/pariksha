package handlers

import (
	"context"
	"encoding/json"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"pariksha/common/pkg/constants"
	"pariksha/common/pkg/models"
	"pariksha/common/pkg/proto"
	"pariksha/common/pkg/utils"
	"pariksha/paper/internal/config/db"
)

// UpsertPaperTestCases handles bulk creation and updates of test cases
func (s *PaperServer) UpsertPaperTestCases(ctx context.Context, req *proto.UpsertTestCasesRequest) (*proto.Empty, error) {
	pc, err := NewPaperContext(ctx)
	if err != nil || pc.Question == nil {
		return nil, status.Error(codes.Internal, "question data not found in context")
	}

	if pc.Question.Type != constants.QUESTION_TYPE_CODING {
		return nil, status.Error(codes.InvalidArgument, "test cases can only be added to coding questions")
	}

	// Detect duplicate IDs in testCases
	seenIds := make(map[int64]bool)
	for _, tc := range req.TestCases {
		if tc.Id != nil {
			if seenIds[*tc.Id] {
				return nil, status.Error(codes.InvalidArgument, "duplicate test case ID found in request")
			}
			seenIds[*tc.Id] = true
		}
	}

	err = utils.TransactionHandler(db.DB, func(tx *gorm.DB) error {
		// Separate test cases into updates and creates
		var (
			toUpdate []models.TestCase
			toCreate []models.TestCase
		)

		for _, tc := range req.TestCases {
			testCase := models.TestCase{
				QuestionID: req.QuestionId,
				Hidden:     tc.Hidden,
			}

			// Create content JSON
			content := models.TestCaseContent{
				Inputs:      tc.Inputs,
				Output:      tc.Output,
				Explanation: tc.Explanation,
			}
			contentBytes, err := json.Marshal(content)
			if err != nil {
				return status.Error(codes.Internal, "failed to marshal test case content")
			}
			testCase.Content = contentBytes

			if tc.Id != nil {
				testCase.ID = *tc.Id
				toUpdate = append(toUpdate, testCase)
			} else {
				toCreate = append(toCreate, testCase)
			}
		}

		// Bulk create new test cases
		if len(toCreate) > 0 {
			if err := tx.Create(&toCreate).Error; err != nil {
				return status.Error(codes.Internal, "failed to create test cases")
			}
		}

		// Bulk update existing test cases
		if len(toUpdate) > 0 {
			if err := tx.Clauses(clause.OnConflict{
				Columns:   []clause.Column{{Name: "id"}},
				DoUpdates: clause.AssignmentColumns([]string{"content", "hidden"}),
			}).Create(&toUpdate).Error; err != nil {
				return status.Error(codes.Internal, "failed to update test cases")
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return &proto.Empty{}, nil
}
