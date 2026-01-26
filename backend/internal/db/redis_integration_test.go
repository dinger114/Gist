package db

import (
	"context"
	"os"
	"testing"
	"time"
)

func TestRedisClient(t *testing.T) {
	// Skip test if Redis environment variables are not set
	redisURL := os.Getenv("UPSTASH_URL")
	redisToken := os.Getenv("UPSTASH_TOKEN")

	if redisURL == "" || redisToken == "" {
		t.Skip("Skipping Redis integration test: UPSTASH_URL and UPSTASH_TOKEN environment variables are required")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Test Redis client creation
	client, err := NewRedisClient(redisURL, redisToken)
	if err != nil {
		t.Fatalf("Failed to create Redis client: %v", err)
	}
	defer client.Close()

	// Test basic operations
	t.Run("Basic Set/Get", func(t *testing.T) {
		key := "test:key"
		value := "test-value"

		// Set a value
		err := client.Set(ctx, key, value, 0).Err()
		if err != nil {
			t.Fatalf("Failed to set key: %v", err)
		}

		// Get the value
		result, err := client.Get(ctx, key).Result()
		if err != nil {
			t.Fatalf("Failed to get key: %v", err)
		}

		if result != value {
			t.Errorf("Expected value %s, got %s", value, result)
		}

		// Clean up
		err = client.Del(ctx, key).Err()
		if err != nil {
			t.Logf("Warning: Failed to delete test key: %v", err)
		}
	})

	t.Run("Feed Key Generation", func(t *testing.T) {
		feedID := int64(12345)
		key := feedKey(feedID)
		expected := "feed:12345"
		if key != expected {
			t.Errorf("Expected key %s, got %s", expected, key)
		}
	})

	t.Run("Entry Key Generation", func(t *testing.T) {
		entryID := int64(67890)
		key := entryKey(entryID)
		expected := "entry:67890"
		if key != expected {
			t.Errorf("Expected key %s, got %s", expected, key)
		}
	})

	t.Run("Settings Key Generation", func(t *testing.T) {
		settingName := "test.setting"
		key := settingKey(settingName)
		expected := "setting:test.setting"
		if key != expected {
			t.Errorf("Expected key %s, got %s", expected, key)
		}
	})

	t.Run("AI Summary Key Generation", func(t *testing.T) {
		entryID := int64(111)
		language := "en"
		key := aiSummaryKey(entryID, language)
		expected := "ai:summary:111:en"
		if key != expected {
			t.Errorf("Expected key %s, got %s", expected, key)
		}
	})
}

func TestRedisStorage(t *testing.T) {
	// Skip test if Redis environment variables are not set
	redisURL := os.Getenv("UPSTASH_URL")
	redisToken := os.Getenv("UPSTASH_TOKEN")

	if redisURL == "" || redisToken == "" {
		t.Skip("Skipping Redis storage test: UPSTASH_URL and UPSTASH_TOKEN environment variables are required")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Test storage creation
	storage, err := NewRedisStorage(redisURL, redisToken)
	if err != nil {
		t.Fatalf("Failed to create Redis storage: %v", err)
	}
	defer storage.Close()

	t.Run("Settings Operations", func(t *testing.T) {
		key := "test:setting"
		value := "test-value"

		// Test setting a value
		err := storage.SetSetting(ctx, key, value)
		if err != nil {
			t.Fatalf("Failed to set setting: %v", err)
		}

		// Test getting the value
		retrieved, err := storage.GetSetting(ctx, key)
		if err != nil {
			t.Fatalf("Failed to get setting: %v", err)
		}

		if retrieved != value {
			t.Errorf("Expected value %s, got %s", value, retrieved)
		}

		// Test deleting the value
		err = storage.DeleteSetting(ctx, key)
		if err != nil {
			t.Fatalf("Failed to delete setting: %v", err)
		}

		// Verify deletion
		_, err = storage.GetSetting(ctx, key)
		if err == nil {
			t.Error("Expected error when getting deleted setting, got nil")
		}
	})

	t.Run("Feed Operations", func(t *testing.T) {
		feed := map[string]interface{}{
			"id":            int64(99999),
			"title":         "Test Feed",
			"url":           "https://example.com/feed.xml",
			"type":          "article",
			"created_at":    time.Now().UTC().Format(time.RFC3339),
			"updated_at":    time.Now().UTC().Format(time.RFC3339),
		}

		// Test creating a feed
		id, err := storage.CreateFeed(ctx, feed)
		if err != nil {
			t.Fatalf("Failed to create feed: %v", err)
		}

		if id != feed["id"] {
			t.Errorf("Expected ID %d, got %d", feed["id"], id)
		}

		// Test getting the feed
		retrieved, err := storage.GetFeed(ctx, id)
		if err != nil {
			t.Fatalf("Failed to get feed: %v", err)
		}

		if retrieved["title"] != feed["title"] {
			t.Errorf("Expected title %s, got %s", feed["title"], retrieved["title"])
		}

		// Test getting feed by URL
		urlRetrieved, err := storage.GetFeedByURL(ctx, feed["url"].(string))
		if err != nil {
			t.Fatalf("Failed to get feed by URL: %v", err)
		}

		if urlRetrieved["id"] != feed["id"] {
			t.Errorf("Expected ID %d, got %v", feed["id"], urlRetrieved["id"])
		}

		// Test deleting the feed
		err = storage.DeleteFeed(ctx, id)
		if err != nil {
			t.Fatalf("Failed to delete feed: %v", err)
		}

		// Verify deletion
		_, err = storage.GetFeed(ctx, id)
		if err == nil {
			t.Error("Expected error when getting deleted feed, got nil")
		}
	})
}