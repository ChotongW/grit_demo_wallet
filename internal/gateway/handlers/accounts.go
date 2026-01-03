package handlers

import (
	"errors"
	"strconv"

	gwerrors "github.com/ChotongW/grit_demo_wallet/internal/gateway/errors"
	pb "github.com/ChotongW/grit_demo_wallet/pb/accounts"
	"github.com/ChotongW/grit_demo_wallet/pkg/requestid"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

type AccountsHandler struct {
	logger *logrus.Logger
	client pb.AccountsServiceClient
}

func NewAccountsHandler(logger *logrus.Logger, conn *grpc.ClientConn) *AccountsHandler {
	return &AccountsHandler{
		logger: logger,
		client: pb.NewAccountsServiceClient(conn),
	}
}

func (h *AccountsHandler) loggerWithRequestID(c *gin.Context) *logrus.Entry {
	reqID := requestid.FromContext(c.Request.Context())
	return h.logger.WithField("request_id", reqID)
}

// CreateAccount godoc
//
//	@Summary		Create new account
//	@Description	Create a new user account with optional initial balance and referral
//	@Tags			Accounts
//	@Accept			json
//	@Produce		json
//	@Param			request	body		object{email=string,initial_balance=string,referrer_account_id=string}	true	"Account creation request"
//	@Success		200		{object}	object{success=bool,account_id=string,message=string,account=object}
//	@Failure		400		{object}	object{error=string}
//	@Failure		500		{object}	object{error=string}
//	@Failure		500		{object}	object{error=string}
//	@Security		ApiKeyAuth
//	@Router			/accounts [post]
func (h *AccountsHandler) CreateAccount(c *gin.Context) {
	logger := h.loggerWithRequestID(c)

	var req struct {
		Email             string `json:"email" binding:"required,email" example:"test@example.com"`
		InitialBalance    string `json:"initial_balance" example:"100"`
		ReferrerAccountID string `json:"referrer_account_id"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		gwerrors.HandleBindingError(c, err)
		return
	}

	resp, err := h.client.CreateAccount(c.Request.Context(), &pb.CreateAccountRequest{
		Email:             req.Email,
		InitialBalance:    req.InitialBalance,
		ReferrerAccountId: req.ReferrerAccountID,
	})

	if err != nil {
		logger.Errorf("failed to create account: %v", err)
		gwerrors.HandleServiceError(c, err)
		return
	}

	logger.Infof("created account: %s", resp.AccountId)
	c.JSON(200, gin.H{
		"success":    resp.Success,
		"account_id": resp.AccountId,
		"message":    resp.Message,
		"account":    resp.Account,
	})
}

// GetAccount godoc
//
//	@Summary		Get account details
//	@Description	Retrieve account information including current balance
//	@Tags			Accounts
//	@Produce		json
//	@Param			account_id	path		string	true	"Account ID"
//	@Success		200			{object}	object{account=object}
//	@Failure		404			{object}	object{error=string}
//	@Failure		404			{object}	object{error=string}
//	@Security		ApiKeyAuth
//	@Router			/accounts/{account_id} [get]
func (h *AccountsHandler) GetAccount(c *gin.Context) {
	logger := h.loggerWithRequestID(c)
	accountID := c.Param("account_id")

	if accountID == "" {
		logger.Errorf("account_id is required")
		gwerrors.HandleServiceError(c, errors.New("account_id is required"))
		return
	}
	resp, err := h.client.GetAccount(c.Request.Context(), &pb.GetAccountRequest{
		AccountId: accountID,
	})

	if err != nil {
		logger.Errorf("failed to get account: %v", err)
		gwerrors.HandleServiceError(c, err)
		return
	}

	logger.Infof("retrieved account: %s", accountID)
	c.JSON(200, gin.H{
		"account": resp.Account,
	})
}

// GetBalance godoc
//
//	@Summary		Get account balance
//	@Description	Retrieve current balance for an account
//	@Tags			Accounts
//	@Produce		json
//	@Param			account_id	path		string	true	"Account ID"
//	@Success		200			{object}	object{account_id=string,balance=string,currency=string}
//	@Failure		404			{object}	object{error=string}
//	@Failure		404			{object}	object{error=string}
//	@Security		ApiKeyAuth
//	@Router			/accounts/{account_id}/balance [get]
func (h *AccountsHandler) GetBalance(c *gin.Context) {
	logger := h.loggerWithRequestID(c)
	accountID := c.Param("account_id")

	if accountID == "" {
		logger.Errorf("account_id is required")
		gwerrors.HandleServiceError(c, errors.New("account_id is required"))
		return
	}
	resp, err := h.client.GetBalance(c.Request.Context(), &pb.GetBalanceRequest{
		AccountId: accountID,
	})

	if err != nil {
		logger.Errorf("failed to get balance: %v", err)
		gwerrors.HandleServiceError(c, err)
		return
	}

	logger.Infof("retrieved balance for account: %s", accountID)
	c.JSON(200, gin.H{
		"account_id": resp.AccountId,
		"balance":    resp.Balance,
		"currency":   resp.Currency,
	})
}

// Deposit godoc
//
//	@Summary		Deposit funds
//	@Description	Deposit funds to an account from PSP
//	@Tags			Wallet
//	@Accept			json
//	@Produce		json
//	@Param			request		body		object{account_id=string,amount=string,description=string}	true	"Deposit request"
//	@Success		200			{object}	object{success=bool,transaction_id=string,new_balance=string,message=string}
//	@Failure		400			{object}	object{error=string}
//	@Failure		500			{object}	object{error=string}
//	@Failure		500			{object}	object{error=string}
//	@Security		ApiKeyAuth
//	@Router			/accounts/deposit [post]
func (h *AccountsHandler) Deposit(c *gin.Context) {
	logger := h.loggerWithRequestID(c)

	var req struct {
		AccountID   string `json:"account_id" binding:"required" example:"1001"`
		Amount      string `json:"amount" binding:"required" example:"500.00"`
		Description string `json:"description"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		gwerrors.HandleBindingError(c, err)
		return
	}

	resp, err := h.client.Deposit(c.Request.Context(), &pb.DepositRequest{
		AccountId:   req.AccountID,
		Amount:      req.Amount,
		Description: req.Description,
	})

	if err != nil {
		logger.Errorf("failed to deposit: %v", err)
		gwerrors.HandleServiceError(c, err)
		return
	}

	logger.Infof("deposit successful: account=%s, amount=%s", req.AccountID, req.Amount)
	c.JSON(200, gin.H{
		"success":        resp.Success,
		"transaction_id": resp.TransactionId,
		"new_balance":    resp.NewBalance,
		"message":        resp.Message,
	})
}

// Withdraw godoc
//
//	@Summary		Withdraw funds
//	@Description	Withdraw funds from an account to PSP
//	@Tags			Wallet
//	@Accept			json
//	@Produce		json
//	@Param			request		body		object{account_id=string,amount=string,description=string}	true	"Withdrawal request"
//	@Success		200			{object}	object{success=bool,transaction_id=string,new_balance=string,message=string}
//	@Failure		400			{object}	object{error=string}
//	@Failure		500			{object}	object{error=string}
//	@Failure		500			{object}	object{error=string}
//	@Security		ApiKeyAuth
//	@Router			/accounts/withdraw [post]
func (h *AccountsHandler) Withdraw(c *gin.Context) {
	logger := h.loggerWithRequestID(c)

	var req struct {
		AccountID   string `json:"account_id" binding:"required" example:"1001"`
		Amount      string `json:"amount" binding:"required" example:"200.00"`
		Description string `json:"description"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		gwerrors.HandleBindingError(c, err)
		return
	}

	resp, err := h.client.Withdraw(c.Request.Context(), &pb.WithdrawRequest{
		AccountId:   req.AccountID,
		Amount:      req.Amount,
		Description: req.Description,
	})

	if err != nil {
		logger.Errorf("failed to withdraw: %v", err)
		gwerrors.HandleServiceError(c, err)
		return
	}

	logger.Infof("withdrawal successful: account=%s, amount=%s", req.AccountID, req.Amount)
	c.JSON(200, gin.H{
		"success":        resp.Success,
		"transaction_id": resp.TransactionId,
		"new_balance":    resp.NewBalance,
		"message":        resp.Message,
	})
}

// Transfer godoc
//
//	@Summary		Transfer funds
//	@Description	Transfer funds from one account to another
//	@Tags			Wallet
//	@Accept			json
//	@Produce		json
//	@Param			request	body		object{from_account_id=string,to_account_id=string,amount=string,description=string}	true	"Transfer request"
//	@Success		200		{object}	object{success=bool,transaction_id=string,new_balance=string,message=string}
//	@Failure		400		{object}	object{error=string}
//	@Failure		500		{object}	object{error=string}
//	@Failure		500		{object}	object{error=string}
//	@Security		ApiKeyAuth
//	@Router			/transfers [post]
func (h *AccountsHandler) Transfer(c *gin.Context) {
	logger := h.loggerWithRequestID(c)

	var req struct {
		FromAccountID string `json:"from_account_id" binding:"required"`
		ToAccountID   string `json:"to_account_id" binding:"required"`
		Amount        string `json:"amount" binding:"required"`
		Description   string `json:"description"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		gwerrors.HandleBindingError(c, err)
		return
	}

	resp, err := h.client.Transfer(c.Request.Context(), &pb.TransferRequest{
		FromAccountId: req.FromAccountID,
		ToAccountId:   req.ToAccountID,
		Amount:        req.Amount,
		Description:   req.Description,
	})

	if err != nil {
		logger.Errorf("failed to transfer: %v", err)
		gwerrors.HandleServiceError(c, err)
		return
	}

	logger.Infof("transfer successful: from=%s, to=%s, amount=%s", req.FromAccountID, req.ToAccountID, req.Amount)
	c.JSON(200, gin.H{
		"success":        resp.Success,
		"transaction_id": resp.TransactionId,
		"new_balance":    resp.NewBalance,
		"message":        resp.Message,
	})
}

// GetTransactionHistory godoc
//
//	@Summary		Get transaction history
//	@Description	Retrieve paginated transaction history for an account
//	@Tags			Accounts
//	@Produce		json
//	@Param			account_id	path		string	true	"Account ID"
//	@Param			page		query		int		false	"Page number (default: 1)"
//	@Param			page_size	query		int		false	"Page size (default: 20, max: 100)"
//	@Success		200			{object}	object{transactions=array,total_count=int,page=int,page_size=int,total_pages=int}
//	@Failure		500			{object}	object{error=string}
//	@Failure		500			{object}	object{error=string}
//	@Security		ApiKeyAuth
//	@Router			/accounts/{account_id}/transactions [get]
func (h *AccountsHandler) GetTransactionHistory(c *gin.Context) {
	logger := h.loggerWithRequestID(c)
	accountID := c.Param("account_id")

	page := 1
	pageSize := 20

	if p := c.Query("page"); p != "" {
		if parsed, err := strconv.Atoi(p); err == nil {
			page = parsed
		}
	}

	if ps := c.Query("page_size"); ps != "" {
		if parsed, err := strconv.Atoi(ps); err == nil && parsed > 0 && parsed <= 100 {
			pageSize = parsed
		}
	}

	resp, err := h.client.GetTransactionHistory(c.Request.Context(), &pb.GetTransactionHistoryRequest{
		AccountId: accountID,
		Page:      int32(page),
		PageSize:  int32(pageSize),
	})

	if err != nil {
		logger.Errorf("failed to get transaction history: %v", err)
		gwerrors.HandleServiceError(c, err)
		return
	}

	logger.Infof("retrieved transaction history: account=%s", accountID)
	c.JSON(200, gin.H{
		"transactions": resp.Transactions,
		"total_count":  resp.TotalCount,
		"page":         resp.Page,
		"page_size":    resp.PageSize,
		"total_pages":  resp.TotalPages,
	})
}
