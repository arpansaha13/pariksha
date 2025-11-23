package logging

import (
	"fmt"
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var globalLogger *zap.Logger

// InitLogger initializes a Zap logger with JSON encoding and configurable log level.
// The logLevel parameter accepts: debug, info, warn, error, fatal.
// If an invalid log level is provided, it defaults to info level with a warning to stderr.
func InitLogger(logLevel string) *zap.Logger {
	return InitLoggerWithOptions(logLevel)
}

// InitLoggerWithOptions initializes a Zap logger with JSON encoding, configurable log level,
// and additional Zap options for advanced configuration.
func InitLoggerWithOptions(logLevel string, opts ...zap.Option) *zap.Logger {
	level, err := zapcore.ParseLevel(logLevel)
	if err != nil {
		fmt.Fprintf(os.Stderr, "invalid log level %q, defaulting to info: %v\n", logLevel, err)
		level = zapcore.InfoLevel
	}

	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "timestamp",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		FunctionKey:    "function",
		MessageKey:     "message",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	encoder := zapcore.NewJSONEncoder(encoderConfig)
	core := zapcore.NewCore(encoder, zapcore.AddSync(os.Stdout), level)

	defaultOpts := []zap.Option{
		zap.AddCaller(),
		zap.AddStacktrace(zapcore.ErrorLevel),
	}
	defaultOpts = append(defaultOpts, opts...)

	return zap.New(core, defaultOpts...)
}

// GetLogger returns the global singleton logger instance.
// If the logger has not been initialized, it initializes one with the default info level.
func GetLogger() *zap.Logger {
	if globalLogger == nil {
		globalLogger = InitLogger("info")
	}
	return globalLogger
}
