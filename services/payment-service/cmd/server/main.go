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
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"

	"github.com/omniful/payment-platform/shared/config"
	"github.com/omniful/payment-platform/shared/kafka"
	"github.com/omniful/payment-platform/shared/logger"
	"github.com/omniful/payment-platform/shared/metrics"
	"github.com/omniful/payment-platform/shared/middleware"
	"github.com/omniful/payment-platform/shared/redis"
	"github.com/omniful/payment-platform/shared/tracing"
	"github.com/omniful/payment-platform/services/payment-service/internal/handler"
	"github.com/omniful/payment-platform/services/payment-service/internal/repository"
	"github.com/omniful/payment-platform/services/payment-service/internal/service"
	"github.com/omniful/payment-platform/services/payment-service/internal/usecase"
)

func main() {
	cfg := config.Load("payment-service")
	cfg.HTTPPort = getEnv("HTTP_PORT", "8002")

	log, _ := logger.New(cfg.LogLevel)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	shutdownTrace, _ := tracing.Init(ctx, cfg.ServiceName, cfg.OTLPEndpoint)
	defer func() { _ = shutdownTrace(context.Background()) }()

	pool, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatal("db connect failed", zap.Error(err))
	}
	defer pool.Close()

	redisClient, _ := redis.NewClient(cfg.RedisURL)
	idempotency := redis.NewIdempotencyStore(redisClient)
	producer := kafka.NewProducer(cfg.KafkaBrokers)
	defer producer.Close()

	repo := repository.NewPaymentRepository(pool)
	uc := usecase.NewPaymentUsecase(repo, idempotency)
	h := handler.NewPaymentHandler(uc)

	outboxWorker := service.NewOutboxWorker(repo, producer, log)
	go outboxWorker.Run(ctx)

	saga := service.NewSagaConsumer(repo, producer, log)
	startConsumer(ctx, cfg, kafka.TopicPaymentCreated, saga.HandlePaymentCreated)
	startConsumer(ctx, cfg, kafka.TopicFraudValidated, saga.HandleFraudValidated)
	startConsumer(ctx, cfg, kafka.TopicProviderSuccess, saga.HandleProviderSuccess)
	startConsumer(ctx, cfg, kafka.TopicProviderFailed, saga.HandleProviderFailed)
	startConsumer(ctx, cfg, kafka.TopicLedgerEntryCreated, saga.HandleLedgerWritten)
	startConsumer(ctx, cfg, kafka.TopicRefundRequested, saga.HandleRefundRequested)
	startConsumer(ctx, cfg, kafka.TopicRefundCompleted, saga.HandleRefundCompleted)
	startConsumer(ctx, cfg, kafka.TopicWebhookReceived, saga.HandleWebhookReceived)

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
	consumer := kafka.NewConsumer(kafka.Config{
		Brokers: cfg.KafkaBrokers,
		GroupID: "payment-group",
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
