package services

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"gorm.io/gorm"

	"pariksha/common/pkg/constants"
	"pariksha/common/pkg/proto"
	"pariksha/common/pkg/types"
	"pariksha/common/pkg/utils"
	"pariksha/common/pkg/utils/generate"
	"pariksha/common/pkg/utils/grpcerror"
	"pariksha/question/internal/models"
	"pariksha/question/internal/repositories"
	"pariksha/question/internal/structs"
)

type Question struct {
	questionRepo    *repositories.Question
	boilerplateRepo *repositories.Boilerplate
	testcaseRepo    *repositories.TestCase
}

func NewQuestion(
	questionRepo *repositories.Question,
	boilerplateRepo *repositories.Boilerplate,
	testcaseRepo *repositories.TestCase,
) *Question {
	return &Question{
		questionRepo:    questionRepo,
		boilerplateRepo: boilerplateRepo,
		testcaseRepo:    testcaseRepo,
	}
}

// CreateQuestion handles creation of a new question
func (s *Question) CreateQuestion(req *proto.CreateQuestionRequest) (*proto.CreateQuestionResponse, error) {
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

	var question *models.Question
	s.questionRepo.Transaction(func(tx *gorm.DB) error {
		var err error
		question, err = s.questionRepo.Create(&models.Question{
			Question:      req.RawQuestion,
			Type:          req.Type,
			PaperIndegree: 1,
			ExamIndegree:  0,
		}, nil)
		if err != nil {
			return grpcerror.Internal(err, "failed to create question")
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

		return nil
	})

	return &proto.CreateQuestionResponse{
		Id:   int64(question.ID),
		Hash: question.Hash,
	}, nil
}

// UpdateQuestion modifies an existing question or creates a copy if it's being used in exams
func (s *Question) UpdateQuestion(req *proto.UpdateQuestionRequest) (*proto.UpdateQuestionResponse, error) {
	var response proto.UpdateQuestionResponse

	err := s.questionRepo.Transaction(func(tx *gorm.DB) error {
		// Get original question
		originalQuestion, err := s.questionRepo.GetQuestionByHash(tx, req.Hash)
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				return status.Error(codes.NotFound, "question not found")
			}
			return grpcerror.Internal(err, "failed to fetch question")
		}

		// Validate if raw_question is provided if question type is changing
		if req.Type != nil && *req.Type != originalQuestion.Type {
			if req.RawQuestion == nil {
				return status.Error(codes.InvalidArgument, "raw_question must be provided when changing question type")
			}
		}

		// For the switch case
		if req.Type == nil {
			req.Type = &originalQuestion.Type
		}

		if req.RawQuestion != nil {
			// Validate question data based on new type
			switch *req.Type {
			case proto.QuestionType_MCQ:
				var mcq structs.MCQQuestion
				if err := utils.StrictUnmarshal(req.RawQuestion, &mcq); err != nil {
					return status.Error(codes.InvalidArgument, fmt.Sprintf("invalid MCQ question format: %s", err.Error()))
				}
				if err := mcq.Validate(); err != nil {
					return err
				}
			case proto.QuestionType_SUBJECTIVE:
				var subjective structs.SubjectiveQuestion
				if err := utils.StrictUnmarshal(req.RawQuestion, &subjective); err != nil {
					return status.Error(codes.InvalidArgument, fmt.Sprintf("invalid subjective question format: %s", err.Error()))
				}
				if err := subjective.Validate(); err != nil {
					return err
				}
			case proto.QuestionType_CODING:
				var coding structs.CodingQuestion
				if err := utils.StrictUnmarshal(req.RawQuestion, &coding); err != nil {
					return status.Error(codes.InvalidArgument, fmt.Sprintf("invalid coding question format: %s", err.Error()))
				}
				if err := coding.Validate(); err != nil {
					return err
				}
			default:
				return status.Error(codes.InvalidArgument, "invalid question type")
			}
		}

		if originalQuestion.ExamIndegree > 0 {
			// Create a copy with updates
			newQuestion := &models.Question{
				Question:      originalQuestion.Question,
				Type:          originalQuestion.Type,
				PaperIndegree: 1,
				ExamIndegree:  0,
			}

			// Apply updates to the copy
			if req.RawQuestion != nil {
				newQuestion.Question = req.RawQuestion
			}
			if req.Type != nil {
				newQuestion.Type = *req.Type
			}

			// Create new question
			newQuestion, err = s.questionRepo.Create(newQuestion, tx)
			if err != nil {
				return grpcerror.Internal(err, "failed to create question copy")
			}

			// Generate and store hash for new question
			newQuestion.Hash = generate.HMACHash(int64(newQuestion.ID))
			if err := s.questionRepo.UpdateQuestionHash(tx, newQuestion.ID, newQuestion.Hash); err != nil {
				return status.Error(codes.Internal, "failed to store question hash")
			}

			response.Id = int64(newQuestion.ID)
			response.Hash = newQuestion.Hash
		} else {
			// Update existing question
			updates := make(map[string]any)
			if req.RawQuestion != nil {
				updates["question"] = req.RawQuestion
			}
			if req.Type != nil {
				updates["type"] = *req.Type
			}

			if err := s.questionRepo.UpdateQuestion(tx, originalQuestion.ID, updates); err != nil {
				return grpcerror.Internal(err, "failed to update question")
			}

			response.Id = int64(originalQuestion.ID)
			response.Hash = originalQuestion.Hash
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return &response, nil
}

func (s *Question) GetQuestionsByIds(req *proto.QuestionIdsRequest) (*proto.GetQuestionsResponse, error) {
	typedQuestionIDs := getTypedQuestionIDs(req.Ids)

	questions, err := s.questionRepo.GetQuestionsByIDs(nil, typedQuestionIDs)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to get questions by IDs: "+err.Error())
	}

	if len(typedQuestionIDs) != len(questions) {
		return nil, status.Error(codes.NotFound, "question not found: invalid or non-existent question id")
	}

	orderedQuestions := OrderByRequestSequence(typedQuestionIDs, questions, func(q models.Question) types.QuestionID {
		return q.ID
	})

	response, err := s.createQuestionsResponse(orderedQuestions)
	if err != nil {
		return nil, err
	}

	return response, nil
}

func (s *Question) GetQuestionsByHashes(hashes []string) (*proto.GetQuestionsResponse, error) {
	questions, err := s.questionRepo.GetQuestionsByHashes(nil, hashes)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to get questions by hashes: "+err.Error())
	}

	if len(hashes) != len(questions) {
		return nil, status.Error(codes.NotFound, "question not found: invalid or non-existent question hash")
	}

	orderedQuestions := OrderByRequestSequence(hashes, questions, func(q models.Question) string {
		return q.Hash
	})

	response, err := s.createQuestionsResponse(orderedQuestions)
	if err != nil {
		return nil, err
	}

	return response, nil
}

