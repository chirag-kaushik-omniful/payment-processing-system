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
	"go.uber.org/zap"

	"github.com/omniful/payment-platform/shared/config"
	"github.com/omniful/payment-platform/shared/kafka"
	"github.com/omniful/payment-platform/shared/logger"
	"github.com/omniful/payment-platform/shared/metrics"
	"github.com/omniful/payment-platform/shared/middleware"
	"github.com/omniful/payment-platform/shared/tracing"
	"github.com/omniful/payment-platform/services/provider-service/internal/provider/factory"
	"github.com/omniful/payment-platform/services/provider-service/internal/service"
	"github.com/omniful/payment-platform/services/provider-service/internal/usecase"
)

func main() {
	cfg := config.Load("provider-service")
	if cfg.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}
	cfg.HTTPPort = getEnv("HTTP_PORT", "8003")

	log, _ := logger.New(cfg.LogLevel)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	shutdownTrace, _ := tracing.Init(ctx, cfg.ServiceName, cfg.OTLPEndpoint)
	defer func() { _ = shutdownTrace(context.Background()) }()

	producer := kafka.NewProducer(cfg.KafkaBrokers)
	defer producer.Close()

	providerFactory := factory.New()
	uc := usecase.NewChargeUsecase(providerFactory)
	consumer := service.NewChargeConsumer(uc, producer, log)

	refundConsumer := service.NewRefundConsumer(providerFactory, producer, log)
	startConsumer(ctx, cfg, kafka.TopicProviderCharge, consumer.HandleProviderCharge)
	startConsumer(ctx, cfg, kafka.TopicProviderRefund, refundConsumer.HandleProviderRefund)

	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(middleware.CorrelationID())
	r.Use(middleware.ServiceName(cfg.ServiceName))
	r.Use(middleware.RequestLogger(log))
	r.GET("/metrics", gin.WrapH(metrics.Handler()))
	r.GET("/health", func(c *gin.Context) { c.JSON(200, gin.H{"status": "ok"}) })

	srv := &http.Server{Addr: ":" + cfg.HTTPPort, Handler: r}
	go func() {
		log.Info("provider-service starting", logger.FieldService(cfg.ServiceName))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("server failed", zap.Error(err))
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	cancel()
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()
	_ = srv.Shutdown(shutdownCtx)
}

func startConsumer(ctx context.Context, cfg config.Base, topic string, handler func(context.Context, string, json.RawMessage) error) {
	consumer := kafka.NewConsumer(kafka.Config{
		Brokers: cfg.KafkaBrokers,
		GroupID: "provider-group",
	}, topic, func(c context.Context, key string, payload json.RawMessage) error {
		return handler(c, key, payload)
	})
	go func() { _ = consumer.Run(ctx) }()
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
