package models

import (
	"mime/multipart"
	"smlcloudplatform/internal/models"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

const slipimageCollectionName = "slipImages"

type SlipImageRequest struct {
	Mode            uint8                 `json:"mode" bson:"mode"` // 0 = slip, 1 = qr
	File            *multipart.FileHeader `json:"file" bson:"file"`
	DocNo           string                `json:"docno" bson:"docno"`
	DocDate         time.Time             `json:"docdate" bson:"docdate"`
	PosID           string                `json:"posid" bson:"posid"`
	MachineCode     string                `json:"machinecode" bson:"machinecode"`
	BranchCode      string                `json:"branchcode" bson:"branchcode"`
	ZoneGroupNumber string                `json:"zonegroupnumber" bson:"zonegroupnumber"`
}

type SlipImage struct {
	Mode            uint8     `json:"mode" bson:"mode"` // 0 = slip, 1 = qr
	URI             string    `json:"uri" bson:"uri"`
	Size            int64     `json:"size" bson:"size"`
	DocNo           string    `json:"docno" bson:"docno"`
	DocDate         time.Time `json:"docdate" bson:"docdate"`
	PosID           string    `json:"posid" bson:"posid"`
	MachineCode     string    `json:"machinecode" bson:"machinecode"`
	BranchCode      string    `json:"branchcode" bson:"branchcode"`
	ZoneGroupNumber string    `json:"zonegroupnumber" bson:"zonegroupnumber"`
}

type SlipImageInfo struct {
	models.DocIdentity `bson:"inline"`
	SlipImage          `bson:"inline"`
}

func (SlipImageInfo) CollectionName() string {
	return slipimageCollectionName
}

type SlipImageData struct {
	models.ShopIdentity `bson:"inline"`
	SlipImageInfo       `bson:"inline"`
}

type SlipImageDoc struct {
	ID                 primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	SlipImageData      `bson:"inline"`
	models.ActivityDoc `bson:"inline"`
}

func (SlipImageDoc) CollectionName() string {
	return slipimageCollectionName
}

type SlipImageItemGuid struct {
}

func (SlipImageItemGuid) CollectionName() string {
	return slipimageCollectionName
}

type SlipImageActivity struct {
	SlipImageData       `bson:"inline"`
	models.ActivityTime `bson:"inline"`
}

func (SlipImageActivity) CollectionName() string {
	return slipimageCollectionName
}

type SlipImageDeleteActivity struct {
	models.Identity     `bson:"inline"`
	models.ActivityTime `bson:"inline"`
}

func (SlipImageDeleteActivity) CollectionName() string {
	return slipimageCollectionName
}
