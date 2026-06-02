package service

import (
	"context"
	"encoding/json"

	"go.uber.org/zap"

	"github.com/omniful/payment-platform/services/notification-service/internal/channels"
	"github.com/omniful/payment-platform/services/notification-service/internal/hub"
)

type NotificationConsumer struct {
	hub      *hub.Hub
	notifier *channels.Notifier
	log      *zap.Logger
}

func NewNotificationConsumer(h *hub.Hub, notifier *channels.Notifier, log *zap.Logger) *NotificationConsumer {
	return &NotificationConsumer{hub: h, notifier: notifier, log: log}
}

func (c *NotificationConsumer) HandleNotificationSend(ctx context.Context, key string, raw json.RawMessage) error {
	var evt struct {
		PaymentID string `json:"payment_id"`
		Status    string `json:"status"`
		UserID    string `json:"user_id"`
	}
	if err := json.Unmarshal(raw, &evt); err != nil {
		var paymentID string
		if err2 := json.Unmarshal(raw, &paymentID); err2 == nil && paymentID != "" {
			evt.PaymentID = paymentID
			evt.Status = "completed"
		} else {
			return err
		}
	}
	if evt.Status == "" {
		evt.Status = "completed"
	}
	c.hub.Broadcast(hub.PaymentStatusMessage{
		PaymentID: evt.PaymentID,
		Status:    evt.Status,
		UserID:    evt.UserID,
	})
	if c.notifier != nil {
		c.notifier.NotifyPayment(evt.UserID, evt.PaymentID, evt.Status)
	}
	c.log.Info("notification broadcast",
		zap.String("payment_id", evt.PaymentID),
		zap.String("status", evt.Status),
	)
	return nil
}
