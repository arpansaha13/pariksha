package services

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"

	"pariksha/common/pkg/constants"
	"pariksha/common/pkg/models"
	"pariksha/common/pkg/proto"
	"pariksha/common/pkg/structs"
	"pariksha/common/pkg/types"
	"pariksha/common/pkg/utils"
	"pariksha/common/pkg/utils/generate"
	"pariksha/common/pkg/utils/grpcerror"
	"pariksha/common/pkg/utils/ptr"
	"pariksha/paper/internal/config/db"
	"pariksha/paper/internal/interceptors"
	"pariksha/paper/internal/repositories"
	"pariksha/paper/internal/utils/validate"
)

type Question struct {
	paperRepo    *repositories.Paper
	questionRepo *repositories.Question
}

func NewQuestion(paperRepo *repositories.Paper, questionRepo *repositories.Question) *Question {
	return &Question{
		paperRepo:    paperRepo,
		questionRepo: questionRepo,
	}
}

// GetPaperQuestions handles fetching all questions for a paper
func (s *Question) GetPaperQuestions(ctx context.Context, paperHash string) (*proto.QuestionList, error) {
	questions, err := s.questionRepo.GetPaperQuestions(nil, paperHash)
	if err != nil {
		return nil, grpcerror.Internal(err, "failed to fetch paper questions")
	}

	response := &proto.QuestionList{
		Questions: make([]*proto.QuestionMinimal, len(questions)),
	}

	for i, question := range questions {
		question.Paper.Hash = paperHash
		response.Questions[i], err = questionToMinimalProto(question)
		if err != nil {
			return nil, err
		}
	}

	return response, nil
}

// GetPaperQuestion handles fetching a single question with its test cases
func (s *Question) GetPaperQuestion(ctx context.Context) (*proto.QuestionResponse, error) {
	question, err := interceptors.GetQuestionFromContext(ctx)
	if err != nil {
		return nil, err
	}

	var testCases []models.TestCase
	if question.Type == proto.QuestionType_CODING {
		testCases, err = s.questionRepo.GetTestCasesForQuestion(nil, question.ID)
		if err != nil {
			return nil, status.Error(codes.Internal, constants.ErrInternalServer)
		}
	}

	return questionToProto(*question, testCases)
}

// CreateQuestion handles the business logic for creating a new question.
func (s *Question) CreateQuestion(ctx context.Context, req *proto.CreateQuestionRequest) (*proto.CreateQuestionResponse, error) {
	if err := validate.MaxScore(req.MaxScore); err != nil {
		return nil, err
	}

	// Validate question data based on type
	var coding structs.CodingQuestion
	switch req.Type {
	case proto.QuestionType_MCQ:
		var mcq structs.MCQQuestion
		if err := utils.StrictUnmarshal(req.RawQuestion, &mcq); err != nil {
			return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("invalid MCQ question format: %s", err.Error()))
		}
		if err := mcq.Validate(); err != nil {
			return nil, err
		}
	case proto.QuestionType_SUBJECTIVE:
		var subjective structs.SubjectiveQuestion
		if err := utils.StrictUnmarshal(req.RawQuestion, &subjective); err != nil {
			return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("invalid subjective question format: %s", err.Error()))
		}
		if err := subjective.Validate(); err != nil {
			return nil, err
		}
	case proto.QuestionType_CODING:
		if err := utils.StrictUnmarshal(req.RawQuestion, &coding); err != nil {
			return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("invalid coding question format: %s", err.Error()))
		}
		if err := coding.Validate(); err != nil {
			return nil, err
		}
	default:
		return nil, status.Error(codes.InvalidArgument, "invalid question type")
	}

	tags, _ := json.Marshal(req.Tags)
	var question models.Question

	err := utils.TransactionHandler(db.DB, func(tx *gorm.DB) error {
		// Get paper by hash
		paper, err := s.paperRepo.GetByHash(tx, req.PaperHash)
		if err != nil {
			return utils.HandleDBError(err, "paper not found")
		}

		// Get max order for this category
		maxOrder, err := s.questionRepo.GetMaxQuestionOrder(tx, req.CategoryId)
		if err != nil {
			return grpcerror.Internal(err, "failed to get max order for category")
		}

		question = models.Question{
			PaperID:    sql.NullInt64{Int64: int64(paper.ID), Valid: true},
			CategoryID: types.CategoryID(req.CategoryId),
			Order:      maxOrder + 1,
			Question:   json.RawMessage(req.RawQuestion),
			Type:       req.Type,
			Tags:       ptr.JsonRawMessage(tags),
			MaxScore:   int16(req.MaxScore),
			CorrectAnswer: sql.NullString{
				String: req.GetCorrectAnswer(),
				Valid:  req.CorrectAnswer != nil,
			},
		}

		if err := s.questionRepo.CreateQuestion(tx, &question); err != nil {
			return err
		}

		// Generate and store hash
		question.Hash = generate.HMACHash(int64(question.ID))
		if err := s.questionRepo.UpdateQuestionHash(tx, question.ID, question.Hash); err != nil {
			return status.Error(codes.Internal, "failed to store question hash")
		}

		// Create boilerplates for coding questions
		if req.Type == proto.QuestionType_CODING {
			if err := upsertBoilerplates(tx, question.ID, coding.InputDefinitions, coding.OutputDefinition); err != nil {
				return status.Error(codes.Internal, "failed to create boilerplates")
			}
		}

		newCounts, err := updateQuestionCounts(paper.QuestionCounts, req.Type, 1)
		if err != nil {
			return err
		}

		return updatePaperStats(tx, *paper, int32(req.MaxScore), newCounts)
	})

	if err != nil {
		return nil, err
	}

	return &proto.CreateQuestionResponse{
		QuestionHash: question.Hash,
	}, nil
}

