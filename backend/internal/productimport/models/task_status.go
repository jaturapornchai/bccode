package models

import "time"

type TaskStatusModel struct {
	TaskID      string     `json:"task_id" ch:"task_id"`
	ShopID      string     `json:"shop_id" ch:"shop_id"`
	Status      string     `json:"status" ch:"status"`
	ErrorMsg    string     `json:"error_message,omitempty" ch:"error_message"`
	Progress    int32      `json:"progress" ch:"progress"`
	CreatedAt   time.Time  `json:"created_at" ch:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at" ch:"updated_at"`
	CompletedAt *time.Time `json:"completed_at,omitempty" ch:"completed_at"`
}

func (TaskStatusModel) TableName() string {
	return "task_status"
}

const (
	TaskStatusModelProcessing = "processing"
	TaskStatusModelCompleted  = "completed"
	TaskStatusModelFailed     = "failed"
)
