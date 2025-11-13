package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/zjoart/distributed-notification-system/push-service/pkg/logger"
)

type RedisCache struct {
	client *redis.Client
}

func NewRedisCache(addr, password string, db int) (*RedisCache, error) {
	logger.Info("initializing redis cache connection")

	client := redis.NewClient(&redis.Options{
		Addr:         addr,
		Password:     password,
		DB:           db,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
		PoolSize:     10,
		MinIdleConns: 5,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return &RedisCache{client: client}, nil
}

func (c *RedisCache) Get(ctx context.Context, key string) (string, error) {
	val, err := c.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return "", fmt.Errorf("key not found: %s", key)
	}
	if err != nil {
		return "", fmt.Errorf("failed to get from cache: %w", err)
	}
	return val, nil
}

func (c *RedisCache) Set(ctx context.Context, key string, value string, ttl int) error {
	err := c.client.Set(ctx, key, value, time.Duration(ttl)*time.Second).Err()
	if err != nil {
		return fmt.Errorf("failed to set cache: %w", err)
	}
	return nil
}

func (c *RedisCache) Delete(ctx context.Context, key string) error {
	err := c.client.Del(ctx, key).Err()
	if err != nil {
		return fmt.Errorf("failed to delete from cache: %w", err)
	}
	return nil
}

func (c *RedisCache) Exists(ctx context.Context, key string) (bool, error) {
	count, err := c.client.Exists(ctx, key).Result()
	if err != nil {
		return false, fmt.Errorf("failed to check existence: %w", err)
	}
	return count > 0, nil
}

func (c *RedisCache) SetNX(ctx context.Context, key string, value string, ttl int) (bool, error) {
	set, err := c.client.SetNX(ctx, key, value, time.Duration(ttl)*time.Second).Result()
	if err != nil {
		return false, fmt.Errorf("failed to set NX: %w", err)
	}
	return set, nil
}

func (c *RedisCache) Increment(ctx context.Context, key string) (int64, error) {
	val, err := c.client.Incr(ctx, key).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to increment: %w", err)
	}
	return val, nil
}

func (c *RedisCache) IncrementWithExpiry(ctx context.Context, key string, ttl int) (int64, error) {
	pipe := c.client.Pipeline()
	incrCmd := pipe.Incr(ctx, key)
	pipe.Expire(ctx, key, time.Duration(ttl)*time.Second)

	_, err := pipe.Exec(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to increment with expiry: %w", err)
	}

	return incrCmd.Val(), nil
}

func (c *RedisCache) GetMulti(ctx context.Context, keys []string) ([]string, error) {
	values, err := c.client.MGet(ctx, keys...).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get multiple values: %w", err)
	}

	results := make([]string, len(values))
	for i, v := range values {
		if v != nil {
			results[i] = v.(string)
		}
	}

	return results, nil
}

func (c *RedisCache) SetMulti(ctx context.Context, pairs map[string]string, ttl int) error {
	pipe := c.client.Pipeline()

	for key, value := range pairs {
		pipe.Set(ctx, key, value, time.Duration(ttl)*time.Second)
	}

	_, err := pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to set multiple values: %w", err)
	}

	return nil
}

func (c *RedisCache) CheckRateLimit(ctx context.Context, key string, limit int64, window int) (bool, error) {
	count, err := c.IncrementWithExpiry(ctx, key, window)
	if err != nil {
		return false, err
	}

	return count <= limit, nil
}

func (c *RedisCache) Health(ctx context.Context) error {
	return c.client.Ping(ctx).Err()
}

func (c *RedisCache) Close() error {
	return c.client.Close()
}

func GetIdempotencyKey(notificationID string) string {
	return fmt.Sprintf("idempotency:notification:%s", notificationID)
}

func GetRateLimitKey(userID string) string {
	return fmt.Sprintf("ratelimit:user:%s", userID)
}

func GetDeviceTokenCacheKey(token string) string {
	return fmt.Sprintf("device:token:%s", token)
}