// DeleteQuestion handles the business logic for deleting a question
func (s *Question) DeleteQuestion(ctx context.Context, questionHash string) error {
	return utils.TransactionHandler(db.DB, func(tx *gorm.DB) error {
		// Get question by hash
		question, err := s.questionRepo.GetQuestionByHash(tx, questionHash)
		if err != nil {
			return utils.HandleDBError(err, "question not found")
		}

		// Get paper details
		paper, err := s.paperRepo.GetDetails(tx, types.PaperID(question.PaperID.Int64))
		if err != nil {
			return utils.HandleDBError(err, "failed to fetch paper details")
		}

		// Update paper max score
		if err := s.paperRepo.UpdateMaxScore(tx, paper.ID, question.MaxScore); err != nil {
			return err
		}

		// Update question counts
		newCounts, err := updateQuestionCounts(paper.QuestionCounts, question.Type, -1)
		if err != nil {
			return err
		}

		if err := s.paperRepo.UpdateQuestionCounts(tx, paper.ID, newCounts); err != nil {
			return err
		}

		// Handle question deletion or unlinking
		if question.Locked {
			if err := s.questionRepo.UnlinkFromPaper(tx, question.ID); err != nil {
				return err
			}
		} else {
			if err := s.questionRepo.Delete(tx, question.ID); err != nil {
				return err
			}
		}

		return nil
	})
}

// ReorderQuestions handles the business logic for reordering questions
func (s *Question) ReorderQuestions(ctx context.Context, categoryID int64, questionHashes []string) error {
	return utils.TransactionHandler(db.DB, func(tx *gorm.DB) error {
		// Verify all questions belong to the category
		count, err := s.questionRepo.ValidateCategoryQuestions(tx, categoryID, questionHashes)
		if err != nil {
			return err
		}

		if int(count) != len(questionHashes) {
			return status.Error(codes.InvalidArgument, "invalid question hashes")
		}

		// Update orders
		for i, questionHash := range questionHashes {
			if err := s.questionRepo.UpdateOrder(tx, questionHash, int16(i+1)); err != nil {
				return err
			}
		}

		return nil
	})
}

