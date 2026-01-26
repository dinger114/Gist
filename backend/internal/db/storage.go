package db

import (
	"context"
	"database/sql"

	"gist/backend/internal/config"
)

// Storage interface abstracts the database operations
// This allows us to support both SQLite and Redis backends
type Storage interface {
	// Settings operations
	GetSetting(ctx context.Context, key string) (string, error)
	SetSetting(ctx context.Context, key, value string) error
	DeleteSetting(ctx context.Context, key string) error

	// Feed operations
	CreateFeed(ctx context.Context, feed map[string]interface{}) (int64, error)
	GetFeed(ctx context.Context, id int64) (map[string]interface{}, error)
	GetFeedByURL(ctx context.Context, url string) (map[string]interface{}, error)
	UpdateFeed(ctx context.Context, id int64, updates map[string]interface{}) error
	DeleteFeed(ctx context.Context, id int64) error
	ListFeeds(ctx context.Context, folderID *int64) ([]map[string]interface{}, error)

	// Entry operations
	CreateEntry(ctx context.Context, entry map[string]interface{}) error
	GetEntry(ctx context.Context, id int64) (map[string]interface{}, error)
	UpdateEntry(ctx context.Context, id int64, updates map[string]interface{}) error
	DeleteEntry(ctx context.Context, id int64) error
	ListEntries(ctx context.Context, filter map[string]interface{}) ([]map[string]interface{}, error)

	// Folder operations
	CreateFolder(ctx context.Context, folder map[string]interface{}) (int64, error)
	GetFolder(ctx context.Context, id int64) (map[string]interface{}, error)
	UpdateFolder(ctx context.Context, id int64, updates map[string]interface{}) error
	DeleteFolder(ctx context.Context, id int64) error
	ListFolders(ctx context.Context) ([]map[string]interface{}, error)

	// AI operations
	GetAISummary(ctx context.Context, entryID int64, language string) (string, error)
	SetAISummary(ctx context.Context, entryID int64, language, summary string) error
	DeleteAISummary(ctx context.Context, entryID int64, language string) error

	GetAITranslation(ctx context.Context, entryID int64, language string) (string, error)
	SetAITranslation(ctx context.Context, entryID int64, language, content string) error
	DeleteAITranslation(ctx context.Context, entryID int64, language string) error

	// Rate limiting
	GetRateLimit(ctx context.Context, host string) (int64, error)
	SetRateLimit(ctx context.Context, host string, intervalSeconds int64) error
	DeleteRateLimit(ctx context.Context, host string) error

	// Cleanup operations
	Close() error
}

// NewStorage creates a new storage instance based on configuration
func NewStorage(cfg config.Config) (Storage, error) {
	switch cfg.StorageType {
	case "redis", "upstash":
		if cfg.RedisURL == "" {
			return nil, ErrRedisConfigRequired
		}
		return NewRedisStorage(cfg.RedisURL, cfg.RedisToken)
	case "sqlite":
		fallthrough
	default:
		return NewSQLiteStorage(cfg.DBPath)
	}
}

// SQLite storage implementation
func NewSQLiteStorage(dbPath string) (*SQLiteStorage, error) {
	db, err := Open(dbPath)
	if err != nil {
		return nil, err
	}
	return &SQLiteStorage{db: db}, nil
}

type SQLiteStorage struct {
	db *sql.DB
}

func (s *SQLiteStorage) Close() error {
	return s.db.Close()
}

// Redis storage implementation
func NewRedisStorage(url, token string) (*RedisStorage, error) {
	client, err := NewRedisClient(url, token)
	if err != nil {
		return nil, err
	}
	return &RedisStorage{client: client}, nil
}

type RedisStorage struct {
	client *RedisClient
}

func (r *RedisStorage) Close() error {
	return r.client.Close()
}

var (
	ErrRedisConfigRequired = &StorageError{msg: "Redis URL and token are required for Redis storage"}
	ErrStorageNotImplemented = &StorageError{msg: "operation not implemented for this storage type"}
)

type StorageError struct {
	msg string
}

func (e *StorageError) Error() string {
	return e.msg
}