package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/aditi2420/fleet-tracker/internal/model"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

const (
	defaultTTL      = 5 * time.Minute
	statusKeyFormat = "vehicle:status:%s"
)

// --- concrete type ----------------------------------------------------------

type redisCache struct {
	rdb *redis.Client
	ttl time.Duration
}

// NewRedis initialize
func NewRedis(addr string, password string, db int, ttl time.Duration) (VehicleCache, error) {
	if ttl == 0 {
		ttl = defaultTTL
	}
	rdb := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})

	if err := rdb.Ping(context.Background()).Err(); err != nil {
		return nil, err
	}
	return &redisCache{rdb: rdb, ttl: ttl}, nil
}

func keyStatus(id uuid.UUID) string {
	return fmt.Sprintf(statusKeyFormat, id.String())
}

func (c *redisCache) GetStatus(ctx context.Context, id uuid.UUID) (*model.Status, error) {
	val, err := c.rdb.Get(ctx, keyStatus(id)).Bytes()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var st model.Status
	if err := json.Unmarshal(val, &st); err != nil {
		_ = c.rdb.Del(ctx, keyStatus(id)).Err()
		return nil, nil
	}
	return &st, nil
}

func (c *redisCache) SetStatus(ctx context.Context, id uuid.UUID, s model.Status) error {
	b, err := json.Marshal(s)
	if err != nil {
		return err
	}
	return c.rdb.Set(ctx, keyStatus(id), b, c.ttl).Err()
}

// TTL exposes the configured expiration – handy for tests & metrics.
func (c *redisCache) TTL() time.Duration { return c.ttl }

// Close satisfies io.Closer for graceful shutdown.
func (c *redisCache) Close() error { return c.rdb.Close() }
