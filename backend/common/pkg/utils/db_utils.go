package utils

import (
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"

	"pariksha/common/pkg/constants"
)

// TransactionHandler wraps a database transaction with common error handling
func TransactionHandler(db *gorm.DB, tx func(*gorm.DB) error) error {
	err := db.Transaction(tx)
	if err != nil {
		// Check if it's already a gRPC status error
		if _, ok := status.FromError(err); ok {
			return err
		}
		return status.Error(codes.Internal, constants.ErrInternalServer)
	}
	return nil
}

// HandleDBError converts database errors to appropriate gRPC status errors
func HandleDBError(err error, notFoundMsg string) error {
	if err == gorm.ErrRecordNotFound {
		return status.Error(codes.NotFound, notFoundMsg)
	}
	return status.Error(codes.Internal, constants.ErrInternalServer)
}

// FindRecord is a generic function to find a record by ID
func FindRecord[T any](db *gorm.DB, id int64, notFoundMsg string) (*T, error) {
	var record T
	if err := db.Take(&record, id).Error; err != nil {
		return nil, HandleDBError(err, notFoundMsg)
	}
	return &record, nil
}
