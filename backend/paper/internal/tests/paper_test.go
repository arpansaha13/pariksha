package tests

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/types/known/emptypb"

	"pariksha/common/pkg/proto"
	"pariksha/common/pkg/test"
	"pariksha/common/pkg/types"
	models "pariksha/paper/internal/domain"
)

func TestGetUserPapers(t *testing.T) {
	type SetupReturn struct {
		Papers []models.Paper
	}

	testCases := []test.TestCase[*emptypb.Empty, *proto.PaperList, *SetupReturn]{
		{
			Name:     "Get multiple papers for user",
			Metadata: defaultMetadata,
			Setup: func(t *testing.T) *SetupReturn {
				papers := createDefaultTestPapers(t, 2)

				return &SetupReturn{Papers: papers}
			},
			GetRequest: func(setupData *SetupReturn) *emptypb.Empty {
				return &emptypb.Empty{}
			},
			ExpectedCode: codes.OK,
			Validate: func(t *testing.T, resp *proto.PaperList, setupData *SetupReturn) {
				require.Equal(t, len(setupData.Papers), len(resp.Papers))
				for i, paper := range resp.Papers {
					assert.Equal(t, setupData.Papers[i].Title, paper.Title)
					assert.Equal(t, int32(setupData.Papers[i].DurationMinutes), paper.DurationMinutes)
					assert.Equal(t, setupData.Papers[i].Hash, paper.PaperHash)
					assert.Equal(t, int64(setupData.Papers[i].CreatedBy), paper.CreatedBy)
					assert.NotNil(t, paper.QuestionCounts)
				}
			},
		},
		{
			Name: "No papers for user",
			Metadata: metadata.MD{
				"user_id": []string{"2"},
			},
			Setup: func(t *testing.T) *SetupReturn {
				papers := createDefaultTestPapers(t, 1)

				return &SetupReturn{Papers: papers}
			},
			GetRequest: func(setupData *SetupReturn) *emptypb.Empty {
				return &emptypb.Empty{}
			},
			ExpectedCode: codes.OK,
			Validate: func(t *testing.T, resp *proto.PaperList, setupData *SetupReturn) {
				assert.Empty(t, resp.Papers)
			},
		},
		{
			Name: "Missing user ID in metadata",
			GetRequest: func(setupData *SetupReturn) *emptypb.Empty {
				return &emptypb.Empty{}
			},
			ExpectedCode: codes.Unauthenticated,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			clearTables(t)
			test.Runner(t, tc, client.GetUserPapers)
		})
	}
}

func TestGetPaper(t *testing.T) {
	type SetupReturn struct {
		Paper     models.Paper
		Questions []models.PaperQuestion
	}

	testCases := []test.TestCase[*proto.PaperRequest, *proto.PaperResponse, *SetupReturn]{
		{
			Name:     "Get existing paper with questions",
			Metadata: defaultMetadata,
			Setup: func(t *testing.T) *SetupReturn {
				papers := createDefaultTestPapers(t, 1)

				questions := createTestQuestions(t, []models.PaperQuestion{
					{
						PaperID:    papers[0].ID,
						QuestionID: 1,
						CategoryID: 1,
						MaxScore:   10,
					},
					{
						PaperID:    papers[0].ID,
						QuestionID: 2,
						CategoryID: 1,
						MaxScore:   15,
					},
				})

				return &SetupReturn{
					Paper:     papers[0],
					Questions: questions,
				}
			},
			GetRequest: func(setupData *SetupReturn) *proto.PaperRequest {
				return &proto.PaperRequest{
					PaperHash: setupData.Paper.Hash,
				}
			},
			ExpectedCode: codes.OK,
			Validate: func(t *testing.T, resp *proto.PaperResponse, setupData *SetupReturn) {
				assert.Equal(t, setupData.Paper.Hash, resp.PaperHash)
				assert.Equal(t, setupData.Paper.Title, resp.Title)
				assert.Equal(t, int32(setupData.Paper.DurationMinutes), resp.DurationMinutes)
				assert.Equal(t, int64(setupData.Paper.CreatedBy), resp.CreatedBy)

				// Total max score should be sum of all question scores
				expectedMaxScore := int32(0)
				for _, q := range setupData.Questions {
					expectedMaxScore += int32(q.MaxScore)
				}
				assert.Equal(t, expectedMaxScore, resp.MaxScore)

				assert.NotNil(t, resp.QuestionCounts)
			},
		},
		{
			Name:     "Non-existent paper hash",
			Metadata: defaultMetadata,
			GetRequest: func(setupData *SetupReturn) *proto.PaperRequest {
				return &proto.PaperRequest{
					PaperHash: "nonexistent",
				}
			},
			ExpectedCode: codes.PermissionDenied,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			clearTables(t)
			test.Runner(t, tc, client.GetPaper)
		})
	}
}

