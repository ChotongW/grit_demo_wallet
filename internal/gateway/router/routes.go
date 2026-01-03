package router

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/ChotongW/grit_demo_wallet/config/gateway"
	_ "github.com/ChotongW/grit_demo_wallet/docs" // Import for swagger docs
	"github.com/ChotongW/grit_demo_wallet/internal/gateway/handlers"
	"github.com/ChotongW/grit_demo_wallet/pkg/requestid"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func NewRouter(config *gateway.ServiceConfig, logger *logrus.Logger) {
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()

	r.Use(requestid.GinMiddleware())

	subledgerConn, err := grpc.NewClient(
		config.SubledgerService,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithUnaryInterceptor(requestid.UnaryClientInterceptor()),
		grpc.WithStreamInterceptor(requestid.StreamClientInterceptor()),
	)
	if err != nil {
		log.Fatalf("did not connect to subledger: %v", err)
	}
	defer subledgerConn.Close()

	accountConn, err := grpc.NewClient(
		config.AccountsService,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithUnaryInterceptor(requestid.UnaryClientInterceptor()),
		grpc.WithStreamInterceptor(requestid.StreamClientInterceptor()),
	)
	if err != nil {
		log.Fatalf("did not connect to accounts: %v", err)
	}
	defer accountConn.Close()

	subLedgerHandlers := handlers.NewSubLedgerHandler(logger, subledgerConn)
	accountsHandlers := handlers.NewAccountsHandler(logger, accountConn)

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	r.NoRoute(func(c *gin.Context) {
		c.JSON(
			400, gin.H{
				"error":   "not found",
				"message": fmt.Sprintf("the requested URL %s was not found on this server", c.Request.URL.Path),
			},
		)
	})

	api := r.Group("/api")
	apiV1 := api.Group("/v1")

	apiV1.GET("/health", HealthCheck)
	apiV1.POST("/accounts", accountsHandlers.CreateAccount)
	apiV1.GET("/accounts/:account_id", accountsHandlers.GetAccount)
	apiV1.GET("/accounts/:account_id/balance", accountsHandlers.GetBalance)
	apiV1.POST("/accounts/deposit", accountsHandlers.Deposit)
	apiV1.POST("/accounts/withdraw", accountsHandlers.Withdraw)
	apiV1.POST("/transfers", accountsHandlers.Transfer)
	apiV1.GET("/accounts/:account_id/transactions", accountsHandlers.GetTransactionHistory)
	apiV1.POST("/transaction", subLedgerHandlers.CreateTransaction)

	HttpServer := http.Server{
		Addr:              fmt.Sprintf(":%d", config.HttpPort),
		Handler:           r,
		ReadHeaderTimeout: config.ReadHeaderTimeout,
		ReadTimeout:       config.ReadTimeout,
		WriteTimeout:      config.WriteTimeout,
		IdleTimeout:       config.IdleTimeout,
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	go func() {
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, os.Interrupt)
		<-quit

		logger.Infof("gracefully shutting down...")

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := HttpServer.Shutdown(ctx); err != nil {
			logger.Errorf("server forced to shutdown: %+v", err)
		}
		logger.Infof("server exited properly")

	}()

	logger.Infof("serving HTTP API at http://127.0.0.1:%d", config.HttpPort)
	if err := HttpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.Errorf("HTTP server listen and serves failed: %+v", err)
	}
}

// HealthCheck godoc
//
//	@Summary		Health check endpoint
//	@Description	Check if the service is running
//	@Tags			Health
//	@Produce		json
//	@Success		200	{object}	map[string]string
//	@Router			/health [get]
func HealthCheck(c *gin.Context) {
	c.JSON(200, gin.H{
		"service": "e-contract",
	})
}
