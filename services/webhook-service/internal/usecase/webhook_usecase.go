package usecase

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"strings"

	apperrors "github.com/omniful/payment-platform/shared/errors"
	"github.com/omniful/payment-platform/shared/redis"
	"github.com/omniful/payment-platform/services/webhook-service/internal/model"
)

type WebhookUsecase struct {
	idempotency *redis.IdempotencyStore
	publisher   EventPublisher
}

type EventPublisher interface {
	PublishWebhook(ctx context.Context, paymentID, status, eventID string) error
}

func NewWebhookUsecase(idempotency *redis.IdempotencyStore, publisher EventPublisher) *WebhookUsecase {
	return &WebhookUsecase{idempotency: idempotency, publisher: publisher}
}

func (u *WebhookUsecase) VerifySignature(provider, signature string, body []byte) error {
	secret := providerSecret(provider)
	if secret == "" {
		return apperrors.Internal("webhook secret not configured", nil)
	}
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)
	expected := hex.EncodeToString(mac.Sum(nil))
	sig := strings.TrimPrefix(strings.TrimSpace(signature), "sha256=")
	if !hmac.Equal([]byte(sig), []byte(expected)) {
		return apperrors.Unauthorized("invalid webhook signature")
	}
	return nil
}

func (u *WebhookUsecase) Process(ctx context.Context, provider string, payload model.WebhookPayload) (model.WebhookResponse, error) {
	key := fmt.Sprintf("webhook:%s:%s", provider, payload.EventID)
	var cached model.WebhookResponse
	found, err := u.idempotency.Get(ctx, key, &cached)
	if err != nil {
		return model.WebhookResponse{}, apperrors.Internal("idempotency lookup failed", err)
	}
	if found {
		cached.Replay = true
		return cached, nil
	}

	locked, err := u.idempotency.TryLock(ctx, key)
	if err != nil {
		return model.WebhookResponse{}, apperrors.Internal("idempotency lock failed", err)
	}
	if !locked {
		if found, _ := u.idempotency.Get(ctx, key, &cached); found {
			cached.Replay = true
			return cached, nil
		}
		return model.WebhookResponse{}, apperrors.Conflict("webhook processing in progress")
	}

	resp := model.WebhookResponse{
		EventID:   payload.EventID,
		Processed: true,
	}
	if err := u.idempotency.Set(ctx, key, resp); err != nil {
		return model.WebhookResponse{}, apperrors.Internal("idempotency store failed", err)
	}
	if u.publisher != nil && payload.PaymentID != "" {
		_ = u.publisher.PublishWebhook(ctx, payload.PaymentID, payload.EventType, payload.EventID)
	}
	return resp, nil
}

func providerSecret(provider string) string {
	key := fmt.Sprintf("WEBHOOK_SECRET_%s", strings.ToUpper(provider))
	return os.Getenv(key)
}
