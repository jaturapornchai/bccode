package models

import "time"

type TaskStatusModel struct {
	TaskID      string     `json:"taskid" ch:"taskid"`
	HoldingCode string     `json:"holdingcode" ch:"holdingcode"`
	Status      string     `json:"status" ch:"status"`
	ErrorMsg    string     `json:"errormessage,omitempty" ch:"errormessage"`
	Progress    int32      `json:"progress" ch:"progress"`
	CreatedAt   time.Time  `json:"createdat" ch:"createdat"`
	UpdatedAt   time.Time  `json:"updatedat" ch:"updatedat"`
	CompletedAt *time.Time `json:"completedat,omitempty" ch:"completedat"`
}

func (TaskStatusModel) TableName() string {
	return "task_status"
}

const (
	TaskStatusModelProcessing = "processing"
	TaskStatusModelCompleted  = "completed"
	TaskStatusModelFailed     = "failed"
)
