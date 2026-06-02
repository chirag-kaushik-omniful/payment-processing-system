package usecase

import (
	"context"
	"fmt"
	"strings"
	"time"

	goredis "github.com/redis/go-redis/v9"

	"github.com/omniful/payment-platform/services/fraud-service/internal/model"
)

const (
	userVelocityLimit    = 10
	userVelocityWindow   = time.Minute
	highAmountThreshold  = 10000
	highRiskScoreReject  = 80
)

var blockedCountries = map[string]bool{"XX": true, "ZZ": true}

type FraudUsecase struct {
	redis *goredis.Client
}

func NewFraudUsecase(redis *goredis.Client) *FraudUsecase {
	return &FraudUsecase{redis: redis}
}

func blocklistKey(kind, id string) string {
	return fmt.Sprintf("fraud:blocklist:%s:%s", kind, id)
}

func userVelocityKey(userID string) string {
	return fmt.Sprintf("fraud:user:%s:velocity", userID)
}

func ipVelocityKey(ip string) string {
	return fmt.Sprintf("fraud:ip:%s", ip)
}

func deviceKey(deviceID string) string {
	return fmt.Sprintf("fraud:device:%s", deviceID)
}

func (u *FraudUsecase) IsBlocked(ctx context.Context, kind, id string) bool {
	if u.redis == nil || id == "" {
		return false
	}
	v, err := u.redis.Exists(ctx, blocklistKey(kind, id)).Result()
	return err == nil && v > 0
}

func (u *FraudUsecase) AddToBlocklist(ctx context.Context, kind, id string) error {
	if u.redis == nil {
		return nil
	}
	return u.redis.Set(ctx, blocklistKey(kind, id), "1", 0).Err()
}

func (u *FraudUsecase) Evaluate(ctx context.Context, evt model.FraudValidateEvent) model.FraudValidatedEvent {
	score := u.computeRiskScore(evt)
	result := model.FraudValidatedEvent{
		PaymentID: evt.PaymentID,
		Amount:    evt.Amount,
		Approved:  score < highRiskScoreReject,
		RiskScore: score,
	}

	if u.IsBlocked(ctx, "user", evt.UserID) || u.IsBlocked(ctx, "ip", evt.IP) || u.IsBlocked(ctx, "device", evt.DeviceID) {
		result.Approved = false
		result.Reason = "blocklisted"
		return result
	}

	if evt.Country != "" && blockedCountries[strings.ToUpper(evt.Country)] {
		result.Approved = false
		result.Reason = "blocked country"
		return result
	}

	if evt.Amount >= highAmountThreshold {
		result.Approved = false
		result.Reason = "amount exceeds threshold"
		return result
	}

	if allowed, reason := u.checkVelocity(ctx, userVelocityKey(evt.UserID), userVelocityLimit, userVelocityWindow); !allowed {
		result.Approved = false
		result.Reason = reason
		return result
	}

	if evt.IP != "" {
		if allowed, reason := u.checkVelocity(ctx, ipVelocityKey(evt.IP), 50, userVelocityWindow); !allowed {
			result.Approved = false
			result.Reason = reason
			return result
		}
	}

	if evt.DeviceID != "" {
		if allowed, reason := u.checkVelocity(ctx, deviceKey(evt.DeviceID), 20, userVelocityWindow); !allowed {
			result.Approved = false
			result.Reason = reason
			return result
		}
	}

	return result
}

func (u *FraudUsecase) computeRiskScore(evt model.FraudValidateEvent) int {
	score := 10
	if evt.Amount > 1000 {
		score += 20
	}
	if evt.Amount > 5000 {
		score += 30
	}
	if evt.Country != "" && evt.Country != "US" {
		score += 15
	}
	if evt.DeviceID == "" {
		score += 10
	}
	return score
}

func (u *FraudUsecase) checkVelocity(ctx context.Context, key string, limit int64, window time.Duration) (bool, string) {
	if u.redis == nil {
		return true, ""
	}
	pipe := u.redis.TxPipeline()
	incr := pipe.Incr(ctx, key)
	pipe.Expire(ctx, key, window)
	if _, err := pipe.Exec(ctx); err != nil {
		return true, ""
	}
	if incr.Val() > limit {
		return false, "velocity limit exceeded"
	}
	return true, ""
}
