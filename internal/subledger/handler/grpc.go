package handler

import (
	"context"

	"github.com/ChotongW/grit_demo_wallet/internal/subledger/repository"
	"github.com/ChotongW/grit_demo_wallet/internal/subledger/service"
	pb "github.com/ChotongW/grit_demo_wallet/pb/subledger"
	"github.com/ChotongW/grit_demo_wallet/pkg/requestid"

	"github.com/shopspring/decimal"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type GRPCHandler struct {
	pb.UnimplementedSubledgerServiceServer
	service *service.Service
	logger  logrus.FieldLogger
}

func NewGRPCHandler(svc *service.Service, logger *logrus.Logger) *GRPCHandler {
	newLogger := logger.WithFields(
		logrus.Fields{
			"package": "handler",
		},
	)
	return &GRPCHandler{
		service: svc,
		logger:  newLogger,
	}
}

func (h *GRPCHandler) loggerWithRequestID(ctx context.Context) logrus.FieldLogger {
	reqID := requestid.FromContext(ctx)
	if reqID != "" {
		return h.logger.WithField("request_id", reqID)
	}
	return h.logger
}

func (h *GRPCHandler) CreateTransaction(ctx context.Context, req *pb.CreateTransactionRequest) (*pb.CreateTransactionResponse, error) {
	logger := h.loggerWithRequestID(ctx)

	entries := make([]repository.TransactionEntry, 0, len(req.Entries))
	for _, e := range req.Entries {
		amount, err := decimal.NewFromString(e.Amount)
		if err != nil {
			logger.Errorf("invalid amount: %v", err)
			return nil, status.Errorf(codes.InvalidArgument, "invalid amount: %v", err)
		}

		entries = append(entries, repository.TransactionEntry{
			AccountID: e.AccountId,
			Amount:    amount,
			Direction: e.Direction,
		})
	}

	logger.Debugf("request body: %+v", req)
	err := h.service.CreateTransaction(ctx, req.ReferenceId, req.Description, entries)
	if err != nil {
		logger.Errorf("failed to create transaction: %v", err)
		return nil, status.Errorf(codes.Internal, "failed to create transaction: %v", err)
	}

	logger.Infof("created transaction: %s", req.ReferenceId)
	return &pb.CreateTransactionResponse{
		Success:       true,
		TransactionId: req.ReferenceId,
	}, nil
}

func (h *GRPCHandler) GetBalance(ctx context.Context, req *pb.GetBalanceRequest) (*pb.GetBalanceResponse, error) {
	logger := h.loggerWithRequestID(ctx)

	balance, err := h.service.GetBalance(ctx, req.AccountId)
	if err != nil {
		logger.Errorf("failed to get balance: %v", err)
		return nil, status.Errorf(codes.Internal, "failed to get balance: %v", err)
	}

	logger.Infof("retrieved balance for account %s", req.AccountId)
	return &pb.GetBalanceResponse{
		AccountId: req.AccountId,
		Currency:  "USD",
		Amount:    balance.String(),
	}, nil
}
