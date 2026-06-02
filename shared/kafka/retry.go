package kafka

import (
	"context"
	"encoding/json"
	"time"
)

type RetryPublisher struct {
	producer   *Producer
	maxRetries int
}

func NewRetryPublisher(producer *Producer, maxRetries int) *RetryPublisher {
	return &RetryPublisher{producer: producer, maxRetries: maxRetries}
}

type RetryEnvelope struct {
	OriginalTopic string          `json:"original_topic"`
	Key           string          `json:"key"`
	Payload       json.RawMessage `json:"payload"`
	Attempt       int             `json:"attempt"`
	LastError     string          `json:"last_error,omitempty"`
	Timestamp     time.Time       `json:"timestamp"`
}

func (r *RetryPublisher) PublishRetry(ctx context.Context, topic, key string, payload json.RawMessage, attempt int, lastErr error) error {
	env := RetryEnvelope{
		OriginalTopic: topic,
		Key:           key,
		Payload:       payload,
		Attempt:       attempt,
		Timestamp:     time.Now(),
	}
	if lastErr != nil {
		env.LastError = lastErr.Error()
	}
	dest := RetryTopic(topic)
	if attempt >= r.maxRetries {
		dest = DLQTopic(topic)
	}
	return r.producer.Publish(ctx, dest, key, env)
}
