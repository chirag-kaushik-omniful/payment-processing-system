package metrics

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	PaymentSuccess = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "payment_success_total",
		Help: "Total successful payments",
	}, []string{"service"})

	PaymentFailed = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "payment_failed_total",
		Help: "Total failed payments",
	}, []string{"service"})

	ProviderLatency = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "provider_latency_seconds",
		Help:    "Provider call latency",
		Buckets: prometheus.DefBuckets,
	}, []string{"provider"})

	KafkaConsumerLag = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "kafka_consumer_lag",
		Help: "Kafka consumer lag estimate",
	}, []string{"group", "topic"})

	RetryCount = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "retry_total",
		Help: "Total retry attempts",
	}, []string{"service", "topic"})

	HTTPRequests = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "http_requests_total",
		Help: "Total HTTP requests",
	}, []string{"service", "method", "path", "status"})
)

func Handler() http.Handler {
	return promhttp.Handler()
}
