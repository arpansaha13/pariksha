package services

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"gorm.io/gorm"

	"pariksha/common/pkg/proto"
	"pariksha/common/pkg/types"
	"pariksha/common/pkg/utils"
	"pariksha/common/pkg/utils/grpcerror"
	"pariksha/paper/internal/interservice"
	"pariksha/paper/internal/models"
	"pariksha/paper/internal/repositories"
	"pariksha/paper/internal/utils/validate"
)

type Question struct {
	paperRepo      *repositories.Paper
	paperQuestRepo *repositories.PaperQuestion
	questionIntSvc *interservice.Question
}

func NewQuestion(paperRepo *repositories.Paper, paperQuestRepo *repositories.PaperQuestion, questionIntSvc *interservice.Question) *Question {
	return &Question{
		paperRepo:      paperRepo,
		paperQuestRepo: paperQuestRepo,
		questionIntSvc: questionIntSvc,
	}
}

// GetPaperQuestions handles fetching all questions for a paper
func (s *Question) GetPaperQuestions(ctx context.Context, paperHash string) (*proto.QuestionList, error) {
	paperQuests, err := s.paperQuestRepo.GetAllByPaperHash(nil, paperHash)
	if err != nil {
		return nil, grpcerror.Internal(err, "failed to fetch paper questions")
	}

	questionIDs := make([]types.QuestionID, len(paperQuests))
	for i, pq := range paperQuests {
		questionIDs[i] = pq.QuestionID
	}

	questions, err := s.questionIntSvc.GetQuestionsMetaByIDs(questionIDs)
	if err != nil {
		return nil, grpcerror.Internal(err, "failed to fetch question meta")
	}

	response := &proto.QuestionList{
		Questions: make([]*proto.QuestionMinimal, len(questions)),
	}

	for i, q := range questions {
		pq := paperQuests[i]
		response.Questions[i] = &proto.QuestionMinimal{
			QuestionHash: q.Hash,
			CategoryId:   int64(pq.CategoryID),
			PaperHash:    paperHash,
			Order:        int32(pq.Order),
			RawQuestion:  q.RawQuestion,
		}
	}

	return response, nil
}

// GetPaperQuestion handles fetching a single question with its test cases
func (s *Question) GetPaperQuestion(req *proto.PaperQuestionRequest) (*proto.PaperQuestionResponse, error) {
	question, err := s.questionIntSvc.GetQuestionByHash(req.QuestionHash)
	if err != nil {
		return nil, grpcerror.Internal(err, "failed to fetch question")
	}

	paperQuest, err := s.paperQuestRepo.GetByPaperHashAndQuestionID(nil, req.PaperHash, types.QuestionID(question.Id))
	if err != nil {
		return nil, grpcerror.Internal(err, "failed to fetch paper question")
	}

	return &proto.PaperQuestionResponse{
		QuestionHash: question.Hash,
		RawQuestion:  question.RawQuestion,
		CategoryId:   int64(paperQuest.CategoryID),
		Type:         question.Type,
		PaperHash:    req.PaperHash,
		MaxScore:     int32(paperQuest.MaxScore),
		TestCases:    question.TestCases,
	}, nil
}

// CreatePaperQuestion handles the business logic for creating a new question.
func (s *Question) CreatePaperQuestion(ctx context.Context, req *proto.CreatePaperQuestionRequest) (*proto.CreatePaperQuestionResponse, error) {
	if err := validate.MaxScore(req.MaxScore); err != nil {
		return nil, err
	}

	resp, err := s.questionIntSvc.CreateQuestion(&proto.CreateQuestionRequest{
		RawQuestion: req.RawQuestion,
		Type:        req.Type,
	})
	if err != nil {
		return nil, grpcerror.Internal(err, "failed to create question")
	}

	err = s.paperRepo.Transaction(func(tx *gorm.DB) error {
		// Get paper by hash
		paper, err := s.paperRepo.GetByHash(tx, req.PaperHash)
		if err != nil {
			return utils.HandleDBError(err, "paper not found")
		}

		// Get max order for this category
		maxOrder, err := s.paperQuestRepo.GetMaxQuestionOrder(tx, req.CategoryId)
		if err != nil {
			return grpcerror.Internal(err, "failed to get max order for category")
		}

		paperQuest := models.PaperQuestion{
			PaperID:    paper.ID,
			QuestionID: types.QuestionID(resp.Id),
			CategoryID: types.CategoryID(req.CategoryId),
			Order:      maxOrder + 1,
			MaxScore:   int16(req.MaxScore),
		}

		return s.paperQuestRepo.Create(tx, &paperQuest)
	})

	if err != nil {
		return nil, err
	}

	return &proto.CreatePaperQuestionResponse{
		QuestionHash: resp.Hash,
	}, nil
}

