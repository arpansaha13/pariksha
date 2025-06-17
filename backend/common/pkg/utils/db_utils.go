package utils

import (
	"database/sql"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
)

// TransactionHandler wraps a database transaction with common error handling
func TransactionHandler(db *gorm.DB, fc func(*gorm.DB) error, opts ...*sql.TxOptions) error {
	err := db.Transaction(fc, opts...)
	if err != nil {
		// Check if it's already a gRPC status error
		if _, ok := status.FromError(err); ok {
			return err
		}
		return status.Error(codes.Internal, err.Error())
	}
	return nil
}

// HandleDBError converts database errors to appropriate gRPC status errors
func HandleDBError(err error, notFoundMsg string) error {
	if err == gorm.ErrRecordNotFound {
		return status.Error(codes.NotFound, notFoundMsg)
	}
	return status.Error(codes.Internal, err.Error())
}

// FindRecord is a generic function to find a record by ID
func FindRecord[T any](db *gorm.DB, id int64, notFoundMsg string) (*T, error) {
	var record T
	if err := db.Take(&record, id).Error; err != nil {
		return nil, HandleDBError(err, notFoundMsg)
	}
	return &record, nil
}

// GetIDFromHash fetches entity ID for a given hash from the database
func GetIDFromHash(db *gorm.DB, hash string, table string) (int64, error) {
	var id int64
	err := db.Table(table).
		Select("id").
		Where("hash = ?", hash).
		Take(&id).Error
	return id, err
}

// GetIDsFromHashes fetches entity IDs for given hashes from the database maintaining order
func GetIDsFromHashes(db *gorm.DB, hashes []string, table string) (map[string]int64, error) {
	var results []struct {
		ID   int64  `gorm:"column:id"`
		Hash string `gorm:"column:hash"`
	}

	err := db.Table(table).
		Select("id, hash").
		Where("hash IN ?", hashes).
		Find(&results).Error
	if err != nil {
		return nil, err
	}

	hashToID := make(map[string]int64)
	for _, result := range results {
		hashToID[result.Hash] = result.ID
	}

	return hashToID, nil
}
