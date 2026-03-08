package models

import (
	"smlcloudplatform/internal/models"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

const paymentmasterCollectionName = "paymentMaster"

type PaymentMaster struct {
	models.PartitionIdentity `bson:"inline"`
	PaymentCode              string  `json:"paymentcode" bson:"paymentcode"`
	CountryCode              string  `json:"countrycode" bson:"countrycode"`
	PaymentLogo              string  `json:"paymentlogo" bson:"paymentlogo"`
	PaymentType              int8    `json:"paymenttype" bson:"paymenttype"`
	FeeRate                  float64 `json:"feerate" bson:"feerate"`
	WalletPayType            int16   `json:"wallettype" bson:"wallettype"`
	models.Name              `bson:"inline"`
}

type PaymentMasterInfo struct {
	models.DocIdentity `bson:"inline"`
	PaymentMaster      `bson:"inline"`
}

func (PaymentMasterInfo) CollectionName() string {
	return paymentmasterCollectionName
}

type PaymentMasterData struct {
	models.ShopIdentity `bson:"inline"`
	PaymentMasterInfo   `bson:"inline"`
}

type PaymentMasterDoc struct {
	ID                 primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	PaymentMasterData  `bson:"inline"`
	models.ActivityDoc `bson:"inline"`
}

func (PaymentMasterDoc) CollectionName() string {
	return paymentmasterCollectionName
}

type PaymentMasterItemGuid struct {
	PaymentCode string `json:"paymentcode" bson:"paymentcode" gorm:"paymentcode"`
}

func (PaymentMasterItemGuid) CollectionName() string {
	return paymentmasterCollectionName
}

type PaymentMasterActivity struct {
	PaymentMasterData   `bson:"inline"`
	models.ActivityTime `bson:"inline"`
}

func (PaymentMasterActivity) CollectionName() string {
	return paymentmasterCollectionName
}

type PaymentMasterDeleteActivity struct {
	models.Identity     `bson:"inline"`
	models.ActivityTime `bson:"inline"`
}

func (PaymentMasterDeleteActivity) CollectionName() string {
	return paymentmasterCollectionName
}

type PaymentMasterInfoResponse struct {
	Success bool              `json:"success"`
	Data    PaymentMasterInfo `json:"data,omitempty"`
}

type PaymentMasterPageResponse struct {
	Success    bool                          `json:"success"`
	Data       []PaymentMasterInfo           `json:"data,omitempty"`
	Pagination models.PaginationDataResponse `json:"pagination,omitempty"`
}
