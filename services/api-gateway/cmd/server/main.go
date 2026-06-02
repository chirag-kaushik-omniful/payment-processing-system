package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/omniful/payment-platform/shared/auth"
	"github.com/omniful/payment-platform/shared/config"
	"github.com/omniful/payment-platform/shared/logger"
	"github.com/omniful/payment-platform/shared/metrics"
	"github.com/omniful/payment-platform/shared/middleware"
	"github.com/omniful/payment-platform/shared/redis"
	"github.com/omniful/payment-platform/shared/tracing"
	"github.com/omniful/payment-platform/services/api-gateway/internal/handler"
)

func main() {
	cfg := config.Load("api-gateway")
	cfg.HTTPPort = getEnv("HTTP_PORT", "8000")

	log, _ := logger.New(cfg.LogLevel)
	ctx := context.Background()
	shutdown, _ := tracing.Init(ctx, cfg.ServiceName, cfg.OTLPEndpoint)
	defer func() { _ = shutdown(context.Background()) }()

	redisClient, err := redis.NewClient(cfg.RedisURL)
	if err != nil {
		log.Warn("redis unavailable, rate limiting disabled", zap.Error(err))
	}

	jwtMgr := auth.NewJWTManager(cfg.JWTSecret, cfg.JWTExpiry, cfg.RefreshExpiry)
	h := handler.NewGatewayHandler(jwtMgr, redisClient)

	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(middleware.CORS())
	r.Use(middleware.CorrelationID())
	r.Use(middleware.ServiceName(cfg.ServiceName))
	r.Use(middleware.RequestLogger(log))
	r.GET("/metrics", gin.WrapH(metrics.Handler()))
	h.RegisterRoutes(r)

	srv := &http.Server{Addr: ":" + cfg.HTTPPort, Handler: r}
	go func() {
		log.Info("api-gateway starting")
		_ = srv.ListenAndServe()
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
