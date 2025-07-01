package paper

import (
	"context"
	"fmt"

	"google.golang.org/grpc/metadata"

	"pariksha/common/pkg/constants"
	"pariksha/common/pkg/proto"
	"pariksha/common/pkg/types"
	"pariksha/exam/internal/config/env"
)

// FetchCategoriesByIds returns category details for the given category IDs.
// Returns error if any of the category IDs don't have corresponding categories.
var FetchCategoriesByIds = fetchCategoriesByIds

func fetchCategoriesByIds(typedCategoryIDs []types.CategoryID) ([]*proto.CategoryBatchItem, error) {
	ensurePaperService()

	categoryIDs := make([]int64, len(typedCategoryIDs))
	for i, id := range typedCategoryIDs {
		categoryIDs[i] = int64(id)
	}

	// Create metadata with exam token
	md := metadata.New(map[string]string{
		constants.X_EXAM_API_TOKEN: env.EXAM_API_TOKEN,
	})
	ctx := metadata.NewOutgoingContext(context.Background(), md)

	// Get categories
	resp, err := pSvc.client.GetCategoriesByIds(ctx, &proto.GetCategoriesByIdsRequest{
		CategoryIds: categoryIDs,
	})
	if err != nil {
		return nil, err
	}

	// Verify we got all categories
	if len(resp.Categories) != len(categoryIDs) {
		return nil, fmt.Errorf("failed to fetch all categories")
	}

	return resp.Categories, nil
}
