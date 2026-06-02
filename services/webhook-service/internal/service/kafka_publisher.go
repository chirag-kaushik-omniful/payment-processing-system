package service

import (
	"context"

	"github.com/omniful/payment-platform/shared/kafka"
)

type KafkaPublisher struct {
	producer *kafka.Producer
}

func NewKafkaPublisher(producer *kafka.Producer) *KafkaPublisher {
	return &KafkaPublisher{producer: producer}
}

func (p *KafkaPublisher) PublishWebhook(ctx context.Context, paymentID, status, eventID string) error {
	return p.producer.Publish(ctx, kafka.TopicWebhookReceived, paymentID, map[string]interface{}{
		"payment_id": paymentID,
		"status":     status,
		"event_id":   eventID,
	})
}
