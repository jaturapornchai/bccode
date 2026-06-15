package models

import (
	"smlcloudplatform/internal/models"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

const documentImageCollectionName = "documentimages"

type DocumentImage struct {
	ImageURI string `json:"imageuri" bson:"imageuri"`
	Name     string `json:"name" bson:"name"`
	// IsReject        bool             `json:"isreject" bson:"isreject"`
	// Status          int8             `json:"status" bson:"status"`
	References      []Reference      `json:"references" bson:"references"`
	ReferenceGroups []ReferenceGroup `json:"referencegroups" bson:"referencegroups"`
	BillCount       float64          `json:"billcount" bson:"billcount"`

	UploadedBy     string      `json:"uploadedby" bson:"uploadedby"`
	UploadedAt     time.Time   `json:"uploadedat" bson:"uploadedat"`
	MetaFileAt     time.Time   `json:"metafileat" bson:"metafileat"`
	CloneImageFrom string      `json:"cloneimagefrom" bson:"cloneimagefrom"`
	Edits          []ImageEdit `json:"edits" bson:"edits"`
	Comments       []Comment   `json:"comments" bson:"comments"`
}

type ReferenceGroup struct {
	GroupType  string `json:"grouptype" bson:"grouptype"`
	ParentGUID string `json:"parentguid" bson:"parentguid"`
	XOrder     int    `json:"xorder" bson:"xorder"`
	XType      int    `json:"xtype" bson:"xtype"`
}

type Reference struct {
	GuidFixed string `json:"guidfixed" bson:"guidfixed"`
	Module    string `json:"module" bson:"module"`
	DocNo     string `json:"docno" bson:"docno" `
}

type Comment struct {
	GuidFixed   string    `json:"guidfixed" bson:"guidfixed"`
	Comment     string    `json:"comment" bson:"comment"`
	CommentedAt time.Time `json:"commentedat" bson:"commentedat"`
	CommentedBy string    `json:"commentedby" bson:"commentedby"`
}

type CommentRequest struct {
	Comment string `json:"comment" bson:"comment"`
}

type ImageEdit struct {
	ImageURI string    `json:"imageuri" bson:"imageuri"`
	EditedBy string    `json:"editedby" bson:"editedby"`
	EditedAt time.Time `json:"editedat" bson:"editedat"`
}

type ImageEditRequest struct {
	ImageURI string    `json:"imageuri" bson:"imageuri"`
	EditedBy string    `json:"editedby" bson:"editedby"`
	EditedAt time.Time `json:"editedat" bson:"editedat"`
}

type DocumentImageRequest struct {
	DocumentImage          `bson:"inline"`
	DocumentImageGroupGUID string    `json:"documentimagegroupguid" bson:"documentimagegroupguid"`
	Tags                   *[]string `json:"tags,omitempty" bson:"tags,omitempty"`
	TaskGUID               string    `json:"taskguid" bson:"taskguid" validate:"required,min=1"`
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
	DocumentImageGuid string `json:"categoryguid" bson:"categoryguid" gorm:"categoryguid"`
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
	DocGUIDRef string `json:"docguidref" bson:"docguidref"`
	Status     int8   `json:"status" bson:"status"`
}

type ImageStatus = int8
