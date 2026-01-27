package db

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

// Implementation of Redis storage methods

// Settings operations
func (r *RedisStorage) GetSetting(ctx context.Context, key string) (string, error) {
	value, err := r.client.Get(ctx, settingKey(key)).Result()
	if err == redis.Nil {
		return "", fmt.Errorf("setting not found: %s", key)
	}
	return value, err
}

func (r *RedisStorage) SetSetting(ctx context.Context, key, value string) error {
	return r.client.Set(ctx, settingKey(key), value, 0).Err()
}

func (r *RedisStorage) DeleteSetting(ctx context.Context, key string) error {
	return r.client.Del(ctx, settingKey(key))
}

// Feed operations
func (r *RedisStorage) CreateFeed(ctx context.Context, feed map[string]interface{}) (int64, error) {
	id, ok := feed["id"].(int64)
	if !ok {
		return 0, fmt.Errorf("feed ID is required")
	}

	// Serialize feed data
	data, err := json.Marshal(feed)
	if err != nil {
		return 0, fmt.Errorf("failed to marshal feed: %w", err)
	}

	pipe := r.client.TxPipeline()

	// Store feed by ID
	pipe.Set(ctx, feedKey(id), data, 0)

	// Store URL to ID mapping for lookups
	if url, ok := feed["url"].(string); ok {
		pipe.Set(ctx, feedURLKey(url), id, 0)
	}

	_, err = pipe.Exec(ctx)
	return id, err
}

func (r *RedisStorage) GetFeed(ctx context.Context, id int64) (map[string]interface{}, error) {
	data, err := r.client.Get(ctx, feedKey(id)).Result()
	if err == redis.Nil {
		return nil, fmt.Errorf("feed not found: %d", id)
	}
	if err != nil {
		return nil, err
	}

	var feed map[string]interface{}
	if err := json.Unmarshal([]byte(data), &feed); err != nil {
		return nil, fmt.Errorf("failed to unmarshal feed: %w", err)
	}

	return feed, nil
}

func (r *RedisStorage) GetFeedByURL(ctx context.Context, url string) (map[string]interface{}, error) {
	id, err := r.client.Get(ctx, feedURLKey(url)).Result()
	if err == redis.Nil {
		return nil, fmt.Errorf("feed not found for URL: %s", url)
	}
	if err != nil {
		return nil, err
	}

	feedID, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid feed ID: %w", err)
	}

	return r.GetFeed(ctx, feedID)
}

func (r *RedisStorage) UpdateFeed(ctx context.Context, id int64, updates map[string]interface{}) error {
	feed, err := r.GetFeed(ctx, id)
	if err != nil {
		return err
	}

	// Merge updates
	for key, value := range updates {
		feed[key] = value
	}

	// Serialize updated feed
	data, err := json.Marshal(feed)
	if err != nil {
		return fmt.Errorf("failed to marshal updated feed: %w", err)
	}

	return r.client.Set(ctx, feedKey(id), data, 0).Err()
}

func (r *RedisStorage) DeleteFeed(ctx context.Context, id int64) error {
	feed, err := r.GetFeed(ctx, id)
	if err != nil {
		return err
	}

	pipe := r.client.TxPipeline()

	// Delete feed by ID
	pipe.Del(ctx, feedKey(id))

	// Delete URL mapping
	if url, ok := feed["url"].(string); ok {
		pipe.Del(ctx, feedURLKey(url))
	}

	_, err = pipe.Exec(ctx)
	return err
}

func (r *RedisStorage) ListFeeds(ctx context.Context, folderID *int64) ([]map[string]interface{}, error) {
	// For simplicity, we'll get all feeds with prefix "feed:"
	// In a production system, you might want to use Redis sets or sorted sets
	iter := r.client.Scan(ctx, 0, "feed:", 0).Iterator()

	var feeds []map[string]interface{}
	for iter.Next(ctx) {
		key := iter.Val()
		
		// Skip URL index keys
		if len(key) > 8 && key[:8] == "feed:url" {
			continue
		}

		data, err := r.client.Get(ctx, key).Result()
		if err != nil {
			continue
		}

		var feed map[string]interface{}
		if err := json.Unmarshal([]byte(data), &feed); err != nil {
			continue
		}

		// Filter by folder ID if specified
		if folderID != nil {
			if feedFolderID, ok := feed["folder_id"]; !ok || feedFolderID == nil {
				if *folderID != 0 {
					continue
				}
			} else if feedFolderIDFloat, ok := feedFolderID.(float64); ok {
				if int64(feedFolderIDFloat) != *folderID {
					continue
				}
			}
		}

		feeds = append(feeds, feed)
	}

	if err := iter.Err(); err != nil {
		return nil, err
	}

	return feeds, nil
}

