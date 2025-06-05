package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

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
	"pariksha/paper/internal/utils/validate"
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

	if len(req.TestCases) > int(constants.MAX_CODING_TEST_CASES_COUNT) {
		return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("Number of test cases cannot be more than %d", constants.MAX_CODING_TEST_CASES_COUNT))
	}

	inputDefinitionsLength, err := getInputDefinitionsLength(db.DB, req.QuestionId)
	if err != nil {
		return nil, status.Error(codes.Internal, "could not get input_definitions length")
	}

	err = utils.TransactionHandler(db.DB, func(tx *gorm.DB) error {
		// Get existing test cases including soft-deleted ones with row-level locking
		var existingTestCases []models.TestCase
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Unscoped(). // Include soft-deleted records
			Where("question_id = ?", req.QuestionId).
			Order("\"order\"").
			Find(&existingTestCases).Error; err != nil {
			return status.Error(codes.Internal, "failed to fetch existing test cases")
		}

		// Process each test case from the request
		var toCreate, toUpdate []models.TestCase
		for idx, tc := range req.TestCases {
			content := models.TestCaseContent{
				Inputs: tc.Inputs,
				Output: tc.Output,
			}

			// Include explanation if it exists
			if tc.Explanation != nil && strings.TrimSpace(*tc.Explanation) != "" {
				content.Explanation = tc.Explanation
			}

			err := validate.TestCase(&content, inputDefinitionsLength)
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
					QuestionID: req.QuestionId,
					Order:      order,
					Content:    contentBytes,
					DataHash:   dataHash,
					Hidden:     tc.Hidden,
					DeletedAt:  gorm.DeletedAt{}, // Revive if soft-deleted
				})
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
				if !tc.DeletedAt.Valid { // Only delete if not already soft-deleted
					idsToDelete = append(idsToDelete, tc.ID)
				}
			}
			if len(idsToDelete) > 0 {
				if err := tx.Delete(&models.TestCase{}, idsToDelete).Error; err != nil {
					return status.Error(codes.Internal, "failed to delete extra test cases")
				}
			}
		}

		// Bulk create new test cases
		if len(toCreate) > 0 {
			if err := tx.Create(&toCreate).Error; err != nil {
				return status.Error(codes.Internal, "failed to create test cases")
			}
		}

		// Bulk update existing test cases (includes reviving soft-deleted ones)
		if len(toUpdate) > 0 {
			for _, tc := range toUpdate {
				if err := tx.Unscoped().Model(&models.TestCase{}).
					Where("id = ?", tc.ID).
					Updates(map[string]any{
						"content":    tc.Content,
						"data_hash":  tc.DataHash,
						"hidden":     tc.Hidden,
						"deleted_at": nil, // Revive soft-deleted record
					}).Error; err != nil {
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

// getInputDefinitionsLength retrieves the length of the InputDefinitions array
// from the JSONB Question field for a given question ID.
func getInputDefinitionsLength(db *gorm.DB, questionID int64) (int, error) {
	var length int
	err := db.Raw(`
        SELECT jsonb_array_length(
            (Question->>'input_definitions')::jsonb
        ) FROM questions WHERE id = ?
    `, questionID).Scan(&length).Error
	return length, err
}