func TestCreatePaper(t *testing.T) {
	testCases := []test.TestCase[*emptypb.Empty, *proto.CreatePaperResponse, any]{
		{
			Name:     "Create paper with default values",
			Metadata: defaultMetadata,
			GetRequest: func(setupData any) *emptypb.Empty {
				return &emptypb.Empty{}
			},
			ExpectedCode: codes.OK,
			Validate: func(t *testing.T, resp *proto.CreatePaperResponse, setupData any) {
				assert.NotEmpty(t, resp.PaperHash)

				var paper models.Paper
				err := dbInst.Where("hash = ?", resp.PaperHash).Take(&paper).Error
				require.NoError(t, err)

				assert.Equal(t, "Untitled Paper", paper.Title)
				assert.Equal(t, defaultUserID, paper.CreatedBy)

				// Verify permission was created
				var perm models.PaperPermission
				err = dbInst.Where("paper_id = ? AND user_id = ?", paper.ID, defaultUserID).Take(&perm).Error
				require.NoError(t, err)
				assert.True(t, perm.CanWrite())

				// Verify default category was created
				var category models.PaperCategory
				err = dbInst.Where("paper_id = ?", paper.ID).Take(&category).Error
				require.NoError(t, err)
				assert.Equal(t, int16(1), category.Order)
			},
		},
		{
			Name: "Missing user ID in metadata",
			GetRequest: func(setupData any) *emptypb.Empty {
				return &emptypb.Empty{}
			},
			ExpectedCode: codes.Unauthenticated,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			clearTables(t)
			test.Runner(t, tc, client.CreatePaper)
		})
	}
}

func TestUpdatePaper(t *testing.T) {
	type SetupReturn struct {
		Paper models.Paper
	}

	testCases := []test.TestCase[*proto.UpdatePaperRequest, *emptypb.Empty, *SetupReturn]{
		{
			Name:     "Update paper title and duration",
			Metadata: defaultMetadata,
			Setup: func(t *testing.T) *SetupReturn {
				papers := createTestPapers(t, []models.Paper{
					{
						Title: "Old Title",
					},
				})
				return &SetupReturn{Paper: papers[0]}
			},
			GetRequest: func(setupData *SetupReturn) *proto.UpdatePaperRequest {
				newTitle := "New Title"
				newDuration := int32(90)
				return &proto.UpdatePaperRequest{
					PaperHash:       setupData.Paper.Hash,
					Title:           &newTitle,
					DurationMinutes: &newDuration,
				}
			},
			ExpectedCode: codes.OK,
			Validate: func(t *testing.T, resp *emptypb.Empty, setupData *SetupReturn) {
				var paper models.Paper
				err := dbInst.Take(&paper, setupData.Paper.ID).Error
				require.NoError(t, err)
				assert.Equal(t, "New Title", paper.Title)
				assert.Equal(t, int16(90), paper.DurationMinutes)
			},
		},
		{
			Name:     "Update title only",
			Metadata: defaultMetadata,
			Setup: func(t *testing.T) *SetupReturn {
				papers := createTestPapers(t, []models.Paper{
					{
						Title: "Old Title",
					},
				})
				return &SetupReturn{Paper: papers[0]}
			},
			GetRequest: func(setupData *SetupReturn) *proto.UpdatePaperRequest {
				newTitle := "New Title"
				return &proto.UpdatePaperRequest{
					PaperHash: setupData.Paper.Hash,
					Title:     &newTitle,
				}
			},
			ExpectedCode: codes.OK,
			Validate: func(t *testing.T, resp *emptypb.Empty, setupData *SetupReturn) {
				var paper models.Paper
				err := dbInst.Take(&paper, setupData.Paper.ID).Error
				require.NoError(t, err)
				assert.Equal(t, "New Title", paper.Title)
				assert.Equal(t, setupData.Paper.DurationMinutes, paper.DurationMinutes)
			},
		},
		{
			Name:     "Invalid duration minutes",
			Metadata: defaultMetadata,
			Setup: func(t *testing.T) *SetupReturn {
				papers := createTestPapers(t, []models.Paper{
					{
						Title: "Test Paper",
					},
				})
				return &SetupReturn{Paper: papers[0]}
			},
			GetRequest: func(setupData *SetupReturn) *proto.UpdatePaperRequest {
				invalidDuration := int32(1500) // > 1440 minutes (24 hours)
				return &proto.UpdatePaperRequest{
					PaperHash:       setupData.Paper.Hash,
					DurationMinutes: &invalidDuration,
				}
			},
			ExpectedCode: codes.InvalidArgument,
		},
		{
			Name:     "Non-existent paper hash",
			Metadata: defaultMetadata,
			GetRequest: func(setupData *SetupReturn) *proto.UpdatePaperRequest {
				newTitle := "New Title"
				return &proto.UpdatePaperRequest{
					PaperHash: "nonexistent",
					Title:     &newTitle,
				}
			},
			ExpectedCode: codes.PermissionDenied,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			clearTables(t)
			test.Runner(t, tc, client.UpdatePaper)
		})
	}
}

