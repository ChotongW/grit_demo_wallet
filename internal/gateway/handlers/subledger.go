package handlers

import (
	pbSub "github.com/ChotongW/grit_demo_wallet/pb/subledger"

	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

type SubLedgerHandler struct {
	logger    *logrus.Logger
	subClient pbSub.SubledgerServiceClient
}

func NewSubLedgerHandler(logger *logrus.Logger, subledgerConn *grpc.ClientConn) *SubLedgerHandler {
	return &SubLedgerHandler{
		logger:    logger,
		subClient: pbSub.NewSubledgerServiceClient(subledgerConn),
	}
}