// UpdateQuestion handles question updates with proper locking to prevent race conditions
func (s *Question) UpdatePaperQuestion(req *proto.UpdatePaperQuestionRequest) (*proto.UpdatePaperQuestionResponse, error) {
	if req.MaxScore != nil {
		if err := validate.MaxScore(*req.MaxScore); err != nil {
			return nil, err
		}
	}

	questionMeta, err := s.questionIntSvc.GetQuestionMetaByHash(req.QuestionHash)
	if err != nil {
		return nil, grpcerror.Internal(err, "failed to fetch question")
	}

	paperQuest, err := s.paperQuestRepo.GetByPaperHashAndQuestionID(nil, req.PaperHash, types.QuestionID(questionMeta.Id))
	if err != nil {
		return nil, grpcerror.Internal(err, "failed to get paper question")
	}

	resp, err := s.questionIntSvc.UpdateQuestion(&proto.UpdateQuestionRequest{
		Hash:        req.QuestionHash,
		Type:        req.Type,
		RawQuestion: req.RawQuestion,
	})
	if err != nil {
		return nil, grpcerror.Internal(err, "failed to update question")
	}

	paperQuest.MaxScore = int16(*req.MaxScore)
	paperQuest.QuestionID = types.QuestionID(resp.Id)

	err = s.paperQuestRepo.Save(nil, paperQuest)
	if err != nil {
		return nil, grpcerror.Internal(err, "failed to save paper question")
	}

	return &proto.UpdatePaperQuestionResponse{
		QuestionHash: resp.Hash,
	}, nil
}

// DeleteQuestion handles the business logic for deleting a question
func (s *Question) DeletePaperQuestion(paperHash string, questionHash string) error {
	return s.paperRepo.Transaction(func(tx *gorm.DB) error {
		question, err := s.questionIntSvc.GetQuestionMetaByHash(questionHash)
		if err != nil {
			return grpcerror.Internal(err, "faled to fetch question id")
		}

		return s.paperQuestRepo.DeleteByID(nil, types.QuestionID(question.Id))
	})
}

// ReorderQuestions handles the business logic for reordering questions
func (s *Question) ReorderQuestions(categoryID int64, questionHashes []string) error {
	return s.paperRepo.Transaction(func(tx *gorm.DB) error {
		questionIDs, err := s.questionIntSvc.GetQuestionIDsByHashes(questionHashes)
		if err != nil {
			return grpcerror.Internal(err, "failed to fetch question ids")
		}

		// Verify all questions belong to the category
		count, err := s.paperQuestRepo.ValidateCategoryQuestions(tx, categoryID, questionIDs)
		if err != nil {
			return err
		}

		if int(count) != len(questionHashes) {
			return status.Error(codes.InvalidArgument, "invalid question hashes")
		}

		// Update orders
		for i, questionID := range questionIDs {
			if err := s.paperQuestRepo.UpdateOrder(tx, questionID, int16(i+1)); err != nil {
				return err
			}
		}

		return nil
	})
}

func (s *Question) GetBoilerplate(req *proto.GetPaperBoilerplateRequest) (*proto.BoilerplateResponse, error) {
	return s.questionIntSvc.GetBoilerplate(&proto.GetBoilerplateRequest{
		QuestionHash: req.QuestionHash,
		LanguageId:   req.LanguageId,
	})
}

func (s *Question) UpsertTestCases(req *proto.UpsertPaperTestCasesRequest) (*emptypb.Empty, error) {
	return s.questionIntSvc.UpsertTestCases(&proto.UpsertTestCasesRequest{
		QuestionHash: req.QuestionHash,
		TestCases:    req.TestCases,
	})
}

func (s *Question) GetPaperQuestionsMeta(req *proto.PaperRequest) (*proto.PaperQuestionsMeta, error) {
	paperQuests, err := s.paperQuestRepo.GetAllByPaperHash(nil, req.PaperHash)
	if err != nil {
		return nil, grpcerror.Internal(err, "failed to fetch paper questions")
	}

	response := &proto.PaperQuestionsMeta{
		Questions: make([]*proto.PaperQuestionMeta, len(paperQuests)),
	}

	for i, pq := range paperQuests {
		response.Questions[i] = &proto.PaperQuestionMeta{
			Id:         int64(pq.QuestionID),
			CategoryId: int64(pq.CategoryID),
			Order:      int32(pq.Order),
			MaxScore:   int32(pq.MaxScore),
		}
	}

	return response, nil
}
