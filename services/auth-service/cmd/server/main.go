package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"

	"github.com/omniful/payment-platform/shared/auth"
	"github.com/omniful/payment-platform/shared/config"
	"github.com/omniful/payment-platform/shared/logger"
	"github.com/omniful/payment-platform/shared/metrics"
	"github.com/omniful/payment-platform/shared/middleware"
	"github.com/omniful/payment-platform/shared/tracing"
	"github.com/omniful/payment-platform/services/auth-service/internal/handler"
	"github.com/omniful/payment-platform/services/auth-service/internal/repository"
	"github.com/omniful/payment-platform/services/auth-service/internal/usecase"
)

func main() {
	cfg := config.Load("auth-service")
	if cfg.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}
	cfg.HTTPPort = getEnv("HTTP_PORT", "8001")

	log, _ := logger.New(cfg.LogLevel)
	ctx := context.Background()
	shutdown, _ := tracing.Init(ctx, cfg.ServiceName, cfg.OTLPEndpoint)
	defer func() { _ = shutdown(context.Background()) }()

	pool, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatal("db connect failed", zap.Error(err))
	}
	defer pool.Close()

	jwtMgr := auth.NewJWTManager(cfg.JWTSecret, cfg.JWTExpiry, cfg.RefreshExpiry)
	repo := repository.NewUserRepository(pool)
	uc := usecase.NewAuthUsecase(repo, jwtMgr)
	h := handler.NewAuthHandler(uc)

	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(middleware.CorrelationID())
	r.Use(middleware.ServiceName(cfg.ServiceName))
	r.Use(middleware.RequestLogger(log))
	r.GET("/metrics", gin.WrapH(metrics.Handler()))
	h.RegisterRoutes(r)

	srv := &http.Server{Addr: ":" + cfg.HTTPPort, Handler: r}
	go func() {
		log.Info("auth-service starting", logger.FieldService(cfg.ServiceName))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("server failed", zap.Error(err))
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = srv.Shutdown(shutdownCtx)
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
