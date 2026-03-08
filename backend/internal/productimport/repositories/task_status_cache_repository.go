package repositories

import (
	"context"
	"encoding/json"
	"fmt"
	"smlcloudplatform/internal/productimport/models"
	"smlcloudplatform/pkg/microservice"
	"time"
)

type TaskStatusCacheRepository struct {
	cache microservice.ICacher
}

func NewTaskStatusCacheRepository(cache microservice.ICacher) ITaskStatusRepository {
	return &TaskStatusCacheRepository{
		cache: cache,
	}
}

func (r *TaskStatusCacheRepository) getCacheKey(shopID, taskID string) string {
	return fmt.Sprintf("task_status:%s:%s", shopID, taskID)
}

func (r *TaskStatusCacheRepository) Create(ctx context.Context, status models.TaskStatusModel) error {
	key := r.getCacheKey(status.ShopID, status.TaskID)

	data, err := json.Marshal(status)
	if err != nil {
		return err
	}

	// Set with 24 hour expiry
	return r.cache.Set(key, string(data), time.Hour*24)
}

func (r *TaskStatusCacheRepository) Update(ctx context.Context, taskID string, status models.TaskStatusModel) error {
	key := r.getCacheKey(status.ShopID, taskID)

	data, err := json.Marshal(status)
	if err != nil {
		return err
	}

	// Set with 24 hour expiry
	return r.cache.Set(key, string(data), time.Hour*24)
}

func (r *TaskStatusCacheRepository) FindByTaskID(ctx context.Context, shopID, taskID string) (models.TaskStatusModel, error) {
	key := r.getCacheKey(shopID, taskID)

	data, err := r.cache.Get(key)
	if err != nil {
		return models.TaskStatusModel{}, fmt.Errorf("task status not found: %w", err)
	}

	var status models.TaskStatusModel
	err = json.Unmarshal([]byte(data), &status)
	if err != nil {
		return models.TaskStatusModel{}, fmt.Errorf("failed to unmarshal task status: %w", err)
	}

	return status, nil
}

func (r *TaskStatusCacheRepository) Delete(ctx context.Context, shopID, taskID string) error {
	key := r.getCacheKey(shopID, taskID)
	return r.cache.Del(key)
}
