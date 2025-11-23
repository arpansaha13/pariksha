package logging

import (
	"context"

	"go.uber.org/zap"
)

type contextKey string

const (
	loggerKey    contextKey = "logger"
	requestIDKey contextKey = "request_id"
)

// storeLoggerInContext stores a logger instance in the given context and returns the updated context.
func storeLoggerInContext(ctx context.Context, logger *zap.Logger) context.Context {
	return context.WithValue(ctx, loggerKey, logger)
}

// GetLoggerFromContext retrieves a logger instance from the context.
// It returns the logger and a boolean indicating whether the logger was found.
func GetLoggerFromContext(ctx context.Context) (*zap.Logger, bool) {
	logger, ok := ctx.Value(loggerKey).(*zap.Logger)
	return logger, ok
}

// storeRequestIDInContext stores a request ID in the given context and returns the updated context.
func storeRequestIDInContext(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, requestIDKey, requestID)
}

// GetRequestIDFromContext retrieves a request ID from the context.
// It returns the request ID and a boolean indicating whether the request ID was found.
func GetRequestIDFromContext(ctx context.Context) (string, bool) {
	requestID, ok := ctx.Value(requestIDKey).(string)
	return requestID, ok
}

// GetOrCreateLoggerWithRequestID returns an enhanced logger with the request_id field set,
// and an updated context containing both the logger and request ID.
func GetOrCreateLoggerWithRequestID(ctx context.Context, logger *zap.Logger, requestID string) (*zap.Logger, context.Context) {
	if logger == nil {
		logger = GetLogger()
	}

	enhancedLogger := logger.With(zap.String("request_id", requestID))
	updatedCtx := storeLoggerInContext(ctx, enhancedLogger)
	updatedCtx = storeRequestIDInContext(updatedCtx, requestID)

	return enhancedLogger, updatedCtx
}
