package models

import (
	"smlcloudplatform/internal/models"
	transmodels "smlcloudplatform/internal/transaction/models"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

const saleorderCollectionName = "transactionpickandpack"

type Pickandpack struct {
	models.PartitionIdentity `bson:"inline"`
	transmodels.Transaction  `bson:"inline"`
	WhCode                   string          `json:"whcode" bson:"whcode"`
	WhNames                  *[]models.NameX `json:"whnames" bson:"whnames"`
	LocationCode             string          `json:"locationcode" bson:"locationcode"`
	LocationNames            *[]models.NameX `json:"locationnames" bson:"locationnames"`
	Sendtype                 string          `json:"sendtype" bson:"sendtype"`
	Email                    string          `json:"email" bson:"email"`
	Phone                    string          `json:"phone" bson:"phone"`
	Address                  string          `json:"address" bson:"address"`
	RefSaleInvoice           string          `json:"refsaleinvoice" bson:"refsaleinvoice"`
	PackStatus               int8            `json:"packstatus" bson:"packstatus"`

	// Print tracking fields
	IsPrint bool       `json:"isprint" bson:"isprint"`
	PrintAt *time.Time `json:"printat,omitempty" bson:"printat,omitempty"`
	PrintBy string     `json:"printby" bson:"printby"`

	// Confirm tracking fields
	IsConfirm bool       `json:"isconfirm" bson:"isconfirm"`
	ConfirmAt *time.Time `json:"confirmat,omitempty" bson:"confirmat,omitempty"`
	ConfirmBy string     `json:"confirmby" bson:"confirmby"`
}

type PickandpackInfo struct {
	models.DocIdentity `bson:"inline"`
	Pickandpack        `bson:"inline"`
}

func (PickandpackInfo) CollectionName() string {
	return saleorderCollectionName
}

type PickandpackData struct {
	models.HoldingCodeentity `bson:"inline"`
	PickandpackInfo          `bson:"inline"`
}

type PickandpackDoc struct {
	ID                 primitive.ObjectID `json:"id" bson:"id,omitempty"`
	PickandpackData    `bson:"inline"`
	models.ActivityDoc `bson:"inline"`
}

func (PickandpackDoc) CollectionName() string {
	return saleorderCollectionName
}

type PickandpackItemGuid struct {
	DocNo string `json:"docno" bson:"docno"`
}

func (PickandpackItemGuid) CollectionName() string {
	return saleorderCollectionName
}

type PickandpackActivity struct {
	PickandpackData     `bson:"inline"`
	models.ActivityTime `bson:"inline"`
}

func (PickandpackActivity) CollectionName() string {
	return saleorderCollectionName
}

type PickandpackDeleteActivity struct {
	models.Identity     `bson:"inline"`
	models.ActivityTime `bson:"inline"`
}

func (PickandpackDeleteActivity) CollectionName() string {
	return saleorderCollectionName
}

type WarehouseLocationGroup struct {
	WhCode       string `json:"whcode" bson:"whcode"`
	LocationCode string `json:"locationcode" bson:"locationcode"`
}

type WarehouseStatusCount struct {
	PackStatus int `json:"packstatus" bson:"packstatus"`
	Count      int `json:"count" bson:"count"`
}

type PickandpackWarehouseDashboard struct {
	models.PartitionIdentity `bson:"inline"`
	ID                       WarehouseLocationGroup `json:"id" bson:"id"`
	WhNames                  *[]models.NameX        `json:"whnames" bson:"whnames"`
	LocationNames            *[]models.NameX        `json:"locationnames" bson:"locationnames"`
	StatusCounts             []WarehouseStatusCount `json:"statuscounts" bson:"statuscounts"`
	TotalCount               int32                  `json:"totalcount" bson:"totalcount"`
}

func (PickandpackWarehouseDashboard) CollectionName() string {
	return saleorderCollectionName
}
