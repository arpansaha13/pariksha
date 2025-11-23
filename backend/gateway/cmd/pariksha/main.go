package main

import (
	"fmt"
	"net/http"

	"go.uber.org/zap"

	"pariksha/common/pkg/logging"
	"pariksha/gateway/internal/config/env"
	"pariksha/gateway/internal/interservice"
	"pariksha/gateway/internal/router"
)

func main() {
	// Initialize logger
	baseLogger := logging.InitLogger(env.GO_ENV)
	defer baseLogger.Sync()

	r := router.SetupRouter()

	port := env.API_PORT
	baseLogger.Info("Server starting", zap.String("port", port), zap.String("address", fmt.Sprintf("localhost:%s", port)))

	defer interservice.CloseAuthConn()

	addr := ":" + port
	if err := http.ListenAndServe(addr, r); err != nil {
		baseLogger.Fatal("Server failed to start", zap.Error(err))
	}
}
