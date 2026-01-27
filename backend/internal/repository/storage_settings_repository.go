package repository

import (
	"context"
	"fmt"
	"strings"
	"time"

	"gist/backend/internal/db"
	"gist/backend/internal/model"
)

type StorageSettingsRepository struct {
	storage db.Storage
}

func NewStorageSettingsRepository(storage db.Storage) *StorageSettingsRepository {
	return &StorageSettingsRepository{storage: storage}
}

func (r *StorageSettingsRepository) Get(ctx context.Context, key string) (*model.Setting, error) {
	value, err := r.storage.GetSetting(ctx, key)
	if err != nil {
		if isNotFoundError(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("get setting: %w", err)
	}
	if value == "" {
		return nil, nil
	}
	timestamp := time.Now().UTC()
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

func (r *StorageSettingsRepository) GetByPrefix(ctx context.Context, prefix string) ([]model.Setting, error) {
	return nil, fmt.Errorf("get by prefix not implemented for storage-based repository")
}

func (r *StorageSettingsRepository) DeleteByPrefix(ctx context.Context, prefix string) (int64, error) {
	return 0, fmt.Errorf("delete by prefix not implemented for storage-based repository")
}

func isNotFoundError(err error) bool {
	errMsg := err.Error()
	if strings.Contains(errMsg, "setting not found") {
		return true
	}
	if strings.Contains(errMsg, "redis: nil") {
		return true
	}
	return false
}