func TestDeletePapers(t *testing.T) {
	type SetupReturn struct {
		Papers    []models.Paper
		Questions []models.PaperQuestion
	}

	testCases := []test.TestCase[*proto.DeletePapersRequest, *emptypb.Empty, *SetupReturn]{
		{
			Name:     "Delete multiple papers with questions",
			Metadata: defaultMetadata,
			Setup: func(t *testing.T) *SetupReturn {
				papers := createDefaultTestPapers(t, 2)

				questions := createTestQuestions(t, []models.PaperQuestion{
					{
						PaperID:    papers[0].ID,
						QuestionID: 1,
						CategoryID: 1,
						MaxScore:   10,
					},
					{
						PaperID:    papers[1].ID,
						QuestionID: 2,
						CategoryID: 1,
						MaxScore:   15,
					},
				})

				return &SetupReturn{
					Papers:    papers,
					Questions: questions,
				}
			},
			GetRequest: func(setupData *SetupReturn) *proto.DeletePapersRequest {
				hashes := make([]string, len(setupData.Papers))
				for i, p := range setupData.Papers {
					hashes[i] = p.Hash
				}
				return &proto.DeletePapersRequest{
					PaperHashes: hashes,
				}
			},
			ExpectedCode: codes.OK,
			Validate: func(t *testing.T, resp *emptypb.Empty, setupData *SetupReturn) {
				// Verify papers are deleted
				var count int64
				err := dbInst.Model(&models.Paper{}).Where("id IN ?", []types.PaperID{
					setupData.Papers[0].ID,
					setupData.Papers[1].ID,
				}).Count(&count).Error
				require.NoError(t, err)
				assert.Equal(t, int64(0), count)

				// Verify paper questions are deleted
				err = dbInst.Model(&models.PaperQuestion{}).Where("paper_id IN ?", []types.PaperID{
					setupData.Papers[0].ID,
					setupData.Papers[1].ID,
				}).Count(&count).Error
				require.NoError(t, err)
				assert.Equal(t, int64(0), count)

				// Verify paper permissions are deleted
				err = dbInst.Model(&models.PaperPermission{}).Where("paper_id IN ?", []types.PaperID{
					setupData.Papers[0].ID,
					setupData.Papers[1].ID,
				}).Count(&count).Error
				require.NoError(t, err)
				assert.Equal(t, int64(0), count)
			},
		},
		{
			Name:     "Empty paper hashes",
			Metadata: defaultMetadata,
			GetRequest: func(setupData *SetupReturn) *proto.DeletePapersRequest {
				return &proto.DeletePapersRequest{
					PaperHashes: []string{},
				}
			},
			ExpectedCode: codes.InvalidArgument,
		},
		{
			Name:     "Non-existent paper hashes",
			Metadata: defaultMetadata,
			GetRequest: func(setupData *SetupReturn) *proto.DeletePapersRequest {
				return &proto.DeletePapersRequest{
					PaperHashes: []string{"nonexistent1", "nonexistent2"},
				}
			},
			ExpectedCode: codes.OK, // Should succeed even if papers don't exist
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			clearTables(t)
			test.Runner(t, tc, client.DeletePapers)
		})
	}
}
