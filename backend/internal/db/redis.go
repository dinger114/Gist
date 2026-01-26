package db

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisClient struct {
	*redis.Client
}

// NewRedisClient creates a new Redis client for Upstash
func NewRedisClient(url, token string) (*RedisClient, error) {
	if url == "" {
		return nil, fmt.Errorf("Redis URL is required")
	}

	opts, err := redis.ParseURL(url)
	if err != nil {
		return nil, fmt.Errorf("failed to parse Redis URL: %w", err)
	}

	// Override password with token if provided
	if token != "" {
		opts.Password = token
	}

	client := redis.NewClient(opts)

	// Test the connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return &RedisClient{client}, nil
}

// Key generation helpers
func feedKey(id int64) string {
	return fmt.Sprintf("feed:%d", id)
}

func feedURLKey(url string) string {
	return fmt.Sprintf("feed:url:%s", url)
}

func entryKey(id int64) string {
	return fmt.Sprintf("entry:%d", id)
}

func entryFeedKey(feedID int64) string {
	return fmt.Sprintf("feed:%d:entries", feedID)
}

func folderKey(id int64) string {
	return fmt.Sprintf("folder:%d", id)
}

func settingKey(key string) string {
	return fmt.Sprintf("setting:%s", key)
}

func aiSummaryKey(entryID int64, language string) string {
	return fmt.Sprintf("ai:summary:%d:%s", entryID, language)
}

func aiTranslationKey(entryID int64, language string) string {
	return fmt.Sprintf("ai:translation:%d:%s", entryID, language)
}

func domainRateLimitKey(host string) string {
	return fmt.Sprintf("ratelimit:%s", host)
}

// Exists checks if a key exists
func (r *RedisClient) Exists(ctx context.Context, key string) (bool, error) {
	result, err := r.Client.Exists(ctx, key).Result()
	if err != nil {
		return false, err
	}
	return result > 0, nil
}

// Del deletes a key
func (r *RedisClient) Del(ctx context.Context, key string) error {
	return r.Client.Del(ctx, key).Err()
}

// TTL returns the TTL of a key
func (r *RedisClient) TTL(ctx context.Context, key string) (time.Duration, error) {
	return r.Client.TTL(ctx, key).Result()
}

// Expire sets expiration on a key
func (r *RedisClient) Expire(ctx context.Context, key string, ttl time.Duration) error {
	return r.Client.Expire(ctx, key, ttl).Err()
}