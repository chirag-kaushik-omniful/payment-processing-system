package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/segmentio/kafka-go"
)

type Producer struct {
	writers map[string]*kafka.Writer
	brokers []string
}

func NewProducer(brokers []string) *Producer {
	return &Producer{
		writers: make(map[string]*kafka.Writer),
		brokers: brokers,
	}
}

func (p *Producer) Publish(ctx context.Context, topic, key string, payload interface{}) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	w := p.writer(topic)
	msg := kafka.Message{
		Topic: topic,
		Key:   []byte(key),
		Value: data,
		Time:  time.Now(),
	}
	return w.WriteMessages(ctx, msg)
}

func (p *Producer) writer(topic string) *kafka.Writer {
	if w, ok := p.writers[topic]; ok {
		return w
	}
	w := &kafka.Writer{
		Addr:         kafka.TCP(p.brokers...),
		Topic:        topic,
		Balancer:     &kafka.Hash{},
		RequiredAcks: kafka.RequireOne,
		Async:        false,
	}
	p.writers[topic] = w
	return w
}

func (p *Producer) Close() error {
	var errs []error
	for topic, w := range p.writers {
		if err := w.Close(); err != nil {
			errs = append(errs, fmt.Errorf("close %s: %w", topic, err))
		}
	}
	if len(errs) > 0 {
		return errs[0]
	}
	return nil
}
