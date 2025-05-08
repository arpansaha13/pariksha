package handlers

import (
	"context"
	"database/sql"
	"strings"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"

	"pariksha/auth/internal/config/db"
	"pariksha/common/pkg/models"
	"pariksha/common/pkg/proto"
	"pariksha/common/pkg/utils"
)

func (s *AuthServer) GetUser(ctx context.Context, req *proto.GetUserRequest) (*proto.UserProfileResponse, error) {
	var user models.User
	if err := db.DB.Take(&user, req.UserId).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, status.Error(codes.NotFound, "user not found")
		}
		return nil, status.Error(codes.Internal, "failed to find user")
	}

	return &proto.UserProfileResponse{
		Id:        user.ID,
		Username:  user.Username,
		Email:     user.Email,
		FirstName: user.FirstName.String,
		LastName:  user.LastName.String,
	}, nil
}

func (s *AuthServer) UpdateUser(ctx context.Context, req *proto.UpdateUserRequest) (*proto.UpdateUserResponse, error) {
	var user models.User
	if err := db.DB.Take(&user, req.UserId).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, status.Error(codes.NotFound, "user not found")
		}
		return nil, status.Error(codes.Internal, "failed to find user")
	}

	isUpdated := false
	notUpdatedFields := make(map[string]string)

	if req.Username != nil && *req.Username != user.Username {
		var existingUser models.User
		if err := db.DB.Where("username = ?", *req.Username).First(&existingUser).Error; err == nil {
			notUpdatedFields["username"] = "username is already taken"
		} else {
			user.Username = *req.Username
			isUpdated = true
		}
	}

	if req.FirstName != nil && *req.FirstName != user.FirstName.String {
		if !utils.IsAlpha(*req.FirstName) {
			return nil, status.Error(codes.InvalidArgument, "first name must contain only alphabets")
		}
		user.FirstName = sql.NullString{String: *req.FirstName, Valid: true}
		isUpdated = true
	}

	if req.LastName != nil && *req.LastName != user.LastName.String {
		if !utils.IsAlpha(*req.LastName) {
			return nil, status.Error(codes.InvalidArgument, "last name must contain only alphabets")
		}
		user.LastName = sql.NullString{String: *req.LastName, Valid: true}
		isUpdated = true
	}

	if isUpdated {
		if err := db.DB.Save(&user).Error; err != nil {
			return nil, status.Error(codes.Internal, "failed to update user info")
		}
	}

	return &proto.UpdateUserResponse{
		User: &proto.UserProfileResponse{
			Id:        user.ID,
			Username:  user.Username,
			Email:     user.Email,
			FirstName: user.FirstName.String,
			LastName:  user.LastName.String,
		},
		NotUpdatedFields: notUpdatedFields,
	}, nil
}

func (s *AuthServer) UpsertUser(ctx context.Context, req *proto.UpsertUserRequest) (*proto.UserProfileResponse, error) {
	var user models.User

	// Try to find existing user by email
	result := db.DB.Where("email = ?", req.Email).First(&user)
	if result.Error != nil {
		if result.Error != gorm.ErrRecordNotFound {
			return nil, status.Error(codes.Internal, "failed to check existing user")
		}

		// Create new user if not found
		username := strings.Split(req.Email, "@")[0]
		user = models.User{
			Email:    req.Email,
			Username: username,
		}

		if req.FirstName != nil {
			user.FirstName = sql.NullString{String: *req.FirstName, Valid: true}
		}
		if req.LastName != nil {
			user.LastName = sql.NullString{String: *req.LastName, Valid: true}
		}

		if err := db.DB.Create(&user).Error; err != nil {
			return nil, status.Error(codes.Internal, "failed to create user")
		}
	} else {
		// Update existing user's name if provided
		isUpdated := false

		if req.FirstName != nil && req.FirstName != &user.FirstName.String {
			user.FirstName = sql.NullString{String: *req.FirstName, Valid: true}
			isUpdated = true
		}
		if req.LastName != nil && req.LastName != &user.LastName.String {
			user.LastName = sql.NullString{String: *req.LastName, Valid: true}
			isUpdated = true
		}

		if isUpdated {
			if err := db.DB.Save(&user).Error; err != nil {
				return nil, status.Error(codes.Internal, "failed to update user info")
			}
		}
	}

	return &proto.UserProfileResponse{
		Id:        user.ID,
		Username:  user.Username,
		Email:     user.Email,
		FirstName: user.FirstName.String,
		LastName:  user.LastName.String,
	}, nil
}