func (s *Question) createQuestionsResponse(questions []models.Question) (*proto.GetQuestionsResponse, error) {
	response := proto.GetQuestionsResponse{
		Questions: make([]*proto.QuestionResponse, len(questions)),
	}
	for i, q := range questions {
		var protoTestCases []*proto.CodingQuestionTestCase
		if q.Type == proto.QuestionType_CODING {
			testCases, err := s.testcaseRepo.GetAllByQuestionID(nil, q.ID)
			if err != nil {
				return nil, err
			}

			protoTestCases, err = testCasesToProto(testCases)
			if err != nil {
				return nil, err
			}
		}

		response.Questions[i] = &proto.QuestionResponse{
			Id:          int64(q.ID),
			Hash:        q.Hash,
			RawQuestion: q.Question,
			Type:        q.Type,
			TestCases:   protoTestCases,
		}
	}

	return &response, nil
}

// GetQuestionsMetaByIds retrieves metadata for questions by IDs
func (s *Question) GetQuestionsMetaByIds(req *proto.QuestionIdsRequest) (*proto.QuestionsMetaResponse, error) {
	typedQuestionIDs := getTypedQuestionIDs(req.Ids)

	meta, err := s.questionRepo.GetQuestionsMetaByIDs(nil, typedQuestionIDs)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to get question metadata by IDs: "+err.Error())
	}

	orderedMeta := OrderByRequestSequence(typedQuestionIDs, meta, func(q models.Question) types.QuestionID {
		return q.ID
	})

	response, err := createQuestionsMetaResponse(orderedMeta)
	if err != nil {
		return nil, err
	}
	return response, nil
}

