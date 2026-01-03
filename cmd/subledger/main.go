package main

import (
	"fmt"
	"log"
	"net"
	"os"

	"github.com/ChotongW/grit_demo_wallet/config/subledger"
	"github.com/ChotongW/grit_demo_wallet/internal/subledger/handler"
	"github.com/ChotongW/grit_demo_wallet/internal/subledger/repository"
	"github.com/ChotongW/grit_demo_wallet/internal/subledger/service"
	pb "github.com/ChotongW/grit_demo_wallet/pb/subledger"
	"github.com/ChotongW/grit_demo_wallet/pkg/database"
	"github.com/ChotongW/grit_demo_wallet/pkg/logger"
	"github.com/ChotongW/grit_demo_wallet/pkg/requestid"

	"google.golang.org/grpc"
)

func main() {
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		configPath = "./config/subledger"
	}

	cfg, err := subledger.LoadConfig(configPath)
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	logger := logger.NewLogger(logger.LogConfig{
		Level:          cfg.LogLevel,
		FormatJson:     cfg.LogFormatJson,
		Color:          cfg.LogColor,
		LogLineDetails: cfg.LogLineDetails,
	})

	logger.Info("starting Subledger Service...")

	db, err := database.New(&cfg.DbConfig, logger)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer db.Close()

	repo := repository.NewRepository(db.Pool, logger)
	svc := service.NewService(repo, logger)
	grpcHandler := handler.NewGRPCHandler(svc, logger)
	grpcPort := os.Getenv("GRPC_PORT")
	if grpcPort == "" {
		grpcPort = "50051"
	}

	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", grpcPort))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(requestid.UnaryServerInterceptor()),
		grpc.StreamInterceptor(requestid.StreamServerInterceptor()),
	)
	pb.RegisterSubledgerServiceServer(grpcServer, grpcHandler)

	logger.Infof("subledger gRPC server listening on port %s", grpcPort)

	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
