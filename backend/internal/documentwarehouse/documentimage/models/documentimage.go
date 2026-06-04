package models

import (
	"smlcloudplatform/internal/models"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

const documentImageCollectionName = "documentImages"

type DocumentImage struct {
	ImageURI string `json:"imageuri" bson:"imageuri"`
	Name     string `json:"name" bson:"name"`
	// IsReject        bool             `json:"isreject" bson:"isreject"`
	// Status          int8             `json:"status" bson:"status"`
	References      []Reference      `json:"references" bson:"references"`
	ReferenceGroups []ReferenceGroup `json:"referencegroups" bson:"referencegroups"`
	BillCount       float64          `json:"billcount" bson:"billcount"`

	UploadedBy     string      `json:"uploaded_by" bson:"uploaded_by"`
	UploadedAt     time.Time   `json:"uploaded_at" bson:"uploaded_at"`
	MetaFileAt     time.Time   `json:"meta_file_at" bson:"meta_file_at"`
	CloneImageFrom string      `json:"cloneimagefrom" bson:"cloneimagefrom"`
	Edits          []ImageEdit `json:"edits" bson:"edits"`
	Comments       []Comment   `json:"comments" bson:"comments"`
}

type ReferenceGroup struct {
	GroupType  string `json:"group_type" bson:"group_type"`
	ParentGUID string `json:"parent_guid" bson:"parent_guid"`
	XOrder     int    `json:"xorder" bson:"xorder"`
	XType      int    `json:"xtype" bson:"xtype"`
}

type Reference struct {
	GuidFixed string `json:"guid_fixed" bson:"guid_fixed"`
	Module    string `json:"module" bson:"module"`
	DocNo     string `json:"docno" bson:"docno" `
}

type Comment struct {
	GuidFixed   string    `json:"guid_fixed" bson:"guid_fixed"`
	Comment     string    `json:"comment" bson:"comment"`
	CommentedAt time.Time `json:"commented_at" bson:"commented_at"`
	CommentedBy string    `json:"commented_by" bson:"commented_by"`
}

type CommentRequest struct {
	Comment string `json:"comment" bson:"comment"`
}

type ImageEdit struct {
	ImageURI string    `json:"imageuri" bson:"imageuri"`
	EditedBy string    `json:"edited_by" bson:"edited_by"`
	EditedAt time.Time `json:"edited_at" bson:"edited_at"`
}

type ImageEditRequest struct {
	ImageURI string    `json:"imageuri" bson:"imageuri"`
	EditedBy string    `json:"edited_by" bson:"edited_by"`
	EditedAt time.Time `json:"edited_at" bson:"edited_at"`
}

type DocumentImageRequest struct {
	DocumentImage          `bson:"inline"`
	DocumentImageGroupGUID string    `json:"document_image_group_guid" bson:"document_image_group_guid"`
	Tags                   *[]string `json:"tags,omitempty" bson:"tags,omitempty"`
	TaskGUID               string    `json:"task_guid" bson:"task_guid" validate:"required,min=1"`
	PathTask               string    `json:"pathtask" bson:"pathtask"`
}

type DocumentImageInfo struct {
	models.DocIdentity `bson:"inline"`
	DocumentImage      `bson:"inline"`
}

func (DocumentImageInfo) CollectionName() string {
	return documentImageCollectionName
}

type DocumentImageData struct {
	models.HoldingCodeentity `bson:"inline"`
	DocumentImageInfo        `bson:"inline"`
}

type DocumentImageDoc struct {
	ID                 primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	DocumentImageData  `bson:"inline"`
	models.ActivityDoc `bson:"inline"`
	models.LastUpdate  `bson:"inline"`
}

func (DocumentImageDoc) CollectionName() string {
	return documentImageCollectionName
}

type DocumentImageItemGuid struct {
	DocumentImageGuid string `json:"category_guid" bson:"category_guid" gorm:"category_guid"`
}

func (DocumentImageItemGuid) CollectionName() string {
	return documentImageCollectionName
}

type DocumentImageInfoResponse struct {
	Success bool              `json:"success"`
	Data    DocumentImageInfo `json:"data,omitempty"`
}

type DocumentImagePageResponse struct {
	Success    bool                          `json:"success"`
	Data       []DocumentImageInfo           `json:"data,omitempty"`
	Pagination models.PaginationDataResponse `json:"pagination,omitempty"`
}

type RequestDocumentImageReject struct {
	IsReject bool `json:"isreject" bson:"isreject"`
}

type DocumentImageStatus struct {
	DocGUIDRef string `json:"doc_guid_ref" bson:"doc_guid_ref"`
	Status     int8   `json:"status" bson:"status"`
}

type ImageStatus = int8
