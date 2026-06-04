package models

import (
	"smlcloudplatform/internal/models"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

const documentImageGroupCollectionName = "documentImageGroups"

const (
	IMAGE_PENDING = iota
	IMAGE_CHECKED
	IMAGE_REJECT
	IMAGE_BANNED
	IMAGE_REJECT_KEYING
	IMAGE_GL_COMPLETED
	IMAGE_FROM_REJECT
)

type DocumentImageGroup struct {
	// DocumentRef     string            `json:"documentref" bson:"documentref"`
	Title               string            `json:"title" bson:"title"`
	OcrAnalyzeAi        string            `json:"ocranalyzeai" bson:"ocranalyzeai"`
	BillCount           float64           `json:"billcount" bson:"billcount"`
	References          []Reference       `json:"references" bson:"references"`
	Tags                *[]string         `json:"tags,omitempty" bson:"tags,omitempty"`
	ImageReferences     *[]ImageReference `json:"imagereferences" bson:"imagereferences"`
	UploadedBy          string            `json:"uploaded_by" bson:"uploaded_by"`
	UploadedAt          time.Time         `json:"uploaded_at" bson:"uploaded_at"`
	Status              int8              `json:"status" bson:"status"`
	Description         string            `json:"description" bson:"description"`
	TaskGUID            string            `json:"task_guid" bson:"task_guid" validate:"required,min=1"`
	PathTask            string            `json:"pathtask" bson:"pathtask"`
	IsTaskCompleted     bool              `json:"iscompleted" bson:"iscompleted"`
	RejectFromGroupGUID string            `json:"reject_from_group_guid" bson:"reject_from_group_guid"`
	XOrder              int               `json:"xorder" bson:"xorder"`
	RejectRemark        string            `json:"rejectremark" bson:"rejectremark"`
	StatusChangedBy     string            `json:"status_changed_by" bson:"status_changed_by"`
	StatusChangedAt     time.Time         `json:"status_changed_at" bson:"status_changed_at"`
	StatusHistories     []StatusHistory   `json:"statushistories" bson:"statushistories"`
}

type StatusHistory struct {
	Status    int8      `json:"status" bson:"status"`
	ChangedBy string    `json:"changed_by" bson:"changed_by"`
	ChangedAt time.Time `json:"changed_at" bson:"changed_at"`
}

type DocumentImageGroupBody struct {
	DocumentImageGroup `bson:"inline"`
	ImageReferences    *[]ImageReferenceBody `json:"imagereferences,omitempty" bson:"imagereferences,omitempty"`
}

type ImageReferenceBody struct {
	XOrder            int    `json:"xorder" bson:"xorder"`
	DocumentImageGUID string `json:"document_image_guid" bson:"document_image_guid"`
}
type ImageReference struct {
	ImageReferenceBody `bson:",inline"`
	ImageURI           string  `json:"imageuri" bson:"imageuri"`
	BillCount          float64 `json:"billcount" bson:"billcount"`
	CloneImageFrom     string  `json:"cloneimagefrom" bson:"cloneimagefrom"`
	// ImageEditURI       string `json:"imageedituri" bson:"imageedituri"`
	Name string `json:"name" bson:"name"`
	// IsReject           bool      `json:"isreject" bson:"isreject"`
	UploadedBy string    `json:"uploaded_by" bson:"uploaded_by"`
	UploadedAt time.Time `json:"uploaded_at" bson:"uploaded_at"`
	MetaFileAt time.Time `json:"meta_file_at" bson:"meta_file_at"`
}

func (DocumentImageGroup) CollectionName() string {
	return documentImageGroupCollectionName
}

type DocumentImageGroupInfo struct {
	models.DocIdentity `bson:"inline"`
	DocumentImageGroup `bson:"inline"`
}

func (DocumentImageGroupInfo) CollectionName() string {
	return documentImageGroupCollectionName
}

type DocumentImageGroupData struct {
	models.HoldingCodeentity `bson:"inline"`
	DocumentImageGroupInfo   `bson:"inline"`
}

type DocumentImageGroupDoc struct {
	ID                     primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	DocumentImageGroupData `bson:"inline"`
	models.ActivityDoc     `bson:"inline"`
	models.LastUpdate      `bson:"inline"`
}

func (DocumentImageGroupDoc) CollectionName() string {
	return documentImageGroupCollectionName
}

type Status struct {
	Status int8 `json:"status"`
}

type XSortDocumentImageGroupRequest struct {
	GUIDFixed string `json:"guid_fixed" bson:"guid_fixed" validate:"required,min=1"`
	XOrder    uint   `json:"xorder" bson:"xorder" validate:"min=0,max=4294967295"`
}

type DocumentImageGroupStatus struct {
	models.HoldingCodeentity `bson:"inline"`
	models.DocIdentity       `bson:"inline"`
	Status                   int8        `json:"status"`
	BillCount                float64     `json:"billcount" bson:"billcount"`
	References               []Reference `json:"references" bson:"references"`
}

func (DocumentImageGroupStatus) CollectionName() string {
	return documentImageGroupCollectionName
}
