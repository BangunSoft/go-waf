package redis_cache

import (
	"context"
	"time"

	"github.com/jahrulnr/go-waf/internal/interface/repository"
	"github.com/jahrulnr/go-waf/pkg/logger"

	"github.com/redis/go-redis/v9"
)

// TTLCache is a Redis-based cache with time-to-live (TTL) expiration.
type TTLCache struct {
	client *redis.Client
	ctx    context.Context
}

// NewCache creates a new TTLCache instance connected to a Redis server.
func NewCache(ctx context.Context, redisClient *redis.Client) repository.CacheInterface {
	return &TTLCache{
		client: redisClient,
		ctx:    ctx,
	}
}

// Set adds a new item to the Redis cache with the specified key, value, and TTL.
func (c *TTLCache) Set(key string, value []byte, ttl time.Duration) {
	err := c.client.Set(c.ctx, key, value, ttl).Err()
	if err != nil {
		logger.Logger("Error setting value in Redis for key: "+key, err.Error()).Error()
	}
}

// Get retrieves the value associated with the given key from the Redis cache.
func (c *TTLCache) Get(key string) ([]byte, bool) {
	serializedValue, err := c.client.Get(c.ctx, key).Result()
	if err == redis.Nil {
		return nil, false // Key does not exist
	} else if err != nil {
		logger.Logger("Error getting value in Redis for key: " + key + ", error: " + err.Error()).Error()
		return nil, false
	}

	return []byte(serializedValue), true
}

// Pop removes and returns the item with the specified key from the Redis cache.
func (c *TTLCache) Pop(key string) ([]byte, bool) {
	serializedValue, err := c.client.GetDel(c.ctx, key).Result()
	if err == redis.Nil {
		return nil, false // Key does not exist
	} else if err != nil {
		logger.Logger("Error popping value in Redis for key: "+key, err.Error()).Error()
		return nil, false
	}

	return []byte(serializedValue), true
}

// Remove removes the item with the specified key from the Redis cache.
func (c *TTLCache) Remove(key string) {
	err := c.client.Del(c.ctx, key).Err()
	if err != nil {
		logger.Logger("Error removing key from Redis for key: "+key, err.Error()).Error()
	}
}

// RemoveByPrefix removes all items with the specified prefix from the Redis cache.
func (c *TTLCache) RemoveByPrefix(prefix string) {
	var cursor uint64
	for {
		keys, newCursor, err := c.client.Scan(c.ctx, cursor, prefix+"*", 0).Result()
		if err != nil {
			logger.Logger("Error retrieving keys from Redis with prefix for key: "+prefix, err.Error()).Error()
			return
		}

		if len(keys) > 0 {
			// Use a pipeline to delete keys more efficiently
			pipe := c.client.Pipeline()
			for _, key := range keys {
				pipe.Del(c.ctx, key)
			}
			if _, err := pipe.Exec(c.ctx); err != nil {
				logger.Logger("Error deleting keys from Redis with prefix for key: "+prefix, err.Error()).Error()
			}
		}

		cursor = newCursor
		if cursor == 0 {
			break
		}
	}
}

// GetTTL returns the remaining time before the specified key expires.
func (c *TTLCache) GetTTL(key string) (time.Duration, bool) {
	ttl, err := c.client.TTL(c.ctx, key).Result()
	if err == redis.Nil {
		return 0, false // Key does not exist
	} else if err != nil {
		logger.Logger("Error getting TTL from Redis for key: "+key, err.Error()).Error()
		return 0, false
	}

	return ttl, true
}
