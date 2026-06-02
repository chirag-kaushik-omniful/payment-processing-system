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
	"github.com/omniful/payment-platform/shared/tracing"
	"github.com/omniful/payment-platform/services/ledger-service/internal/handler"
	"github.com/omniful/payment-platform/services/ledger-service/internal/repository"
	"github.com/omniful/payment-platform/services/ledger-service/internal/service"
	"github.com/omniful/payment-platform/services/ledger-service/internal/usecase"
)

func main() {
	cfg := config.Load("ledger-service")
	if cfg.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}
	cfg.HTTPPort = getEnv("HTTP_PORT", "8005")

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

	repo := repository.NewLedgerRepository(pool)
	uc := usecase.NewLedgerUsecase(repo)
	ledgerConsumer := service.NewLedgerConsumer(uc, log)
	compConsumer := service.NewCompensationConsumer(uc, log)

	startConsumer(ctx, cfg, kafka.TopicLedgerEntryCreated, ledgerConsumer.HandleLedgerEntryCreated)
	startConsumer(ctx, cfg, kafka.TopicLedgerCompensation, compConsumer.HandleCompensation)

	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(middleware.CorrelationID())
	r.Use(middleware.ServiceName(cfg.ServiceName))
	r.Use(middleware.RequestLogger(log))
	r.GET("/metrics", gin.WrapH(metrics.Handler()))
	h := handler.NewLedgerHandler(uc)
	h.RegisterRoutes(r)
	r.GET("/health", func(c *gin.Context) { c.JSON(200, gin.H{"status": "ok"}) })

	srv := &http.Server{Addr: ":" + cfg.HTTPPort, Handler: r}
	go func() {
		log.Info("ledger-service starting", logger.FieldService(cfg.ServiceName))
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
		GroupID: "ledger-group",
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
