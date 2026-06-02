package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/omniful/payment-platform/shared/config"
	"github.com/omniful/payment-platform/shared/kafka"
	"github.com/omniful/payment-platform/shared/logger"
	"github.com/omniful/payment-platform/shared/metrics"
	"github.com/omniful/payment-platform/shared/middleware"
	"github.com/omniful/payment-platform/shared/redis"
	"github.com/omniful/payment-platform/shared/tracing"
	"github.com/omniful/payment-platform/services/webhook-service/internal/handler"
	"github.com/omniful/payment-platform/services/webhook-service/internal/service"
	"github.com/omniful/payment-platform/services/webhook-service/internal/usecase"
)

func main() {
	cfg := config.Load("webhook-service")
	cfg.HTTPPort = getEnv("HTTP_PORT", "8008")

	log, _ := logger.New(cfg.LogLevel)
	ctx := context.Background()

	shutdownTrace, _ := tracing.Init(ctx, cfg.ServiceName, cfg.OTLPEndpoint)
	defer func() { _ = shutdownTrace(context.Background()) }()

	redisClient, _ := redis.NewClient(cfg.RedisURL)
	idempotency := redis.NewIdempotencyStore(redisClient)
	producer := kafka.NewProducer(cfg.KafkaBrokers)
	defer producer.Close()
	publisher := service.NewKafkaPublisher(producer)
	uc := usecase.NewWebhookUsecase(idempotency, publisher)
	h := handler.NewWebhookHandler(uc)

	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(middleware.CorrelationID())
	r.Use(middleware.ServiceName(cfg.ServiceName))
	r.Use(middleware.RequestLogger(log))
	r.GET("/metrics", gin.WrapH(metrics.Handler()))
	h.RegisterRoutes(r)

	srv := &http.Server{Addr: ":" + cfg.HTTPPort, Handler: r}
	go func() { _ = srv.ListenAndServe() }()

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