// GetQuestionHashesByIds retrieves hashes of questions by IDs
func (s *Question) GetQuestionHashesByIds(req *proto.QuestionIdsRequest) (*proto.GetQuestionHashesByIdsResponse, error) {
	typedQuestionIDs := getTypedQuestionIDs(req.Ids)

	questions, err := s.questionRepo.GetQuestionHashesByIDs(nil, typedQuestionIDs)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to get question hashes by IDs: "+err.Error())
	}

	orderedQuestions := OrderByRequestSequence(typedQuestionIDs, questions, func(q models.Question) types.QuestionID {
		return q.ID
	})

	hashes := make([]string, len(orderedQuestions))
	for i, q := range orderedQuestions {
		hashes[i] = q.Hash
	}

	return &proto.GetQuestionHashesByIdsResponse{
		Hashes: hashes,
	}, nil
}

// GetQuestionsMetaByHashes retrieves metadata for questions by hashes
func (s *Question) GetQuestionsMetaByHashes(hashes []string) (*proto.QuestionsMetaResponse, error) {
	meta, err := s.questionRepo.GetQuestionMetaByHashes(nil, hashes)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to get metadata by hashes: "+err.Error())
	}

	if len(meta) != len(hashes) {
		return nil, status.Error(codes.NotFound, "could not find questions with given hashes")
	}

	orderedMeta := OrderByRequestSequence(hashes, meta, func(q models.Question) string {
		return q.Hash
	})

	response, err := createQuestionsMetaResponse(orderedMeta)
	if err != nil {
		return nil, err
	}
	return response, nil
}

// GetQuestionIdsByHashes resolves hashes to internal question IDs
func (s *Question) GetQuestionIdsByHashes(hashes []string) (*proto.GetQuestionIdsByHashesResponse, error) {
	questions, err := s.questionRepo.GetQuestionIDsByHashes(nil, hashes)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to map hashes to IDs: "+err.Error())
	}

	orderedQuestions := OrderByRequestSequence(hashes, questions, func(q models.Question) string {
		return q.Hash
	})

	questionIDs := make([]int64, len(orderedQuestions))
	for i, q := range orderedQuestions {
		questionIDs[i] = int64(q.ID)
	}

	return &proto.GetQuestionIdsByHashesResponse{Ids: questionIDs}, nil
}

// GetBoilerplate retrieves boilerplate code for a question
func (s *Question) GetBoilerplate(req *proto.GetBoilerplateRequest) (*proto.BoilerplateResponse, error) {
	boilerplate, err := s.boilerplateRepo.GetByQuestionAndLanguage(req.QuestionHash, req.LanguageId)
	if err != nil {
		return nil, status.Error(codes.NotFound, "boilerplate not found: "+err.Error())
	}

	return &proto.BoilerplateResponse{
		Code: boilerplate.Code,
	}, nil
}

// IncPaperIndegreeByIds increments paper_indegree counts for given question IDs
func (s *Question) IncPaperIndegreeByIds(ids []int64) (*emptypb.Empty, error) {
	if err := s.questionRepo.UpdatePaperIndegree(ids, true, nil); err != nil {
		return nil, status.Error(codes.Internal, "failed to increment paper indegree: "+err.Error())
	}
	return &emptypb.Empty{}, nil
}

