package handlers

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"pariksha/common/pkg/constants"
	"pariksha/common/pkg/proto"
	"pariksha/paper/internal/config/db"
	"pariksha/paper/internal/models"
)

func TestGetUserPapers(t *testing.T) {
	tests := []struct {
		name         string
		setup        func(t *testing.T)
		userID       int32
		expectedCode codes.Code
		validate     func(t *testing.T, resp *proto.PaperList)
	}{
		{
			name: "Success - Get user papers",
			setup: func(t *testing.T) {
				createTestPaper(t, int(userID))
				createTestPaper(t, int(userID))
			},
			userID:       userID,
			expectedCode: codes.OK,
			validate: func(t *testing.T, resp *proto.PaperList) {
				assert.Equal(t, 2, len(resp.Papers))
				for _, paper := range resp.Papers {
					assert.NotEmpty(t, paper.Title)
					assert.Equal(t, constants.PAPER_OWNERSHIP_TYPE_OWNER, paper.Ownership.Type)
				}
			},
		},
		{
			name:         "Success - No papers",
			userID:       userID,
			expectedCode: codes.OK,
			validate: func(t *testing.T, resp *proto.PaperList) {
				assert.Equal(t, 0, len(resp.Papers))
			},
		},
		{
			name:         "Invalid user ID",
			userID:       0,
			expectedCode: codes.InvalidArgument,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clearTables(t)
			if tt.setup != nil {
				tt.setup(t)
			}

			ctx := createContextWithUserID(tt.userID)
			resp, err := client.GetUserPapers(ctx, &proto.Empty{})

			if tt.expectedCode != codes.OK {
				assert.Equal(t, tt.expectedCode, status.Code(err))
				return
			}

			require.NoError(t, err)
			require.NotNil(t, resp)
			tt.validate(t, resp)
		})
	}
}

func TestCreatePaper(t *testing.T) {
	tests := []struct {
		name         string
		setup        func(t *testing.T)
		userID       int32
		expectedCode codes.Code
		validate     func(t *testing.T, resp *proto.PaperResponse)
	}{
		{
			name:         "Success - Create paper",
			userID:       userID,
			expectedCode: codes.OK,
			validate: func(t *testing.T, resp *proto.PaperResponse) {
				assert.NotZero(t, resp.Id)
				assert.Equal(t, "Untitled Paper", resp.Title)
				assert.Equal(t, constants.PAPER_OWNERSHIP_TYPE_OWNER, resp.Ownership.Type)

				// Verify default category was created
				var categories []models.QuestionCategory
				err := db.DB.Where("paper_id = ?", resp.Id).Find(&categories).Error
				require.NoError(t, err)
				assert.Equal(t, 1, len(categories))
				assert.Equal(t, "Category 1", categories[0].Name)
			},
		},
		{
			name:         "Invalid user ID",
			userID:       0,
			expectedCode: codes.InvalidArgument,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clearTables(t)
			if tt.setup != nil {
				tt.setup(t)
			}

			ctx := createContextWithUserID(tt.userID)
			resp, err := client.CreatePaper(ctx, &proto.Empty{})

			if tt.expectedCode != codes.OK {
				assert.Equal(t, tt.expectedCode, status.Code(err))
				return
			}

			require.NoError(t, err)
			require.NotNil(t, resp)
			tt.validate(t, resp)
		})
	}
}

