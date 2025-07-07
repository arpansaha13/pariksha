package services

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"

	"pariksha/common/pkg/proto"
	"pariksha/common/pkg/types"
	"pariksha/question/internal/models"
	"pariksha/question/internal/structs"
	"pariksha/question/internal/utils/boilerplate"
)

func OrderByRequestSequence[T comparable, V any](requestedKeys []T, items []V, keySelector func(V) T) []V {
	// Build the map from key to item
	itemMap := make(map[T]V)
	for _, item := range items {
		itemMap[keySelector(item)] = item
	}

	// Preserve request order
	ordered := make([]V, len(requestedKeys))
	for i, key := range requestedKeys {
		ordered[i] = itemMap[key]
	}
	return ordered
}

// upsertBoilerplates creates or updates boilerplate code for all languages
func upsertBoilerplates(tx *gorm.DB, questionID types.QuestionID, inputs []structs.InputDefinition, output structs.OutputDefinition) error {
	var languages []models.Language
	if err := tx.Where("is_enabled = ?", true).Find(&languages).Error; err != nil {
		return err
	}

	for _, lang := range languages {
		code := boilerplate.Generate(&lang, inputs, output)
		if code == "" {
			continue // Skip if language is not supported
		}

		// Upsert using ON CONFLICT
		if err := tx.Exec(`
					INSERT INTO boilerplates (question_id, language_id, code)
					VALUES (?, ?, ?)
					ON CONFLICT (question_id, language_id) DO UPDATE
					SET code = EXCLUDED.code
			`, questionID, lang.ID, code).Error; err != nil {
			return err
		}
	}
	return nil
}

func testCasesToProto(testCases []models.TestCase) ([]*proto.CodingQuestionTestCase, error) {
	protoTestCases := make([]*proto.CodingQuestionTestCase, 0, len(testCases))
	for _, tc := range testCases {
		protoTestCase, err := testCaseToProto(&tc)
		if err != nil {
			return nil, err
		}
		protoTestCases = append(protoTestCases, protoTestCase)
	}
	return protoTestCases, nil
}

func testCaseToProto(tc *models.TestCase) (*proto.CodingQuestionTestCase, error) {
	var content models.TestCaseContent
	if err := json.Unmarshal(tc.Content, &content); err != nil {
		return nil, status.Error(codes.Internal, "invalid test case format")
	}

	return &proto.CodingQuestionTestCase{
		Inputs:      content.Inputs,
		Output:      content.Output,
		Explanation: content.Explanation,
		Hidden:      tc.Hidden,
		Order:       int32(tc.Order),
	}, nil
}

func getTypedQuestionIDs(ids []int64) []types.QuestionID {
	typedQuestionIDs := make([]types.QuestionID, len(ids))
	for i := range ids {
		typedQuestionIDs[i] = types.QuestionID(ids[i])
	}
	return typedQuestionIDs
}

func createQuestionsMetaResponse(meta []models.Question) (*proto.QuestionsMetaResponse, error) {
	response := proto.QuestionsMetaResponse{
		Meta: make([]*proto.QuestionMeta, len(meta)),
	}
	for i, m := range meta {
		response.Meta[i] = &proto.QuestionMeta{
			Id:          int64(m.ID),
			Hash:        m.Hash,
			Type:        m.Type,
			RawQuestion: m.Question,
		}
	}
	return &response, nil
}

// generateDataHash creates a SHA256 hash of test case content
func generateDataHash(content models.TestCaseContent, hidden bool) string {
	data, _ := json.Marshal(struct {
		Content models.TestCaseContent `json:"content"`
		Hidden  bool                   `json:"hidden"`
	}{
		Content: content,
		Hidden:  hidden,
	})
	return fmt.Sprintf("%x", sha256.Sum256(data))
}
