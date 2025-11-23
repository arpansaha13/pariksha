package logging

import (
	"context"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// LoggingInterceptorConfig holds configuration for the gRPC logging interceptor.
type LoggingInterceptorConfig struct {
	Logger              *zap.Logger
	IncludeRequestBody  bool
	IncludeResponseBody bool
	SkipMethods         map[string]struct{}
}

// NewLoggingInterceptor creates a gRPC unary server interceptor with default configuration.
// It logs all gRPC requests and responses with request tracing.
func NewLoggingInterceptor(baseLogger *zap.Logger) grpc.UnaryServerInterceptor {
	config := &LoggingInterceptorConfig{
		Logger:              baseLogger,
		IncludeRequestBody:  false,
		IncludeResponseBody: false,
		SkipMethods:         make(map[string]struct{}),
	}
	return NewLoggingInterceptorWithConfig(config)
}

// NewLoggingInterceptorWithConfig creates a gRPC unary server interceptor with custom configuration.
func NewLoggingInterceptorWithConfig(config *LoggingInterceptorConfig) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		if config.SkipMethods != nil {
			if _, skip := config.SkipMethods[info.FullMethod]; skip {
				return handler(ctx, req)
			}
		}

		// Extract request_id from incoming gRPC metadata. Do NOT generate a new one here;
		// the HTTP gateway must have injected the request_id in metadata.
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Errorf(codes.InvalidArgument, "missing request_id: no correlation metadata propagated from upstream")
		}
		vals := md.Get("request_id")
		if len(vals) == 0 || vals[0] == "" {
			return nil, status.Errorf(codes.InvalidArgument, "missing request_id: no correlation metadata propagated from upstream")
		}
		requestID := vals[0]

		logger, updatedCtx := GetOrCreateLoggerWithRequestID(ctx, config.Logger, requestID)

		logFields := []zap.Field{
			zap.String("full_method", info.FullMethod),
		}
		if config.IncludeRequestBody {
			logFields = append(logFields, zap.Any("request", req))
		}

		logger.Info("gRPC request started", logFields...)

		startTime := time.Now()
		resp, handlerErr := handler(updatedCtx, req)
		duration := time.Since(startTime).Seconds() * 1000

		statusCode := extractStatusCodeFromError(handlerErr)

		responseLogFields := []zap.Field{
			zap.String("full_method", info.FullMethod),
			zap.Float64("duration_ms", duration),
			zap.String("status_code", statusCode),
		}
		if config.IncludeResponseBody && resp != nil {
			responseLogFields = append(responseLogFields, zap.Any("response", resp))
		}
		if handlerErr != nil {
			responseLogFields = append(responseLogFields, zap.Error(handlerErr))
			logger.Error("gRPC request completed", responseLogFields...)
		} else {
			logger.Info("gRPC request completed", responseLogFields...)
		}

		return resp, handlerErr
	}
}

// ExtractStatusCodeFromError extracts the gRPC status code from an error.
// Returns the status code as a string, or "OK" if the error is nil.
func ExtractStatusCodeFromError(err error) string {
	return extractStatusCodeFromError(err)
}

func generateRequestID() (string, error) {
	v7, err := uuid.NewV7()
	if err != nil {
		return "", err
	}
	return v7.String(), nil
}

func extractStatusCodeFromError(err error) string {
	if err == nil {
		return "OK"
	}
	st, _ := status.FromError(err)
	return st.Code().String()
}