// GetQuestionsByIds handles fetching multiple questions by their IDs
func (s *Question) GetQuestionsByIds(ctx context.Context, questionIDs []int64) (*proto.GetQuestionsByIdsResponse, error) {
	questions, err := s.questionRepo.GetByIds(nil, questionIDs)
	if err != nil {
		return nil, grpcerror.Internal(err, "failed to fetch questions")
	}

	response := &proto.GetQuestionsByIdsResponse{
		Questions: make([]*proto.QuestionBatchItem, len(questions)),
	}

	for i, question := range questions {
		response.Questions[i] = &proto.QuestionBatchItem{
			QuestionId:   int64(question.ID),
			QuestionHash: question.Hash,
			MaxScore:     int32(question.MaxScore),
			Type:         question.Type,
			RawQuestion:  question.Question,
		}
	}

	return response, nil
}

// GetExamQuestion handles fetching question data for exam taking
func (s *Question) GetExamQuestion(ctx context.Context, hash string) (*proto.QuestionResponse, error) {
	question, err := s.questionRepo.GetExamQuestion(nil, hash)
	if err != nil {
		return nil, status.Error(codes.NotFound, constants.ErrNotFound)
	}

	var testCases []models.TestCase
	if question.Type == proto.QuestionType_CODING {
		testCases, err = s.questionRepo.GetNonHiddenTestCases(nil, question.ID)
		if err != nil {
			return nil, status.Error(codes.Internal, constants.ErrInternalServer)
		}
	}

	return questionToProto(*question, testCases)
}

// GetQuestionHashes handles the business logic for fetching question hashes
func (s *Question) GetQuestionHashes(ctx context.Context, questionIDs []int64) (*proto.GetQuestionHashesResponse, error) {
	if len(questionIDs) == 0 {
		return &proto.GetQuestionHashesResponse{}, nil
	}

	questions, err := s.questionRepo.GetQuestionHashes(nil, questionIDs)
	if err != nil {
		return nil, status.Error(codes.Internal, constants.ErrInternalServer)
	}

	// Create a map for quick hash lookups
	hashMap := make(map[int64]string, len(questions))
	for _, q := range questions {
		hashMap[int64(q.ID)] = q.Hash
	}

	// Create response maintaining same sequence as request
	hashes := make([]string, len(questionIDs))
	for i, id := range questionIDs {
		if hash, ok := hashMap[id]; ok {
			hashes[i] = hash
		}
	}

	return &proto.GetQuestionHashesResponse{
		QuestionHashes: hashes,
	}, nil
}

// GetQuestionIds handles the business logic for getting question IDs from hashes
func (s *Question) GetQuestionIds(ctx context.Context, questionHashes []string) (*proto.GetQuestionIdsResponse, error) {
	if len(questionHashes) == 0 {
		return &proto.GetQuestionIdsResponse{}, nil
	}

	questions, err := s.questionRepo.GetQuestionsByHashes(nil, questionHashes)
	if err != nil {
		return nil, status.Error(codes.Internal, constants.ErrInternalServer)
	}

	// Create a map for quick ID lookups
	idMap := make(map[string]int64, len(questions))
	for _, q := range questions {
		idMap[q.Hash] = int64(q.ID)
	}

	// Create response maintaining same sequence as request
	ids := make([]int64, len(questionHashes))
	for i, hash := range questionHashes {
		if id, ok := idMap[hash]; ok {
			ids[i] = id
		}
	}

	return &proto.GetQuestionIdsResponse{
		QuestionIds: ids,
	}, nil
}

// GetCodingQuestionInputDefinitions fetches input definitions for a coding question by hash.
func (s *Question) GetCodingQuestionInputDefinitions(ctx context.Context, questionHash string) (*proto.GetCodingQuestionInputDefinitionsResponse, error) {
	inputDefs, err := s.questionRepo.GetInputDefinitionsByHash(nil, questionHash)
	if err != nil {
		return nil, grpcerror.Internal(err, "failed to fetch input definitions")
	}
	resp := &proto.GetCodingQuestionInputDefinitionsResponse{
		InputDefinitions: make([]*proto.InputDefinition, len(inputDefs)),
	}
	for i, def := range inputDefs {
		resp.InputDefinitions[i] = &proto.InputDefinition{
			VariableName: def.VariableName,
			Type:         def.Type,
			Items:        def.Items,
		}
	}
	return resp, nil
}
