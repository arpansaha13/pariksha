package services

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"

	"pariksha/common/pkg/constants"
	"pariksha/common/pkg/models"
	"pariksha/common/pkg/proto"
	"pariksha/common/pkg/types"
	"pariksha/common/pkg/utils"
	"pariksha/common/pkg/utils/generate"
	"pariksha/paper/internal/config/db"
	"pariksha/paper/internal/repositories"
	"pariksha/paper/internal/utils/validate"
)

type Paper struct {
	paperRepo *repositories.Paper
}

func NewPaper(paperRepo *repositories.Paper) *Paper {
	return &Paper{paperRepo: paperRepo}
}

// GetUserPapers handles the business logic for fetching user's papers
func (s *Paper) GetUserPapers(ctx context.Context, userID types.UserID) (*proto.PaperList, error) {
	papers, err := s.paperRepo.GetAllByUserId(nil, userID)
	if err != nil {
		return nil, status.Error(codes.Internal, constants.ErrInternalServer)
	}

	response := &proto.PaperList{
		Papers: make([]*proto.PaperResponse, len(papers)),
	}
	for i, paper := range papers {
		response.Papers[i] = paperToProto(paper)
	}

	return response, nil
}

// CreatePaper handles the business logic for creating a new paper
func (s *Paper) CreatePaper(ctx context.Context, userID types.UserID) (*proto.PaperResponse, error) {
	var paper models.Paper

	err := utils.TransactionHandler(db.DB, func(tx *gorm.DB) error {
		paper = models.Paper{CreatedBy: userID}
		if err := s.paperRepo.Create(tx, &paper, userID); err != nil {
			return status.Error(codes.Internal, constants.ErrInternalServer)
		}

		paper.Hash = generate.HMACHash(int64(paper.ID))
		if err := s.paperRepo.UpdateHash(tx, &paper, paper.Hash); err != nil {
			return status.Error(codes.Internal, "failed to store paper hash")
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return paperToProto(paper), nil
}

// UpdatePaper handles the business logic for updating a paper
func (s *Paper) UpdatePaper(ctx context.Context, req *proto.UpdatePaperRequest) error {
	return utils.TransactionHandler(db.DB, func(tx *gorm.DB) error {
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
	return utils.TransactionHandler(db.DB, func(tx *gorm.DB) error {
		paperIDs, err := s.paperRepo.GetIDsByHashes(tx, hashes)
		if err != nil {
			return status.Error(codes.Internal, "failed to fetch paper IDs")
		}

		if err := s.paperRepo.BulkDelete(tx, paperIDs); err != nil {
			return status.Error(codes.Internal, err.Error())
		}

		return nil
	})
}
