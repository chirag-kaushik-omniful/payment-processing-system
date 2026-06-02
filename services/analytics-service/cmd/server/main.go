package main

import (
	"context"
	"encoding/json"
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
	"github.com/omniful/payment-platform/shared/tracing"
	"github.com/omniful/payment-platform/services/analytics-service/internal/handler"
	"github.com/omniful/payment-platform/services/analytics-service/internal/service"
	"github.com/omniful/payment-platform/services/analytics-service/internal/usecase"
)

func main() {
	cfg := config.Load("analytics-service")
	cfg.HTTPPort = getEnv("HTTP_PORT", "8010")

	log, _ := logger.New(cfg.LogLevel)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	shutdownTrace, _ := tracing.Init(ctx, cfg.ServiceName, cfg.OTLPEndpoint)
	defer func() { _ = shutdownTrace(context.Background()) }()

	uc := usecase.NewAnalyticsUsecase()
	consumer := service.NewAnalyticsConsumer(uc, log)
	startConsumer(ctx, cfg, kafka.TopicPaymentCreated, consumer.HandlePaymentCreated)
	startConsumer(ctx, cfg, kafka.TopicPaymentCompleted, consumer.HandlePaymentCompleted)
	startConsumer(ctx, cfg, kafka.TopicPaymentFailed, consumer.HandlePaymentFailed)

	h := handler.NewAnalyticsHandler(uc)
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
	cancel()
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()
	_ = srv.Shutdown(shutdownCtx)
}

func startConsumer(ctx context.Context, cfg config.Base, topic string, handler func(context.Context, string, json.RawMessage) error) {
	c := kafka.NewConsumer(kafka.Config{
		Brokers: cfg.KafkaBrokers,
		GroupID: "analytics-group",
	}, topic, func(c context.Context, key string, payload json.RawMessage) error {
		return handler(c, key, payload)
	})
	go func() { _ = c.Run(ctx) }()
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
