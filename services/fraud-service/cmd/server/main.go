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
	"github.com/omniful/payment-platform/shared/redis"
	"github.com/omniful/payment-platform/shared/tracing"
	"github.com/omniful/payment-platform/services/fraud-service/internal/handler"
	"github.com/omniful/payment-platform/services/fraud-service/internal/service"
	"github.com/omniful/payment-platform/services/fraud-service/internal/usecase"
)

func main() {
	cfg := config.Load("fraud-service")
	cfg.HTTPPort = getEnv("HTTP_PORT", "8006")

	log, _ := logger.New(cfg.LogLevel)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	shutdownTrace, _ := tracing.Init(ctx, cfg.ServiceName, cfg.OTLPEndpoint)
	defer func() { _ = shutdownTrace(context.Background()) }()

	redisClient, _ := redis.NewClient(cfg.RedisURL)
	producer := kafka.NewProducer(cfg.KafkaBrokers)
	defer producer.Close()

	uc := usecase.NewFraudUsecase(redisClient)
	consumer := service.NewFraudConsumer(uc, producer, log)
	startConsumer(ctx, cfg, kafka.TopicFraudValidate, consumer.HandleFraudValidate)

	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(middleware.CorrelationID())
	r.Use(middleware.ServiceName(cfg.ServiceName))
	r.Use(middleware.RequestLogger(log))
	fh := handler.NewFraudHandler(uc)
	fh.RegisterRoutes(r)
	r.GET("/health", func(c *gin.Context) { c.JSON(200, gin.H{"status": "ok"}) })
	r.GET("/metrics", gin.WrapH(metrics.Handler()))

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
		GroupID: "fraud-group",
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
