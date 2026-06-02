package kafka

import (
	"context"
	"encoding/json"
	"time"

	"github.com/segmentio/kafka-go"

	"github.com/omniful/payment-platform/shared/metrics"
)

type HandlerFunc func(ctx context.Context, key string, payload json.RawMessage) error

type Consumer struct {
	reader     *kafka.Reader
	handler    HandlerFunc
	retryPub   *RetryPublisher
	topic      string
	maxRetries int
}

func NewConsumer(cfg Config, topic string, handler HandlerFunc) *Consumer {
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers:        cfg.Brokers,
		GroupID:        cfg.GroupID,
		Topic:          topic,
		MinBytes:       1,
		MaxBytes:       10e6,
		CommitInterval: time.Second,
		StartOffset:    kafka.LastOffset,
	})
	return &Consumer{reader: r, handler: handler, topic: topic, maxRetries: 3}
}

func NewConsumerWithRetry(cfg Config, topic string, producer *Producer, handler HandlerFunc) *Consumer {
	c := NewConsumer(cfg, topic, handler)
	c.retryPub = NewRetryPublisher(producer, 3)
	return c
}

func (c *Consumer) Run(ctx context.Context) error {
	for {
		msg, err := c.reader.FetchMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return ctx.Err()
			}
			continue
		}
		if err := c.handler(ctx, string(msg.Key), msg.Value); err != nil {
			metrics.RetryCount.WithLabelValues(c.reader.Config().GroupID, c.topic).Inc()
			if c.retryPub != nil {
				_ = c.retryPub.PublishRetry(ctx, c.topic, string(msg.Key), msg.Value, 1, err)
			}
			continue
		}
		_ = c.reader.CommitMessages(ctx, msg)
	}
}

func (c *Consumer) Close() error {
	return c.reader.Close()
}
