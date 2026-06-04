package repositories

import (
	"context"
	"errors"
	"smlcloudplatform/internal/productimport/models"
	"smlcloudplatform/pkg/microservice"
)

type TaskStatusClickHouseRepository struct {
	pst microservice.IPersisterClickHouse
}

func NewTaskStatusClickHouseRepository(pst microservice.IPersisterClickHouse) ITaskStatusRepository {
	return &TaskStatusClickHouseRepository{
		pst: pst,
	}
}

func (repo *TaskStatusClickHouseRepository) Create(ctx context.Context, status models.TaskStatusModel) error {
	// สำหรับ Create ถ้า CompletedAt เป็น nil ควรตั้งเป็น nil ใน database
	return repo.pst.Create(ctx, &status)
}

func (repo *TaskStatusClickHouseRepository) Update(ctx context.Context, taskID string, status models.TaskStatusModel) error {
	// ✅ ดึง existing record ก่อนเพื่อใช้ created_at เดิม
	existing, err := repo.FindByTaskID(ctx, status.HoldingCode, taskID)
	if err != nil {
		// ถ้าไม่เจอ record เดิม ให้ใช้ created_at ใหม่
		status.CreatedAt = status.UpdatedAt
	} else {
		// ใช้ created_at จาก record เดิม
		status.CreatedAt = existing.CreatedAt
	}

	// เนื่องจาก ClickHouse ไม่อนุญาตให้ UPDATE column ที่อยู่ใน ORDER BY
	// จึงใช้ INSERT แทน เพื่อสร้าง record ใหม่
	return repo.pst.Create(ctx, &status)
}

func (repo *TaskStatusClickHouseRepository) FindByTaskID(ctx context.Context, holdingCode, taskID string) (models.TaskStatusModel, error) {
	results := []models.TaskStatusModel{}

	sqlExpr := "SELECT * FROM task_status WHERE holding_code = ? AND task_id = ? ORDER BY updated_at DESC LIMIT 1"
	err := repo.pst.Select(ctx, &results, sqlExpr, holdingCode, taskID)

	if err != nil {
		return models.TaskStatusModel{}, err
	}

	if len(results) == 0 {
		return models.TaskStatusModel{}, errors.New("task status not found")
	}

	return results[0], nil
}

func (repo *TaskStatusClickHouseRepository) Delete(ctx context.Context, holdingCode, taskID string) error {
	// ใช้ ALTER TABLE DELETE เหมือน ProductImport
	return repo.pst.Exec(ctx,
		"ALTER TABLE task_status DELETE WHERE holding_code = ? AND task_id = ?",
		holdingCode, taskID)
}
