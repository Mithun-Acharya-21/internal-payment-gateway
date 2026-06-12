package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Mithun-Acharya-21/internal-payment-gateway/internal/config"
	"github.com/Mithun-Acharya-21/internal-payment-gateway/internal/database"
	"github.com/Mithun-Acharya-21/internal-payment-gateway/internal/handler"
	"github.com/Mithun-Acharya-21/internal-payment-gateway/internal/middleware"
	"github.com/Mithun-Acharya-21/internal-payment-gateway/internal/repository"
	"github.com/Mithun-Acharya-21/internal-payment-gateway/internal/service"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load config: %v\n", err)
		os.Exit(1)
	}

	logger, err := buildLogger(cfg.Env)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to init logger: %v\n", err)
		os.Exit(1)
	}
	defer logger.Sync() 

	db, err := database.NewPostgres(cfg.DatabaseURL)
	if err != nil {
		logger.Fatal("failed to connect to database", zap.Error(err))
	}
	defer db.Close()

	if err := database.Migrate(db); err != nil {
		logger.Fatal("failed to run migrations", zap.Error(err))
	}

	txRepo := repository.NewTransactionRepository(db)
	walletRepo := repository.NewWalletRepository(db)

	paymentSvc := service.NewPaymentService(txRepo, walletRepo, logger)
	walletSvc := service.NewWalletService(walletRepo, logger)

	paymentHandler := handler.NewPaymentHandler(paymentSvc, logger)
	walletHandler := handler.NewWalletHandler(walletSvc, logger)
	healthHandler := handler.NewHealthHandler(db)

	if cfg.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()
	router.Use(
		middleware.RequestID(),
		middleware.Logger(logger),
		middleware.Recovery(logger),
		middleware.CORS(cfg.AllowedOrigins),
		middleware.RateLimiter(cfg.RateLimitRPS),
	)

	router.GET("/healthz", healthHandler.Health)
	router.GET("/readyz", healthHandler.Ready)

	v1 := router.Group("/api/v1")
	v1.Use(middleware.JWTAuth(cfg.JWTSecret))
	{
		payments := v1.Group("/payments")
		{
			payments.POST("", paymentHandler.InitiatePayment)
			payments.GET("/:id", paymentHandler.GetPayment)
			payments.GET("", paymentHandler.ListPayments)
			payments.POST("/:id/refund", paymentHandler.RefundPayment)
		}

		wallets := v1.Group("/wallets")
		{
			wallets.POST("", walletHandler.CreateWallet)
			wallets.GET("/:id", walletHandler.GetWallet)
			wallets.GET("/:id/balance", walletHandler.GetBalance)
			wallets.POST("/:id/topup", walletHandler.TopUp)
		}
	}

	router.StaticFile("/", "./web/dashboard.html")

	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		logger.Info("server listening", zap.String("addr", srv.Addr))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("server error", zap.Error(err))
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("shutting down server gracefully...")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("forced shutdown", zap.Error(err))
	}
	logger.Info("server exited")
}

func buildLogger(env string) (*zap.Logger, error) {
	if env == "production" {
		return zap.NewProduction()
	}
	return zap.NewDevelopment()
}

