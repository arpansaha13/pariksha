package services

import (
	"context"

	"go.uber.org/zap"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"

	"pariksha/common/pkg/constants"
	"pariksha/common/pkg/logging"
	"pariksha/common/pkg/proto"
	"pariksha/common/pkg/types"
	"pariksha/common/pkg/utils"
	"pariksha/common/pkg/utils/generate"
	"pariksha/common/pkg/utils/grpcerror"
	"pariksha/paper/internal/interservice"
	"pariksha/paper/internal/models"
	"pariksha/paper/internal/repositories"
	"pariksha/paper/internal/utils/validate"
)

type Paper struct {
	paperRepo      *repositories.Paper
	paperCatRepo   *repositories.PaperCategory
	paperPermRepo  *repositories.PaperPermission
	paperQuestRepo *repositories.PaperQuestion
	questionIntSvc *interservice.Question
}

func NewPaper(
	paperRepo *repositories.Paper,
	paperCatRepo *repositories.PaperCategory,
	paperPermRepo *repositories.PaperPermission,
	paperQuestRepo *repositories.PaperQuestion,
	questionIntSvc *interservice.Question,
) *Paper {
	return &Paper{
		paperRepo:      paperRepo,
		paperCatRepo:   paperCatRepo,
		paperPermRepo:  paperPermRepo,
		paperQuestRepo: paperQuestRepo,
		questionIntSvc: questionIntSvc,
	}
}

// GetUserPapers handles the business logic for fetching user's papers
func (s *Paper) GetUserPapers(ctx context.Context, userID types.UserID) (*proto.PaperList, error) {
	logger := logging.GetLogger()
	logger.Debug("GetUserPapers called", zap.Int64("user_id", int64(userID)))

	papers, err := s.paperRepo.GetAllByUserId(nil, userID)
	if err != nil {
		logger.Error("failed to fetch user papers", zap.Int64("user_id", int64(userID)), zap.Error(err))
		return nil, status.Error(codes.Internal, constants.ErrInternalServer)
	}

	response := &proto.PaperList{
		Papers: make([]*proto.PaperResponse, len(papers)),
	}
	for i, p := range papers {
		paperResp, err := s.GetPaper(ctx, p.Hash)
		if err != nil {
			return nil, grpcerror.Internal(err, "failed to get paper response")
		}

		response.Papers[i] = paperResp
	}

	return response, nil
}

// GetUserPapers handles the business logic for fetching user's papers
func (s *Paper) GetPaper(ctx context.Context, paperHash string) (*proto.PaperResponse, error) {
	logger := logging.GetLogger()
	logger.Debug("GetPaper called", zap.String("paper_hash", paperHash))

	paper, err := s.paperRepo.GetByHash(nil, paperHash)
	if err != nil {
		logger.Error("failed to get paper", zap.String("paper_hash", paperHash), zap.Error(err))
		return nil, grpcerror.Internal(err, "failed to get paper")
	}

	paperQuests, err := s.paperQuestRepo.GetAllByPaperHash(nil, paperHash)
	if err != nil {
		return nil, grpcerror.Internal(err, "failed to get paper questions")
	}

	questionIDs := make([]types.QuestionID, len(paperQuests))
	totalMaxScore := int32(0)
	for i, pq := range paperQuests {
		totalMaxScore += int32(pq.MaxScore)
		questionIDs[i] = pq.QuestionID
	}

	questions, err := s.questionIntSvc.GetQuestionsMetaByIDs(ctx, questionIDs)
	questionCounts := proto.QuestionCount{
		Mcq:        0,
		Subjective: 0,
		Coding:     0,
	}

	for _, q := range questions {
		switch q.Type {
		case proto.QuestionType_MCQ:
			questionCounts.Mcq++
		case proto.QuestionType_SUBJECTIVE:
			questionCounts.Subjective++
		case proto.QuestionType_CODING:
			questionCounts.Coding++
		}
	}

	return &proto.PaperResponse{
		PaperHash:       paper.Hash,
		Title:           paper.Title,
		MaxScore:        totalMaxScore,
		DurationMinutes: int32(paper.DurationMinutes),
		CreatedBy:       int64(paper.CreatedBy),
		QuestionCounts:  &questionCounts,
	}, nil
}