// DecPaperIndegreeByIds decrements paper_indegree counts for given question IDs
func (s *Question) DecPaperIndegreeByIds(ids []int64) (*emptypb.Empty, error) {
	if err := s.questionRepo.UpdatePaperIndegree(ids, false, nil); err != nil {
		return nil, status.Error(codes.Internal, "failed to decrement paper indegree: "+err.Error())
	}
	return &emptypb.Empty{}, nil
}

// IncExamIndegreeByIds increments exam_indegree counts for given question IDs
func (s *Question) IncExamIndegreeByIds(ids []int64) (*emptypb.Empty, error) {
	if err := s.questionRepo.UpdateExamIndegree(ids, true, nil); err != nil {
		return nil, status.Error(codes.Internal, "failed to decrement exam indegree: "+err.Error())
	}
	return &emptypb.Empty{}, nil
}

// DecExamIndegreeByIds decrements exam_indegree counts for given question IDs
func (s *Question) DecExamIndegreeByIds(ids []int64) (*emptypb.Empty, error) {
	if err := s.questionRepo.UpdateExamIndegree(ids, false, nil); err != nil {
		return nil, status.Error(codes.Internal, "failed to decrement exam indegree: "+err.Error())
	}
	return &emptypb.Empty{}, nil
}

// UpsertTestCases handles bulk creation and updates of test cases
func (s *Question) UpsertTestCases(req *proto.UpsertTestCasesRequest) (*emptypb.Empty, error) {
	question, err := s.questionRepo.GetQuestionTypeByHash(nil, req.QuestionHash)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, status.Error(codes.NotFound, "could not find question")
		}
		return nil, grpcerror.Internal(err, "failed to get question type by hash")
	}

	if question.Type != proto.QuestionType_CODING {
		return nil, status.Error(codes.InvalidArgument, "test cases can only be added to coding questions")
	}

	if len(req.TestCases) > int(constants.MAX_CODING_TEST_CASES_COUNT) {
		return nil, status.Error(codes.InvalidArgument,
			fmt.Sprintf("Number of test cases cannot be more than %d", constants.MAX_CODING_TEST_CASES_COUNT))
	}

	err = s.testcaseRepo.Transaction(func(tx *gorm.DB) error {
		inputDefinitionsLength, err := s.questionRepo.GetInputDefinitionsLength(tx, question.ID)
		if err != nil {
			return grpcerror.Internal(err, "could not get input_definitions length")
		}

		existingTestCases, err := s.testcaseRepo.GetUnscopedTestCasesByQuestionId(tx, question.ID)
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
			dataHash := generateDataHash(content, tc.Hidden)

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
				if err := s.testcaseRepo.DeleteByIds(tx, idsToDelete); err != nil {
					return status.Error(codes.Internal, "failed to delete extra test cases")
				}
			}
		}

		// Bulk create new test cases
		if len(toCreate) > 0 {
			if err := s.testcaseRepo.Create(tx, &toCreate); err != nil {
				return err
			}
		}

		// Bulk update existing test cases (includes reviving soft-deleted ones)
		for _, tc := range toUpdate {
			if err := s.testcaseRepo.UpdateUnscopedTestCase(tx, tc); err != nil {
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

func (s *Question) GetCodingQuestionInputDefinitions(req *proto.GetCodingQuestionInputDefinitionsRequest) (*proto.GetCodingQuestionInputDefinitionsResponse, error) {
	inputDefs, err := s.questionRepo.GetInputDefinitionsByHash(nil, req.QuestionHash)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, status.Error(codes.NotFound, "could not find question")
		}
		if err == sql.ErrNoRows {
			return nil, status.Error(codes.FailedPrecondition, "not a coding question")
		}
		return nil, grpcerror.Internal(err, "failed to fetch input definitions")
	}

	response := proto.GetCodingQuestionInputDefinitionsResponse{
		InputDefinitions: make([]*proto.InputDefinition, len(inputDefs)),
	}
	for i, def := range inputDefs {
		response.InputDefinitions[i] = &proto.InputDefinition{
			VariableName: def.VariableName,
			Type:         def.Type,
			Items:        def.Items,
		}
	}

	return &response, nil
}
