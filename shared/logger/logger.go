package logger

import (
	"context"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type ctxKey struct{}

func New(level string) (*zap.Logger, error) {
	cfg := zap.NewProductionConfig()
	switch level {
	case "debug":
		cfg.Level = zap.NewAtomicLevelAt(zapcore.DebugLevel)
	case "warn":
		cfg.Level = zap.NewAtomicLevelAt(zapcore.WarnLevel)
	case "error":
		cfg.Level = zap.NewAtomicLevelAt(zapcore.ErrorLevel)
	default:
		cfg.Level = zap.NewAtomicLevelAt(zapcore.InfoLevel)
	}
	return cfg.Build()
}

func WithContext(ctx context.Context, log *zap.Logger) context.Context {
	return context.WithValue(ctx, ctxKey{}, log)
}

func FromContext(ctx context.Context) *zap.Logger {
	if log, ok := ctx.Value(ctxKey{}).(*zap.Logger); ok && log != nil {
		return log
	}
	return zap.NewNop()
}

func WithFields(ctx context.Context, fields ...zap.Field) context.Context {
	log := FromContext(ctx).With(fields...)
	return WithContext(ctx, log)
}

func FieldTraceID(id string) zap.Field  { return zap.String("trace_id", id) }
func FieldPaymentID(id string) zap.Field { return zap.String("payment_id", id) }
func FieldService(name string) zap.Field { return zap.String("service", name) }
func FieldUserID(id string) zap.Field   { return zap.String("user_id", id) }
