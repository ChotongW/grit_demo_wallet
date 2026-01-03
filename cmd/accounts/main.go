package main

import (
	"fmt"
	"log"
	"net"
	"os"

	"github.com/ChotongW/grit_demo_wallet/config/accounts"
	"github.com/ChotongW/grit_demo_wallet/internal/accounts/handler"
	"github.com/ChotongW/grit_demo_wallet/internal/accounts/repository"
	"github.com/ChotongW/grit_demo_wallet/internal/accounts/service"
	pb "github.com/ChotongW/grit_demo_wallet/pb/accounts"
	pbSub "github.com/ChotongW/grit_demo_wallet/pb/subledger"
	"github.com/ChotongW/grit_demo_wallet/pkg/database"
	"github.com/ChotongW/grit_demo_wallet/pkg/logger"
	"github.com/ChotongW/grit_demo_wallet/pkg/requestid"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		configPath = "./config/accounts"
	}

	cfg, err := accounts.LoadConfig(configPath)
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	logger := logger.NewLogger(logger.LogConfig{
		Level:          cfg.LogLevel,
		FormatJson:     cfg.LogFormatJson,
		Color:          cfg.LogColor,
		LogLineDetails: cfg.LogLineDetails,
	})

	logger.Info("starting accounts service...")

	db, err := database.New(&cfg.DbConfig, logger)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer db.Close()

	subledgerAddr := os.Getenv("SUBLEDGER_RPC_ADDR")
	if subledgerAddr == "" {
		subledgerAddr = "localhost:50051"
	}
	conn, err := grpc.Dial(
		subledgerAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithUnaryInterceptor(requestid.UnaryClientInterceptor()),
		grpc.WithStreamInterceptor(requestid.StreamClientInterceptor()),
	)
	if err != nil {
		log.Fatalf("failed to connect to subledger service: %v", err)
	}
	defer conn.Close()

	subledgerClient := pbSub.NewSubledgerServiceClient(conn)
	logger.Infof("connected to subledger service at %s", subledgerAddr)

	repo := repository.NewRepository(db.Pool, logger)
	svc := service.NewService(repo, subledgerClient, logger)
	grpcHandler := handler.NewGRPCHandler(svc, logger)
	grpcPort := os.Getenv("GRPC_PORT")
	if grpcPort == "" {
		grpcPort = "50052"
	}

	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", grpcPort))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(requestid.UnaryServerInterceptor()),
		grpc.StreamInterceptor(requestid.StreamServerInterceptor()),
	)
	pb.RegisterAccountsServiceServer(grpcServer, grpcHandler)

	logger.Infof("accounts gRPC server listening on port %s", grpcPort)

	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
