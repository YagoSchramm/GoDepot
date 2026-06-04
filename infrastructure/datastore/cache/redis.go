package cache

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

func NewRedisCache(client *redis.Client, defaultTTL time.Duration) Cache {
	return &redisCache{
		client:     client,
		defaultTTL: defaultTTL,
	}
}

type redisCache struct {
	client     *redis.Client
	defaultTTL time.Duration
}

func (c *redisCache) Get(key string) ([]byte, bool) {
	data, err := c.client.Get(context.Background(), key).Bytes()
	if err != nil {
		return nil, false
	}

	copied := make([]byte, len(data))
	copy(copied, data)
	return copied, true
}

func (c *redisCache) Set(key string, data []byte) {
	copied := make([]byte, len(data))
	copy(copied, data)

	_ = c.client.Set(context.Background(), key, copied, c.defaultTTL).Err()
}

func (c *redisCache) Invalidate(name string) {
	ctx := context.Background()
	iter := c.client.Scan(ctx, 0, escapeRedisGlob(name)+"*", 100).Iterator()

	var keys []string
	for iter.Next(ctx) {
		keys = append(keys, iter.Val())
		if len(keys) >= 100 {
			_ = c.client.Del(ctx, keys...).Err()
			keys = keys[:0]
		}
	}

	if len(keys) > 0 {
		_ = c.client.Del(ctx, keys...).Err()
	}
}

func escapeRedisGlob(value string) string {
	var escaped []byte
	for i := 0; i < len(value); i++ {
		switch value[i] {
		case '*', '?', '[', ']', '\\':
			escaped = append(escaped, '\\')
		}
		escaped = append(escaped, value[i])
	}
	return string(escaped)
}
