package handlers

import (
	gwerrors "github.com/ChotongW/grit_demo_wallet/internal/gateway/errors"
	pbSub "github.com/ChotongW/grit_demo_wallet/pb/subledger"
	"github.com/ChotongW/grit_demo_wallet/pkg/requestid"

	"github.com/gin-gonic/gin"
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

// CreateTransaction godoc
//
//	@Summary		Create transaction
//	@Description	Create a new transaction in the subledger
//	@Tags			Subledger
//	@Accept			json
//	@Produce		json
//	@Param			request	body		object{reference_id=string,description=string,entries=array}	true	"Transaction request"
//	@Success		200		{object}	object{transaction_id=string,status=string}
//	@Failure		400		{object}	object{error=string}
//	@Failure		500		{object}	object{error=string}
//	@Router			/transaction [post]
func (h *SubLedgerHandler) CreateTransaction(c *gin.Context) {
	reqID := requestid.FromContext(c.Request.Context())
	logger := h.logger.WithField("request_id", reqID)

	var req pbSub.CreateTransactionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Errorf("invalid request body: %v", err)
		gwerrors.HandleBindingError(c, err)
		return
	}

	resp, err := h.subClient.CreateTransaction(c.Request.Context(), &req)
	if err != nil {
		logger.Errorf("failed to create transaction: %v", err)
		gwerrors.HandleServiceError(c, err)
		return
	}

	logger.Infof("created transaction: %s", req.ReferenceId)
	c.JSON(200, resp)
}
