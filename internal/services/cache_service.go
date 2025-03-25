package services

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
)

// CacheService provides Redis-based caching functionality
type CacheService struct {
	client *redis.Client
	ctx    context.Context
}

// NewCacheService creates a new Redis cache service
func NewCacheService(redisURL string) (*CacheService, error) {
	opts, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, fmt.Errorf("invalid Redis URL: %v", err)
	}

	client := redis.NewClient(opts)
	ctx := context.Background()

	// Test connection
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %v", err)
	}

	return &CacheService{
		client: client,
		ctx:    ctx,
	}, nil
}

// Set stores a value in the cache with the specified key and expiration
func (s *CacheService) Set(key string, value interface{}, expiration time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal value: %v", err)
	}

	return s.client.Set(s.ctx, key, data, expiration).Err()
}

// Get retrieves a value from the cache
func (s *CacheService) Get(key string, dest interface{}) error {
	data, err := s.client.Get(s.ctx, key).Bytes()
	if err != nil {
		if err == redis.Nil {
			return fmt.Errorf("key not found: %s", key)
		}
		return fmt.Errorf("failed to get value: %v", err)
	}

	if err := json.Unmarshal(data, dest); err != nil {
		return fmt.Errorf("failed to unmarshal value: %v", err)
	}

	return nil
}

// Delete removes a value from the cache
func (s *CacheService) Delete(key string) error {
	return s.client.Del(s.ctx, key).Err()
}

// GetOrSet gets a value from the cache or sets it if not found
func (s *CacheService) GetOrSet(key string, dest interface{}, expiration time.Duration, fn func() (interface{}, error)) error {
	err := s.Get(key, dest)
	if err == nil {
		return nil
	}

	value, err := fn()
	if err != nil {
		return fmt.Errorf("failed to generate value: %v", err)
	}

	if err := s.Set(key, value, expiration); err != nil {
		return fmt.Errorf("failed to set value: %v", err)
	}

	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal value: %v", err)
	}

	return json.Unmarshal(data, dest)
}

// ClearPattern deletes all keys matching the pattern
func (s *CacheService) ClearPattern(pattern string) error {
	keys, err := s.client.Keys(s.ctx, pattern).Result()
	if err != nil {
		return fmt.Errorf("failed to find keys: %v", err)
	}

	if len(keys) == 0 {
		return nil
	}

	return s.client.Del(s.ctx, keys...).Err()
}

// Close closes the Redis connection
func (s *CacheService) Close() error {
	return s.client.Close()
} 