func TestUpdatePaper(t *testing.T) {
	tests := []struct {
		name         string
		setup        func(t *testing.T) *models.Paper
		userID       int32
		title        string
		expectedCode codes.Code
		validate     func(t *testing.T, paper *models.Paper)
	}{
		{
			name: "Success - Update title",
			setup: func(t *testing.T) *models.Paper {
				paper := createTestPaper(t, int(userID))
				return &paper
			},
			userID:       userID,
			title:        "Updated Title",
			expectedCode: codes.OK,
			validate: func(t *testing.T, paper *models.Paper) {
				var updated models.Paper
				err := db.DB.First(&updated, paper.ID).Error
				require.NoError(t, err)
				assert.Equal(t, "Updated Title", updated.Title)
			},
		},
		{
			name: "Paper not found",
			setup: func(t *testing.T) *models.Paper {
				return &models.Paper{ID: 999}
			},
			userID:       userID,
			title:        "Updated Title",
			expectedCode: codes.NotFound,
		},
		{
			name: "Invalid user ID",
			setup: func(t *testing.T) *models.Paper {
				paper := createTestPaper(t, int(userID))
				return &paper
			},
			userID:       0,
			expectedCode: codes.InvalidArgument,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clearTables(t)
			paper := tt.setup(t)

			ctx := createContextWithUserID(tt.userID)
			_, err := client.UpdatePaper(ctx, &proto.UpdatePaperRequest{
				PaperId: int32(paper.ID),
				Title:   tt.title,
			})

			if tt.expectedCode != codes.OK {
				assert.Equal(t, tt.expectedCode, status.Code(err))
				return
			}

			require.NoError(t, err)
			if tt.validate != nil {
				tt.validate(t, paper)
			}
		})
	}
}

func TestGetPaper(t *testing.T) {
	tests := []struct {
		name         string
		setup        func(t *testing.T) *models.Paper
		userID       int32
		expectedCode codes.Code
		validate     func(t *testing.T, resp *proto.PaperResponse)
	}{
		{
			name: "Success - Get owned paper",
			setup: func(t *testing.T) *models.Paper {
				paper := createTestPaper(t, int(userID))
				return &paper
			},
			userID:       userID,
			expectedCode: codes.OK,
			validate: func(t *testing.T, resp *proto.PaperResponse) {
				assert.NotZero(t, resp.Id)
				assert.Equal(t, "Test Paper", resp.Title)
				assert.Equal(t, constants.PAPER_OWNERSHIP_TYPE_OWNER, resp.Ownership.Type)
				assert.Equal(t, int32(60), resp.DurationMinutes)
			},
		},
		{
			name: "Success - Get shared paper",
			setup: func(t *testing.T) *models.Paper {
				// Create paper owned by user 2
				paper := createTestPaper(t, 2)
				// Add shared access for test user
				ownership := models.PaperOwnership{
					UserID:  int(userID),
					PaperID: paper.ID,
					Type:    constants.PAPER_OWNERSHIP_TYPE_SHARED,
				}
				err := db.DB.Create(&ownership).Error
				require.NoError(t, err)
				return &paper
			},
			userID:       userID,
			expectedCode: codes.OK,
			validate: func(t *testing.T, resp *proto.PaperResponse) {
				assert.NotZero(t, resp.Id)
				assert.Equal(t, constants.PAPER_OWNERSHIP_TYPE_SHARED, resp.Ownership.Type)
			},
		},
		{
			name: "Paper not found",
			setup: func(t *testing.T) *models.Paper {
				return &models.Paper{ID: 999}
			},
			userID:       userID,
			expectedCode: codes.NotFound,
		},
		{
			name: "No access to paper",
			setup: func(t *testing.T) *models.Paper {
				// Create paper owned by user 2 with no sharing
				paper := createTestPaper(t, 2)
				return &paper
			},
			userID:       userID,
			expectedCode: codes.PermissionDenied,
		},
		{
			name: "Invalid user ID",
			setup: func(t *testing.T) *models.Paper {
				paper := createTestPaper(t, int(userID))
				return &paper
			},
			userID:       0,
			expectedCode: codes.InvalidArgument,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clearTables(t)
			paper := tt.setup(t)

			ctx := createContextWithUserID(tt.userID)
			resp, err := client.GetPaper(ctx, &proto.PaperRequest{
				PaperId: int32(paper.ID),
			})

			if tt.expectedCode != codes.OK {
				assert.Equal(t, tt.expectedCode, status.Code(err))
				return
			}

			require.NoError(t, err)
			require.NotNil(t, resp)
			tt.validate(t, resp)
		})
	}
}
