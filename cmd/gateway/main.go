package main

import (
	"log"
	"os"

	"github.com/ChotongW/grit_demo_wallet/config/gateway"
	"github.com/ChotongW/grit_demo_wallet/internal/gateway/router"
	"github.com/ChotongW/grit_demo_wallet/pkg/logger"
)

// @title           Grit Demo Wallet API
// @version         1.0
// @description     API Gateway for the Demo Wallet
// @BasePath        /api/v1
// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name X-API-KEY

func main() {
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		configPath = "./config/gateway"
	}

	cfg, err := gateway.LoadConfig(configPath)
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}
	logger := logger.NewLogger(logger.LogConfig{
		Level:          cfg.LogLevel,
		FormatJson:     cfg.LogFormatJson,
		Color:          cfg.LogColor,
		LogLineDetails: cfg.LogLineDetails,
	})

	logger.Info("starting gateway service...")

	router.NewRouter(cfg, logger)
}
