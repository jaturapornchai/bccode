package models

import (
	"smlcloudplatform/internal/models"
	transmodels "smlcloudplatform/internal/transaction/models"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

const depositrefundCollectionName = "transactionDepositRefund"

type DepositRefund struct {
	models.PartitionIdentity `bson:"inline"`
	transmodels.Transaction  `bson:"inline"`
}
type DepositRefundInfo struct {
	models.DocIdentity `bson:"inline"`
	DepositRefund      `bson:"inline"`
}

func (DepositRefundInfo) CollectionName() string {
	return depositrefundCollectionName
}

type DepositRefundData struct {
	models.HoldingCodeentity `bson:"inline"`
	DepositRefundInfo        `bson:"inline"`
}

type DepositRefundDoc struct {
	ID                 primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	DepositRefundData  `bson:"inline"`
	models.ActivityDoc `bson:"inline"`
}

func (DepositRefundDoc) CollectionName() string {
	return depositrefundCollectionName
}

type DepositRefundItemGuid struct {
	DocNo string `json:"docno" bson:"docno"`
}

func (DepositRefundItemGuid) CollectionName() string {
	return depositrefundCollectionName
}

type DepositRefundActivity struct {
	DepositRefundData   `bson:"inline"`
	models.ActivityTime `bson:"inline"`
}

func (DepositRefundActivity) CollectionName() string {
	return depositrefundCollectionName
}

type DepositRefundDeleteActivity struct {
	models.Identity     `bson:"inline"`
	models.ActivityTime `bson:"inline"`
}

func (DepositRefundDeleteActivity) CollectionName() string {
	return depositrefundCollectionName
}
