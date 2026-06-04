package models

import (
	"smlcloudplatform/internal/models"
	transmodels "smlcloudplatform/internal/transaction/models"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

const paidadvancerefundCollectionName = "transactionPaidAdvanceRefund"

type PaidAdvanceRefund struct {
	models.PartitionIdentity `bson:"inline"`
	transmodels.Transaction  `bson:"inline"`
}
type PaidAdvanceRefundInfo struct {
	models.DocIdentity `bson:"inline"`
	PaidAdvanceRefund  `bson:"inline"`
}

func (PaidAdvanceRefundInfo) CollectionName() string {
	return paidadvancerefundCollectionName
}

type PaidAdvanceRefundData struct {
	models.HoldingCodeentity `bson:"inline"`
	PaidAdvanceRefundInfo    `bson:"inline"`
}

type PaidAdvanceRefundDoc struct {
	ID                    primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	PaidAdvanceRefundData `bson:"inline"`
	models.ActivityDoc    `bson:"inline"`
}

func (PaidAdvanceRefundDoc) CollectionName() string {
	return paidadvancerefundCollectionName
}

type PaidAdvanceRefundItemGuid struct {
	DocNo string `json:"docno" bson:"docno"`
}

func (PaidAdvanceRefundItemGuid) CollectionName() string {
	return paidadvancerefundCollectionName
}

type PaidAdvanceRefundActivity struct {
	PaidAdvanceRefundData `bson:"inline"`
	models.ActivityTime   `bson:"inline"`
}

func (PaidAdvanceRefundActivity) CollectionName() string {
	return paidadvancerefundCollectionName
}

type PaidAdvanceRefundDeleteActivity struct {
	models.Identity     `bson:"inline"`
	models.ActivityTime `bson:"inline"`
}

func (PaidAdvanceRefundDeleteActivity) CollectionName() string {
	return paidadvancerefundCollectionName
}
