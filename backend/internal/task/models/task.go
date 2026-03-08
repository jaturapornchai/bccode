package models

import (
	"smlcloudplatform/internal/models"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

const taskCollectionName = "tasks"

const (
	TaskPending = iota
	TaskUplaoded
	TaskChecking
	TaskCompleted
	TaskGlCompleted
	TaskCanceled
)

type Task struct {
	models.PartitionIdentity `bson:"inline"`
	Code                     string         `json:"code" bson:"code"`
	Name                     string         `json:"name" bson:"name"`
	Module                   string         `json:"module" bson:"module"`
	Status                   int8           `json:"status" bson:"status"`
	ParentGUIDFixed          string         `json:"parentguidfixed" bson:"parentguidfixed"`
	Path                     string         `json:"path" bson:"path"`
	IsFavorit                bool           `json:"isfavorit" bson:"isfavorit"`
	Tags                     *[]string      `json:"tags" bson:"tags"`
	Description              string         `json:"description" bson:"description"`
	TotalDocument            int            `json:"totaldocument" bson:"totaldocument"`
	TotalDocumentStatus      *[]TotalStatus `json:"totaldocumentstatus" bson:"totaldocumentstatus"`
	OwnerAt                  time.Time      `json:"ownerat" bson:"ownerat"`
	OwnerBy                  string         `json:"ownerby" bson:"ownerby"`
	BillCount                float64        `json:"billcount" bson:"billcount"`
	ReferenceCount           float64        `json:"referencecount" bson:"referencecount"`
	ReferenceBalance         float64        `json:"referencebalance" bson:"referencebalance"`
	RejectFromTaskGUID       string         `json:"rejectfromtaskguid" bson:"rejectfromtaskguid"`
	RejectedAt               time.Time      `json:"rejectedat,omitempty" bson:"rejectedat,omitempty"`
	RejectedBy               string         `json:"rejectedby,omitempty" bson:"rejectedby,omitempty"`
	
	// ToTalReject              int       `json:"totalreject" bson:"totalreject"`
}

type TotalStatus struct {
	Status int8 `json:"status" bson:"status"`
	Total  int  `json:"total" bson:"total"`
}

type TaskInfo struct {
	models.DocIdentity `bson:"inline"`
	Task               `bson:"inline"`
	TaskChild          TaskChild `json:"taskchild" bson:"taskchild"`
}

func (TaskInfo) CollectionName() string {
	return taskCollectionName
}

type TaskData struct {
	models.ShopIdentity `bson:"inline"`
	TaskInfo            `bson:"inline"`
}

type TaskDoc struct {
	ID                 primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	TaskData           `bson:"inline"`
	models.ActivityDoc `bson:"inline"`
}

func (TaskDoc) CollectionName() string {
	return taskCollectionName
}

type TaskItemGuid struct {
	Name string `json:"name" bson:"name"`
}

func (TaskItemGuid) CollectionName() string {
	return taskCollectionName
}

type TaskActivity struct {
	TaskData            `bson:"inline"`
	models.ActivityTime `bson:"inline"`
}

func (TaskActivity) CollectionName() string {
	return taskCollectionName
}

type TaskDeleteActivity struct {
	models.Identity     `bson:"inline"`
	models.ActivityTime `bson:"inline"`
}

func (TaskDeleteActivity) CollectionName() string {
	return taskCollectionName
}

type TaskStatus struct {
	Status int8 `json:"status"`
}

type TaskDocumentTotal struct {
	BillCount           float64        `json:"billcount" bson:"billcount"`
	ReferenceCount      float64        `json:"referencecount" bson:"referencecount"`
	ReferenceBalance    float64        `json:"referencebalance" bson:"referencebalance"`
	TotalDocument       int            `json:"totaldocument" bson:"totaldocument"`
	TotalDocumentStatus *[]TotalStatus `json:"totaldocumentstatus" bson:"totaldocumentstatus"`
}

func (TaskDocumentTotal) CollectionName() string {
	return taskCollectionName
}

type TaskTotalReject struct {
	ToTalReject int `json:"totalreject" bson:"totalreject"`
}

func (TaskTotalReject) CollectionName() string {
	return taskCollectionName
}

type TaskChild struct {
	GuidFixed string `json:"guidfixed" bson:"guidfixed"`
	Code      string `json:"code" bson:"code"`
	Name      string `json:"name" bson:"name"`
	Status    int8   `json:"status" bson:"status"`
}

func (TaskChild) CollectionName() string {
	return taskCollectionName
}
