package repositories

import (
	"context"
	"smlcloudplatform/internal/productimport/models"
)

type ITaskStatusRepository interface {
	Create(ctx context.Context, status models.TaskStatusModel) error
	Update(ctx context.Context, taskID string, status models.TaskStatusModel) error
	FindByTaskID(ctx context.Context, shopID, taskID string) (models.TaskStatusModel, error)
	Delete(ctx context.Context, shopID, taskID string) error
}