// Entry operations
func (r *RedisStorage) CreateEntry(ctx context.Context, entry map[string]interface{}) error {
	id, ok := entry["id"].(int64)
	if !ok {
		return fmt.Errorf("entry ID is required")
	}

	feedID, ok := entry["feed_id"].(int64)
	if !ok {
		return fmt.Errorf("feed ID is required for entry")
	}

	// Serialize entry data
	data, err := json.Marshal(entry)
	if err != nil {
		return fmt.Errorf("failed to marshal entry: %w", err)
	}

	pipe := r.client.TxPipeline()

	// Store entry by ID
	pipe.Set(ctx, entryKey(id), data, 0)

	// Add to feed's entry set
	pipe.SAdd(ctx, entryFeedKey(feedID), id)

	_, err = pipe.Exec(ctx)
	return err
}

func (r *RedisStorage) GetEntry(ctx context.Context, id int64) (map[string]interface{}, error) {
	data, err := r.client.Get(ctx, entryKey(id)).Result()
	if err == redis.Nil {
		return nil, fmt.Errorf("entry not found: %d", id)
	}
	if err != nil {
		return nil, err
	}

	var entry map[string]interface{}
	if err := json.Unmarshal([]byte(data), &entry); err != nil {
		return nil, fmt.Errorf("failed to unmarshal entry: %w", err)
	}

	return entry, nil
}

func (r *RedisStorage) UpdateEntry(ctx context.Context, id int64, updates map[string]interface{}) error {
	entry, err := r.GetEntry(ctx, id)
	if err != nil {
		return err
	}

	// Merge updates
	for key, value := range updates {
		entry[key] = value
	}

	// Serialize updated entry
	data, err := json.Marshal(entry)
	if err != nil {
		return fmt.Errorf("failed to marshal updated entry: %w", err)
	}

	return r.client.Set(ctx, entryKey(id), data, 0).Err()
}

func (r *RedisStorage) DeleteEntry(ctx context.Context, id int64) error {
	entry, err := r.GetEntry(ctx, id)
	if err != nil {
		return err
	}

	pipe := r.client.TxPipeline()

	// Delete entry by ID
	pipe.Del(ctx, entryKey(id))

	// Remove from feed's entry set
	if feedID, ok := entry["feed_id"].(float64); ok {
		pipe.SRem(ctx, entryFeedKey(int64(feedID)), id)
	}

	_, err = pipe.Exec(ctx)
	return err
}

func (r *RedisStorage) ListEntries(ctx context.Context, filter map[string]interface{}) ([]map[string]interface{}, error) {
	// This is a simplified implementation
	// For production use, consider using Redis search or a more sophisticated indexing strategy
	iter := r.client.Scan(ctx, 0, "entry:", 0).Iterator()

	var entries []map[string]interface{}
	for iter.Next(ctx) {
		data, err := r.client.Get(ctx, iter.Val()).Result()
		if err != nil {
			continue
		}

		var entry map[string]interface{}
		if err := json.Unmarshal([]byte(data), &entry); err != nil {
			continue
		}

		// Apply filters
		matches := true
		for key, value := range filter {
			if entryValue, exists := entry[key]; !exists || entryValue != value {
				matches = false
				break
			}
		}

		if matches {
			entries = append(entries, entry)
		}
	}

	if err := iter.Err(); err != nil {
		return nil, err
	}

	return entries, nil
}

// Folder operations
func (r *RedisStorage) CreateFolder(ctx context.Context, folder map[string]interface{}) (int64, error) {
	id, ok := folder["id"].(int64)
	if !ok {
		return 0, fmt.Errorf("folder ID is required")
	}

	data, err := json.Marshal(folder)
	if err != nil {
		return 0, fmt.Errorf("failed to marshal folder: %w", err)
	}

	return id, r.client.Set(ctx, folderKey(id), data, 0).Err()
}

