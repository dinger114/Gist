package db

import (
	"context"
	"database/sql"
	"fmt"
)

// Implementation of SQLite storage methods

// Settings operations
func (s *SQLiteStorage) GetSetting(ctx context.Context, key string) (string, error) {
	var value string
	err := s.db.QueryRowContext(ctx, "SELECT value FROM settings WHERE key = ?", key).Scan(&value)
	if err == sql.ErrNoRows {
		return "", fmt.Errorf("setting not found: %s", key)
	}
	return value, err
}

func (s *SQLiteStorage) SetSetting(ctx context.Context, key, value string) error {
	_, err := s.db.ExecContext(ctx,
		"INSERT OR REPLACE INTO settings (key, value, updated_at) VALUES (?, ?, datetime('now'))",
		key, value)
	return err
}

func (s *SQLiteStorage) DeleteSetting(ctx context.Context, key string) error {
	_, err := s.db.ExecContext(ctx, "DELETE FROM settings WHERE key = ?", key)
	return err
}

// Feed operations
func (s *SQLiteStorage) CreateFeed(ctx context.Context, feed map[string]interface{}) (int64, error) {
	id, ok := feed["id"].(int64)
	if !ok {
		return 0, fmt.Errorf("feed ID is required")
	}

	// Extract feed fields with proper type handling
	folderID, _ := getInt64Pointer(feed["folder_id"])
	title, _ := getString(feed["title"])
	url, _ := getString(feed["url"])
	siteURL, _ := getStringPointer(feed["site_url"])
	description, _ := getStringPointer(feed["description"])
	type_, _ := getString(feed["type"])
	etag, _ := getStringPointer(feed["etag"])
	lastModified, _ := getStringPointer(feed["last_modified"])
	errorMessage, _ := getStringPointer(feed["error_message"])
	createdAt, _ := getString(feed["created_at"])
	updatedAt, _ := getString(feed["updated_at"])

	_, err := s.db.ExecContext(ctx,
		`INSERT INTO feeds (id, folder_id, title, url, site_url, description, type, etag, last_modified, error_message, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		id, folderID, title, url, siteURL, description, type_, etag, lastModified, errorMessage, createdAt, updatedAt)

	return id, err
}

func (s *SQLiteStorage) GetFeed(ctx context.Context, id int64) (map[string]interface{}, error) {
	row := s.db.QueryRowContext(ctx,
		"SELECT id, folder_id, title, url, site_url, description, icon_path, type, etag, last_modified, error_message, created_at, updated_at FROM feeds WHERE id = ?",
		id)

	var (
		feedID          int64
		folderID        sql.NullInt64
		title           string
		url             string
		siteURL         sql.NullString
		description     sql.NullString
		iconPath        sql.NullString
		type_           string
		etag            sql.NullString
		lastModified    sql.NullString
		errorMessage    sql.NullString
		createdAt       string
		updatedAt       string
	)

	err := row.Scan(&feedID, &folderID, &title, &url, &siteURL, &description, &iconPath, &type_, &etag, &lastModified, &errorMessage, &createdAt, &updatedAt)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("feed not found: %d", id)
	}
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"id":            feedID,
		"folder_id":     nullableInt64ToInterface(folderID),
		"title":         title,
		"url":           url,
		"site_url":      nullableStringToInterface(siteURL),
		"description":   nullableStringToInterface(description),
		"icon_path":     nullableStringToInterface(iconPath),
		"type":          type_,
		"etag":          nullableStringToInterface(etag),
		"last_modified": nullableStringToInterface(lastModified),
		"error_message": nullableStringToInterface(errorMessage),
		"created_at":    createdAt,
		"updated_at":    updatedAt,
	}, nil
}

func (s *SQLiteStorage) GetFeedByURL(ctx context.Context, url string) (map[string]interface{}, error) {
	row := s.db.QueryRowContext(ctx,
		"SELECT id, folder_id, title, url, site_url, description, icon_path, type, etag, last_modified, error_message, created_at, updated_at FROM feeds WHERE url = ?",
		url)

	var (
		feedID          int64
		folderID        sql.NullInt64
		title           string
		url_            string
		siteURL         sql.NullString
		description     sql.NullString
		iconPath        sql.NullString
		type_           string
		etag            sql.NullString
		lastModified    sql.NullString
		errorMessage    sql.NullString
		createdAt       string
		updatedAt       string
	)

	err := row.Scan(&feedID, &folderID, &title, &url_, &siteURL, &description, &iconPath, &type_, &etag, &lastModified, &errorMessage, &createdAt, &updatedAt)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("feed not found for URL: %s", url)
	}
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"id":            feedID,
		"folder_id":     nullableInt64ToInterface(folderID),
		"title":         title,
		"url":           url_,
		"site_url":      nullableStringToInterface(siteURL),
		"description":   nullableStringToInterface(description),
		"icon_path":     nullableStringToInterface(iconPath),
		"type":          type_,
		"etag":          nullableStringToInterface(etag),
		"last_modified": nullableStringToInterface(lastModified),
		"error_message": nullableStringToInterface(errorMessage),
		"created_at":    createdAt,
		"updated_at":    updatedAt,
	}, nil
}

// Placeholder implementations for other methods
func (s *SQLiteStorage) UpdateFeed(ctx context.Context, id int64, updates map[string]interface{}) error {
	return &StorageError{msg: "UpdateFeed not yet implemented for SQLite storage"}
}

func (s *SQLiteStorage) DeleteFeed(ctx context.Context, id int64) error {
	return &StorageError{msg: "DeleteFeed not yet implemented for SQLite storage"}
}

func (s *SQLiteStorage) ListFeeds(ctx context.Context, folderID *int64) ([]map[string]interface{}, error) {
	return nil, &StorageError{msg: "ListFeeds not yet implemented for SQLite storage"}
}

func (s *SQLiteStorage) CreateEntry(ctx context.Context, entry map[string]interface{}) error {
	return &StorageError{msg: "CreateEntry not yet implemented for SQLite storage"}
}

func (s *SQLiteStorage) GetEntry(ctx context.Context, id int64) (map[string]interface{}, error) {
	return nil, &StorageError{msg: "GetEntry not yet implemented for SQLite storage"}
}

func (s *SQLiteStorage) UpdateEntry(ctx context.Context, id int64, updates map[string]interface{}) error {
	return &StorageError{msg: "UpdateEntry not yet implemented for SQLite storage"}
}

func (s *SQLiteStorage) DeleteEntry(ctx context.Context, id int64) error {
	return &StorageError{msg: "DeleteEntry not yet implemented for SQLite storage"}
}

func (s *SQLiteStorage) ListEntries(ctx context.Context, filter map[string]interface{}) ([]map[string]interface{}, error) {
	return nil, &StorageError{msg: "ListEntries not yet implemented for SQLite storage"}
}

func (s *SQLiteStorage) CreateFolder(ctx context.Context, folder map[string]interface{}) (int64, error) {
	return 0, &StorageError{msg: "CreateFolder not yet implemented for SQLite storage"}
}

func (s *SQLiteStorage) GetFolder(ctx context.Context, id int64) (map[string]interface{}, error) {
	return nil, &StorageError{msg: "GetFolder not yet implemented for SQLite storage"}
}

func (s *SQLiteStorage) UpdateFolder(ctx context.Context, id int64, updates map[string]interface{}) error {
	return &StorageError{msg: "UpdateFolder not yet implemented for SQLite storage"}
}

func (s *SQLiteStorage) DeleteFolder(ctx context.Context, id int64) error {
	return &StorageError{msg: "DeleteFolder not yet implemented for SQLite storage"}
}

func (s *SQLiteStorage) ListFolders(ctx context.Context) ([]map[string]interface{}, error) {
	return nil, &StorageError{msg: "ListFolders not yet implemented for SQLite storage"}
}

func (s *SQLiteStorage) GetAISummary(ctx context.Context, entryID int64, language string) (string, error) {
	return "", &StorageError{msg: "GetAISummary not yet implemented for SQLite storage"}
}

func (s *SQLiteStorage) SetAISummary(ctx context.Context, entryID int64, language, summary string) error {
	return &StorageError{msg: "SetAISummary not yet implemented for SQLite storage"}
}

func (s *SQLiteStorage) DeleteAISummary(ctx context.Context, entryID int64, language string) error {
	return &StorageError{msg: "DeleteAISummary not yet implemented for SQLite storage"}
}

func (s *SQLiteStorage) GetAITranslation(ctx context.Context, entryID int64, language string) (string, error) {
	return "", &StorageError{msg: "GetAITranslation not yet implemented for SQLite storage"}
}

func (s *SQLiteStorage) SetAITranslation(ctx context.Context, entryID int64, language, content string) error {
	return &StorageError{msg: "SetAITranslation not yet implemented for SQLite storage"}
}

func (s *SQLiteStorage) DeleteAITranslation(ctx context.Context, entryID int64, language string) error {
	return &StorageError{msg: "DeleteAITranslation not yet implemented for SQLite storage"}
}

func (s *SQLiteStorage) GetRateLimit(ctx context.Context, host string) (int64, error) {
	return 0, &StorageError{msg: "GetRateLimit not yet implemented for SQLite storage"}
}

func (s *SQLiteStorage) SetRateLimit(ctx context.Context, host string, intervalSeconds int64) error {
	return &StorageError{msg: "SetRateLimit not yet implemented for SQLite storage"}
}

func (s *SQLiteStorage) DeleteRateLimit(ctx context.Context, host string) error {
	return &StorageError{msg: "DeleteRateLimit not yet implemented for SQLite storage"}
}

// Helper functions for type conversion
func getInt64Pointer(value interface{}) (*int64, bool) {
	if value == nil {
		return nil, true
	}
	if v, ok := value.(int64); ok {
		return &v, true
	}
	if v, ok := value.(float64); ok {
		i := int64(v)
		return &i, true
	}
	return nil, false
}

func getString(value interface{}) (string, bool) {
	if value == nil {
		return "", false
	}
	if v, ok := value.(string); ok {
		return v, true
	}
	return "", false
}

func getStringPointer(value interface{}) (*string, bool) {
	if value == nil {
		return nil, true
	}
	if v, ok := value.(string); ok {
		return &v, true
	}
	return nil, false
}

func nullableInt64ToInterface(ni sql.NullInt64) interface{} {
	if ni.Valid {
		return ni.Int64
	}
	return nil
}

func nullableStringToInterface(ns sql.NullString) interface{} {
	if ns.Valid {
		return ns.String
	}
	return nil
}