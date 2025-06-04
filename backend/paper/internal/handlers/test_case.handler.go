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
	paperUtils "pariksha/paper/internal/utils"
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

	err = utils.TransactionHandler(db.DB, func(tx *gorm.DB) error {
		// Get existing test cases with row-level locking
		// The clause.Locking{Strength: "UPDATE"} translates to SELECT ... FOR UPDATE in SQL
		var existingTestCases []models.TestCase
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("question_id = ?", req.QuestionId).
			Order("\"order\"").
			Find(&existingTestCases).Error; err != nil {
			return status.Error(codes.Internal, "failed to fetch existing test cases")
		}

		// Process each test case from the request
		var toCreate, toUpdate []models.TestCase
		for idx, tc := range req.TestCases {
			content := models.TestCaseContent{
				Inputs:      tc.Inputs,
				Output:      tc.Output,
				Explanation: tc.Explanation,
			}
			contentBytes, err := json.Marshal(content)
			if err != nil {
				return status.Error(codes.Internal, "failed to marshal test case content")
			}

			order := int16(idx + 1)

			// Hash the incoming content for comparison
			dataHash := paperUtils.GenerateDataHash(content, tc.Hidden)

			// Check if there's an existing test case at this order
			if idx < len(existingTestCases) {
				existing := existingTestCases[idx]
				if existing.DataHash != dataHash {
					toUpdate = append(toUpdate, models.TestCase{
						ID:         existing.ID,
						QuestionID: req.QuestionId,
						Order:      order,
						Content:    contentBytes,
						DataHash:   dataHash,
						Hidden:     tc.Hidden,
					})
				}
			} else {
				toCreate = append(toCreate, models.TestCase{
					QuestionID: req.QuestionId,
					Order:      order,
					Content:    contentBytes,
					DataHash:   dataHash,
					Hidden:     tc.Hidden,
				})
			}
		}

		// Delete extra test cases if request list is shorter
		if len(req.TestCases) < len(existingTestCases) {
			idsToDelete := make([]int64, 0)
			for _, tc := range existingTestCases[len(req.TestCases):] {
				idsToDelete = append(idsToDelete, tc.ID)
			}
			if err := tx.Delete(&models.TestCase{}, idsToDelete).Error; err != nil {
				return status.Error(codes.Internal, "failed to delete extra test cases")
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
			for _, tc := range toUpdate {
				if err := tx.Model(&models.TestCase{}).Where("id = ?", tc.ID).Updates(tc).Error; err != nil {
					return status.Error(codes.Internal, "failed to update test cases")
				}
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return &proto.Empty{}, nil
}