func (r *RedisStorage) GetFolder(ctx context.Context, id int64) (map[string]interface{}, error) {
	data, err := r.client.Get(ctx, folderKey(id)).Result()
	if err == redis.Nil {
		return nil, fmt.Errorf("folder not found: %d", id)
	}
	if err != nil {
		return nil, err
	}

	var folder map[string]interface{}
	if err := json.Unmarshal([]byte(data), &folder); err != nil {
		return nil, fmt.Errorf("failed to unmarshal folder: %w", err)
	}

	return folder, nil
}

func (r *RedisStorage) UpdateFolder(ctx context.Context, id int64, updates map[string]interface{}) error {
	folder, err := r.GetFolder(ctx, id)
	if err != nil {
		return err
	}

	for key, value := range updates {
		folder[key] = value
	}

	data, err := json.Marshal(folder)
	if err != nil {
		return fmt.Errorf("failed to marshal updated folder: %w", err)
	}

	return r.client.Set(ctx, folderKey(id), data, 0).Err()
}

func (r *RedisStorage) DeleteFolder(ctx context.Context, id int64) error {
	return r.client.Del(ctx, folderKey(id))
}

func (r *RedisStorage) ListFolders(ctx context.Context) ([]map[string]interface{}, error) {
	iter := r.client.Scan(ctx, 0, "folder:", 0).Iterator()

	var folders []map[string]interface{}
	for iter.Next(ctx) {
		data, err := r.client.Get(ctx, iter.Val()).Result()
		if err != nil {
			continue
		}

		var folder map[string]interface{}
		if err := json.Unmarshal([]byte(data), &folder); err != nil {
			continue
		}

		folders = append(folders, folder)
	}

	if err := iter.Err(); err != nil {
		return nil, err
	}

	return folders, nil
}

// AI operations
func (r *RedisStorage) GetAISummary(ctx context.Context, entryID int64, language string) (string, error) {
	key := aiSummaryKey(entryID, language)
	summary, err := r.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return "", fmt.Errorf("AI summary not found for entry %d, language %s", entryID, language)
	}
	return summary, err
}

func (r *RedisStorage) SetAISummary(ctx context.Context, entryID int64, language, summary string) error {
	key := aiSummaryKey(entryID, language)
	// Cache for 30 days
	return r.client.Set(ctx, key, summary, 30*24*time.Hour).Err()
}

func (r *RedisStorage) DeleteAISummary(ctx context.Context, entryID int64, language string) error {
	key := aiSummaryKey(entryID, language)
	return r.client.Del(ctx, key)
}

func (r *RedisStorage) GetAITranslation(ctx context.Context, entryID int64, language string) (string, error) {
	key := aiTranslationKey(entryID, language)
	translation, err := r.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return "", fmt.Errorf("AI translation not found for entry %d, language %s", entryID, language)
	}
	return translation, err
}

func (r *RedisStorage) SetAITranslation(ctx context.Context, entryID int64, language, content string) error {
	key := aiTranslationKey(entryID, language)
	// Cache for 30 days
	return r.client.Set(ctx, key, content, 30*24*time.Hour).Err()
}

func (r *RedisStorage) DeleteAITranslation(ctx context.Context, entryID int64, language string) error {
	key := aiTranslationKey(entryID, language)
	return r.client.Del(ctx, key)
}

// Rate limiting operations
func (r *RedisStorage) GetRateLimit(ctx context.Context, host string) (int64, error) {
	key := domainRateLimitKey(host)
	interval, err := r.client.Get(ctx, key).Int64()
	if err == redis.Nil {
		return 0, fmt.Errorf("rate limit not found for host: %s", host)
	}
	return interval, err
}

func (r *RedisStorage) SetRateLimit(ctx context.Context, host string, intervalSeconds int64) error {
	key := domainRateLimitKey(host)
	// Store indefinitely
	return r.client.Set(ctx, key, intervalSeconds, 0).Err()
}

func (r *RedisStorage) DeleteRateLimit(ctx context.Context, host string) error {
	key := domainRateLimitKey(host)
	return r.client.Del(ctx, key)
}