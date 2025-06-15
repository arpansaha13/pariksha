package handlers

import (
	"context"
	"database/sql"

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
	"pariksha/paper/internal/interceptors"
	"pariksha/paper/internal/utils/validate"
)

type PaperServer struct {
	proto.UnimplementedPaperServer
}

func (s *PaperServer) GetUserPapers(ctx context.Context, _ *proto.Empty) (*proto.PaperList, error) {
	userID, err := utils.GetUserIDFromMetadata(ctx)
	if err != nil {
		return nil, err
	}

	var papers []models.Paper
	err = db.DB.
		Select("papers.id, papers.hash, papers.title, papers.max_score, papers.duration_minutes, papers.question_counts, papers.created_by").
		Joins("INNER JOIN permissions ON permissions.paper_id = papers.id").
		Where("permissions.user_id = ?", userID).
		Find(&papers).Error

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

func (s *PaperServer) CreatePaper(ctx context.Context, _ *proto.Empty) (*proto.PaperResponse, error) {
	userID, err := utils.GetUserIDFromMetadata(ctx)
	if err != nil {
		return nil, err
	}

	var paper models.Paper
	err = utils.TransactionHandler(db.DB, func(tx *gorm.DB) error {
		paper = models.Paper{CreatedBy: userID} // Will use database default for Title
		if err := tx.Create(&paper).Error; err != nil {
			return status.Error(codes.Internal, constants.ErrInternalServer)
		}

		// Generate and store hash
		paper.Hash = generate.HMACHash(int64(paper.ID))
		if err := tx.Model(&paper).Update("hash", paper.Hash).Error; err != nil {
			return status.Error(codes.Internal, "failed to store paper hash")
		}

		// Create default category
		defaultCategory := models.QuestionCategory{
			PaperID: sql.NullInt64{Int64: int64(paper.ID), Valid: true},
			Name:    "Category 1",
			Order:   1,
		}
		if err := tx.Create(&defaultCategory).Error; err != nil {
			return status.Error(codes.Internal, constants.ErrInternalServer)
		}

		// Create permissions entry with write access
		permissions := models.PaperPermission{
			PaperID: paper.ID,
			UserID:  userID,
		}
		permissions.SetWrite()

		if err := tx.Create(&permissions).Error; err != nil {
			return status.Error(codes.Internal, constants.ErrInternalServer)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return paperToProto(paper), nil
}

func (s *PaperServer) GetPaper(ctx context.Context, req *proto.PaperRequest) (*proto.PaperResponse, error) {
	var paper models.Paper
	if err := db.DB.Where("hash = ?", req.PaperHash).First(&paper).Error; err != nil {
		return nil, utils.HandleDBError(err, "paper not found")
	}

	return paperToProto(paper), nil
}

func (s *PaperServer) UpdatePaper(ctx context.Context, req *proto.UpdatePaperRequest) (*proto.Empty, error) {
	err := utils.TransactionHandler(db.DB, func(tx *gorm.DB) error {
		var paper models.Paper
		if err := tx.Where("hash = ?", req.PaperHash).First(&paper).Error; err != nil {
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
			if err := tx.Save(&paper).Error; err != nil {
				return status.Error(codes.Internal, constants.ErrInternalServer)
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return &proto.Empty{}, nil
}

func (s *PaperServer) GetPaperPermissions(ctx context.Context, req *proto.PaperRequest) (*proto.PaperPermissionsResponse, error) {
	perm, err := interceptors.GetPermissionFromContext(ctx)
	if err != nil {
		return nil, err
	}

	return &proto.PaperPermissionsResponse{
		CanRead:  perm.CanRead(),
		CanWrite: perm.CanWrite(),
	}, nil
}

func (s *PaperServer) DeletePapers(ctx context.Context, req *proto.DeletePapersRequest) (*proto.Empty, error) {
	err := utils.TransactionHandler(db.DB, func(tx *gorm.DB) error {
		// Get paper IDs from hashes
		var paperIDs []types.PaperID
		if err := tx.Model(&models.Paper{}).
			Where("hash IN ?", req.PaperHashes).
			Pluck("id", &paperIDs).Error; err != nil {
			return status.Error(codes.Internal, "failed to fetch paper IDs")
		}

		// Delete all specified papers
		if err := tx.Where("id IN ?", paperIDs).Delete(&models.Paper{}).Error; err != nil {
			return status.Error(codes.Internal, err.Error())
		}

		// Delete all non-locked questions for these papers
		if err := tx.Where("paper_id IN ? AND locked = ?", paperIDs, false).
			Delete(&models.Question{}).Error; err != nil {
			return status.Error(codes.Internal, err.Error())
		}

		// Delete all non-locked categories for these papers
		if err := tx.Where("paper_id IN ? AND locked = ?", paperIDs, false).
			Delete(&models.QuestionCategory{}).Error; err != nil {
			return status.Error(codes.Internal, err.Error())
		}

		// Delete all permissions for these papers
		if err := tx.Where("paper_id IN ?", paperIDs).
			Delete(&models.PaperPermission{}).Error; err != nil {
			return status.Error(codes.Internal, err.Error())
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return &proto.Empty{}, nil
}

func (s *PaperServer) GetQuestionHashes(ctx context.Context, req *proto.GetQuestionHashesRequest) (*proto.GetQuestionHashesResponse, error) {
	if len(req.QuestionIds) == 0 {
		return &proto.GetQuestionHashesResponse{}, nil
	}

	var questions []models.Question
	if err := db.DB.Select("id", "hash").
		Where("id IN ?", req.QuestionIds).
		Find(&questions).Error; err != nil {
		return nil, status.Error(codes.Internal, constants.ErrInternalServer)
	}

	// Create a map for quick hash lookups
	hashMap := make(map[int64]string, len(questions))
	for _, q := range questions {
		hashMap[int64(q.ID)] = q.Hash
	}

	// Create response maintaining same sequence as request
	hashes := make([]string, len(req.QuestionIds))
	for i, id := range req.QuestionIds {
		if hash, ok := hashMap[id]; ok {
			hashes[i] = hash
		}
	}

	return &proto.GetQuestionHashesResponse{
		QuestionHashes: hashes,
	}, nil
}

func (s *PaperServer) GetQuestionIds(ctx context.Context, req *proto.GetQuestionIdsRequest) (*proto.GetQuestionIdsResponse, error) {
	if len(req.QuestionHashes) == 0 {
		return &proto.GetQuestionIdsResponse{}, nil
	}

	var questions []models.Question
	if err := db.DB.Select("id", "hash").
		Where("hash IN ?", req.QuestionHashes).
		Find(&questions).Error; err != nil {
		return nil, status.Error(codes.Internal, constants.ErrInternalServer)
	}

	// Create a map for quick ID lookups
	idMap := make(map[string]int64, len(questions))
	for _, q := range questions {
		idMap[q.Hash] = int64(q.ID)
	}

	// Create response maintaining same sequence as request
	ids := make([]int64, len(req.QuestionHashes))
	for i, hash := range req.QuestionHashes {
		if id, ok := idMap[hash]; ok {
			ids[i] = id
		}
	}

	return &proto.GetQuestionIdsResponse{
		QuestionIds: ids,
	}, nil
}
