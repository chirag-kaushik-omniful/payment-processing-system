package config

import (
	"os"
	"strconv"
	"time"
)

type Base struct {
	ServiceName string
	HTTPPort    string
	GRPCPort    string
	Env         string
	LogLevel    string

	DatabaseURL string
	RedisURL    string
	KafkaBrokers []string

	JWTSecret       string
	JWTExpiry       time.Duration
	RefreshExpiry   time.Duration

	OTLPEndpoint string
}

func Load(serviceName string) Base {
	return Base{
		ServiceName: serviceName,
		HTTPPort:    getEnv("HTTP_PORT", "8080"),
		GRPCPort:    getEnv("GRPC_PORT", "9090"),
		Env:         getEnv("ENV", "development"),
		LogLevel:    getEnv("LOG_LEVEL", "info"),
		DatabaseURL: getEnv("DATABASE_URL", "postgres://payment:payment@localhost:5432/payment_platform?sslmode=disable"),
		RedisURL:    getEnv("REDIS_URL", "redis://localhost:6379/0"),
		KafkaBrokers: splitEnv("KAFKA_BROKERS", "localhost:9092"),
		JWTSecret:       getEnv("JWT_SECRET", "dev-secret-change-in-production"),
		JWTExpiry:       getDurationEnv("JWT_EXPIRY", 15*time.Minute),
		RefreshExpiry:   getDurationEnv("REFRESH_EXPIRY", 7*24*time.Hour),
		OTLPEndpoint: getEnv("OTEL_EXPORTER_OTLP_ENDPOINT", ""),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getDurationEnv(key string, fallback time.Duration) time.Duration {
	if v := os.Getenv(key); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			return d
		}
	}
	return fallback
}

func splitEnv(key, fallback string) []string {
	v := getEnv(key, fallback)
	var brokers []string
	for _, b := range splitComma(v) {
		if b != "" {
			brokers = append(brokers, b)
		}
	}
	return brokers
}

func splitComma(s string) []string {
	var out []string
	start := 0
	for i := 0; i <= len(s); i++ {
		if i == len(s) || s[i] == ',' {
			part := trim(s[start:i])
			if part != "" {
				out = append(out, part)
			}
			start = i + 1
		}
	}
	return out
}

func trim(s string) string {
	for len(s) > 0 && (s[0] == ' ' || s[0] == '\t') {
		s = s[1:]
	}
	for len(s) > 0 && (s[len(s)-1] == ' ' || s[len(s)-1] == '\t') {
		s = s[:len(s)-1]
	}
	return s
}

func GetIntEnv(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return fallback
}
