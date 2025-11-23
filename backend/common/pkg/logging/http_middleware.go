package logging

import (
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"

	"go.uber.org/zap"
)

// ResponseWriterWrapper wraps an http.ResponseWriter to capture the status code and response body.
type ResponseWriterWrapper struct {
	ResponseWriter http.ResponseWriter
	StatusCode     int32
	Written        bool
}

// Header returns the header map of the underlying ResponseWriter.
func (w *ResponseWriterWrapper) Header() http.Header {
	return w.ResponseWriter.Header()
}

// Write writes data to the underlying ResponseWriter and captures the status code if not already written.
func (w *ResponseWriterWrapper) Write(data []byte) (int, error) {
	if !w.Written {
		w.StatusCode = 200
		w.Written = true
	}
	return w.ResponseWriter.Write(data)
}

// WriteHeader writes the status code to the underlying ResponseWriter.
func (w *ResponseWriterWrapper) WriteHeader(statusCode int) {
	if !w.Written {
		w.StatusCode = int32(statusCode)
		w.Written = true
	}
	w.ResponseWriter.WriteHeader(statusCode)
}

// NewHTTPLoggingMiddleware creates an HTTP middleware that logs requests and responses.
func NewHTTPLoggingMiddleware(baseLogger *zap.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requestID, err := generateRequestID()
			if err != nil {
				requestID = fmt.Sprintf("%d", time.Now().UnixNano())
			}

			logger, ctx := GetOrCreateLoggerWithRequestID(r.Context(), baseLogger, requestID)
			r = r.WithContext(ctx)

			clientIP := GetClientIP(r)
			startTime := time.Now()

			wrapped := &ResponseWriterWrapper{
				ResponseWriter: w,
				StatusCode:     200,
				Written:        false,
			}

			logger.Info("HTTP request started",
				zap.String("method", r.Method),
				zap.String("path", r.URL.Path),
				zap.String("client_ip", clientIP),
			)

			next.ServeHTTP(wrapped, r)

			duration := time.Since(startTime).Seconds() * 1000
			statusCode := wrapped.StatusCode

			logLevel := zap.InfoLevel
			if statusCode >= 400 && statusCode < 500 {
				logLevel = zap.WarnLevel
			} else if statusCode >= 500 {
				logLevel = zap.ErrorLevel
			}

			if ce := logger.Check(logLevel, "HTTP request completed"); ce != nil {
				ce.Write(
					zap.String("method", r.Method),
					zap.String("path", r.URL.Path),
					zap.Int32("status_code", statusCode),
					zap.Float64("duration_ms", duration),
				)
			}

			if wrapped.Written {
				wrapped.ResponseWriter.WriteHeader(int(statusCode))
			}
		})
	}
}

// GetClientIP extracts the client IP address from the HTTP request.
// It checks X-Forwarded-For first, then X-Real-IP, then RemoteAddr, with a fallback to "unknown".
func GetClientIP(r *http.Request) string {
	if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
		ips := strings.Split(forwarded, ",")
		if len(ips) > 0 {
			return strings.TrimSpace(ips[0])
		}
	}

	if realIP := r.Header.Get("X-Real-IP"); realIP != "" {
		return realIP
	}

	if remoteAddr := r.RemoteAddr; remoteAddr != "" {
		ip, _, err := net.SplitHostPort(remoteAddr)
		if err == nil {
			return ip
		}
		return remoteAddr
	}

	return "unknown"
}
