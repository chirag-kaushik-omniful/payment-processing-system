package redis

import (
	"context"
	"fmt"
	"time"

	goredis "github.com/redis/go-redis/v9"
)

type RateLimiter struct {
	client *goredis.Client
	limit  int64
	window time.Duration
}

func NewRateLimiter(client *goredis.Client, limit int64, window time.Duration) *RateLimiter {
	return &RateLimiter{client: client, limit: limit, window: window}
}

func rateLimitKey(userID string) string {
	return fmt.Sprintf("ratelimit:%s", userID)
}

// Allow implements a sliding-window counter using Redis INCR + EXPIRE.
func (r *RateLimiter) Allow(ctx context.Context, userID string) (bool, error) {
	key := rateLimitKey(userID)
	pipe := r.client.TxPipeline()
	incr := pipe.Incr(ctx, key)
	pipe.Expire(ctx, key, r.window)
	if _, err := pipe.Exec(ctx); err != nil {
		return false, err
	}
	return incr.Val() <= r.limit, nil
}
