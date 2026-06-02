package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	goredis "github.com/redis/go-redis/v9"
)

const idempotencyTTL = 48 * time.Hour

type IdempotencyStore struct {
	client *goredis.Client
}

func NewIdempotencyStore(client *goredis.Client) *IdempotencyStore {
	return &IdempotencyStore{client: client}
}

func idempotencyKey(key string) string {
	return fmt.Sprintf("idempotency:%s", key)
}

func (s *IdempotencyStore) Get(ctx context.Context, key string, dest interface{}) (bool, error) {
	val, err := s.client.Get(ctx, idempotencyKey(key)).Result()
	if err == goredis.Nil {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	if err := json.Unmarshal([]byte(val), dest); err != nil {
		return false, err
	}
	return true, nil
}

func (s *IdempotencyStore) Set(ctx context.Context, key string, value interface{}) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return s.client.Set(ctx, idempotencyKey(key), data, idempotencyTTL).Err()
}

func (s *IdempotencyStore) TryLock(ctx context.Context, key string) (bool, error) {
	return s.client.SetNX(ctx, idempotencyKey(key)+":lock", "1", 30*time.Second).Result()
}