// CreatePaper handles the business logic for creating a new paper
func (s *Paper) CreatePaper(ctx context.Context, userID types.UserID) (*proto.CreatePaperResponse, error) {
	var paper models.Paper

	logger := logging.GetLogger()
	logger.Debug("CreatePaper called", zap.Int64("user_id", int64(userID)))

	err := s.paperRepo.Transaction(func(tx *gorm.DB) error {
		paper = models.Paper{CreatedBy: userID}
		if err := s.paperRepo.Create(tx, &paper, userID); err != nil {
			return grpcerror.Internal(err, "failed to create paper")
		}

		if err := s.paperPermRepo.Create(tx, paper.ID, userID); err != nil {
			return grpcerror.Internal(err, "failed to create paper permission")
		}

		// Create default category
		category, err := s.questionIntSvc.CreateCategory(ctx, "Category 1")
		if err != nil {
			return grpcerror.Internal(err, "failed to create default category")
		}
		if err := s.paperCatRepo.Create(tx, &models.PaperCategory{
			PaperID:    paper.ID,
			CategoryID: types.CategoryID(category.Id),
			Order:      1,
		}); err != nil {
			return grpcerror.Internal(err, "failed to create category")
		}

		paper.Hash = generate.HMACHash(int64(paper.ID))
		if err := s.paperRepo.UpdateHash(tx, &paper, paper.Hash); err != nil {
			return status.Error(codes.Internal, "failed to store paper hash")
		}

		return nil
	})

	if err != nil {
		logger.Error("CreatePaper transaction failed", zap.Error(err))
		return nil, err
	}

	return &proto.CreatePaperResponse{
		PaperHash: paper.Hash,
	}, nil
}

// UpdatePaper handles the business logic for updating a paper
func (s *Paper) UpdatePaper(ctx context.Context, req *proto.UpdatePaperRequest) error {
	return s.paperRepo.Transaction(func(tx *gorm.DB) error {
		logger, ok := logging.GetLoggerFromContext(ctx)
		if !ok {
			logger = logging.GetLogger()
		}
		logger.Debug("UpdatePaper called", zap.String("paper_hash", req.PaperHash))
		paper, err := s.paperRepo.GetByHash(tx, req.PaperHash)
		if err != nil {
			return utils.HandleDBError(err, "paper not found")
		}

		isUpdated := false

		if req.Title != nil && req.GetTitle() != "" && req.GetTitle() != paper.Title {
			paper.Title = req.GetTitle()
			isUpdated = true
		}

		if req.DurationMinutes != nil {
			if err := validate.PaperDuration(req.GetDurationMinutes()); err != nil {
				return err
			}
			if int16(req.GetDurationMinutes()) != paper.DurationMinutes {
				paper.DurationMinutes = int16(req.GetDurationMinutes())
				isUpdated = true
			}
		}

		if isUpdated {
			if err := s.paperRepo.Update(tx, paper); err != nil {
				return status.Error(codes.Internal, constants.ErrInternalServer)
			}
		}

		return nil
	})
}

// DeletePapers handles the business logic for deleting multiple papers
func (s *Paper) DeletePapers(ctx context.Context, hashes []string) error {
	logger, ok := logging.GetLoggerFromContext(ctx)
	if !ok {
		logger = logging.GetLogger()
	}
	logger.Debug("DeletePapers called", zap.Int("count", len(hashes)))

	return s.paperRepo.Transaction(func(tx *gorm.DB) error {
		paperIDs, err := s.paperRepo.GetIDsByHashes(tx, hashes)
		if err != nil {
			return status.Error(codes.Internal, "failed to fetch paper IDs")
		}

		// Ignore non-existent papers
		if len(paperIDs) == 0 {
			return nil
		}

		// Get question IDs before deletion
		var questionIDs []types.QuestionID
		if err := tx.Model(&models.PaperQuestion{}).
			Where("paper_id IN ?", paperIDs).
			Pluck("question_id", &questionIDs).Error; err != nil {
			return grpcerror.Internal(err, "failed to fetch question IDs")
		}

		// Decrease question paper indegree if there are questions
		if len(questionIDs) > 0 {
			if err := s.questionIntSvc.DecQuestionPaperIndegreeByIds(ctx, questionIDs); err != nil {
				return grpcerror.Internal(err, "failed to decrease question paper indegree")
			}
		}
		if err := s.paperPermRepo.BulkDeleteByPaperIDs(tx, paperIDs); err != nil {
			return grpcerror.Internal(err, "failed to delete paper permissions")
		}

		if err := s.paperQuestRepo.BulkDeleteByPaperIDs(tx, paperIDs); err != nil {
			return grpcerror.Internal(err, "failed to delete paper questions")
		}

		if err := s.paperCatRepo.BulkDeleteByPaperIDs(tx, paperIDs); err != nil {
			return grpcerror.Internal(err, "failed to delete paper categories")
		}

		if err := s.paperRepo.BulkDelete(tx, paperIDs); err != nil {
			return grpcerror.Internal(err, "failed to delete papers")
		}

		return nil
	})
}
