package repository

import (
	"context"
	"fmt"
	"time"

	"gist/backend/internal/db"
	"gist/backend/internal/model"
)

// StorageSettingsRepository implements SettingsRepository using the storage abstraction
// This allows it to work with both SQLite and Redis backends
type StorageSettingsRepository struct {
	storage db.Storage
}

func NewStorageSettingsRepository(storage db.Storage) *StorageSettingsRepository {
	return &StorageSettingsRepository{storage: storage}
}

func (r *StorageSettingsRepository) Get(ctx context.Context, key string) (*model.Setting, error) {
	value, err := r.storage.GetSetting(ctx, key)
	if err != nil {
		return nil, fmt.Errorf("get setting: %w", err)
	}

	timestamp := time.Now().UTC().Format(time.RFC3339)
	return &model.Setting{
		Key:       key,
		Value:     value,
		UpdatedAt: timestamp,
	}, nil
}

func (r *StorageSettingsRepository) Set(ctx context.Context, key, value string) error {
	return r.storage.SetSetting(ctx, key, value)
}

func (r *StorageSettingsRepository) Delete(ctx context.Context, key string) error {
	return r.storage.DeleteSetting(ctx, key)
}

func (r *StorageSettingsRepository) GetAll(ctx context.Context) ([]model.Setting, error) {
	// This is a simplified implementation
	// In a real implementation, you'd need to scan for all settings
	return nil, fmt.Errorf("GetAll not implemented for storage-based repository")
}