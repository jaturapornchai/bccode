package models

import (
	"smlcloudplatform/internal/models"
	transmodels "smlcloudplatform/internal/transaction/models"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

const advancepaymentrefundCollectionName = "transactionAdvancePaymentRefund"

type AdvancePaymentRefund struct {
	models.PartitionIdentity `bson:"inline"`
	transmodels.Transaction  `bson:"inline"`
}
type AdvancePaymentRefundInfo struct {
	models.DocIdentity   `bson:"inline"`
	AdvancePaymentRefund `bson:"inline"`
}

func (AdvancePaymentRefundInfo) CollectionName() string {
	return advancepaymentrefundCollectionName
}

type AdvancePaymentRefundData struct {
	models.HoldingCodeentity `bson:"inline"`
	AdvancePaymentRefundInfo `bson:"inline"`
}

type AdvancePaymentRefundDoc struct {
	ID                       primitive.ObjectID `json:"id" bson:"id,omitempty"`
	AdvancePaymentRefundData `bson:"inline"`
	models.ActivityDoc       `bson:"inline"`
}

func (AdvancePaymentRefundDoc) CollectionName() string {
	return advancepaymentrefundCollectionName
}

type AdvancePaymentRefundItemGuid struct {
	DocNo string `json:"docno" bson:"docno"`
}

func (AdvancePaymentRefundItemGuid) CollectionName() string {
	return advancepaymentrefundCollectionName
}

type AdvancePaymentRefundActivity struct {
	AdvancePaymentRefundData `bson:"inline"`
	models.ActivityTime      `bson:"inline"`
}

func (AdvancePaymentRefundActivity) CollectionName() string {
	return advancepaymentrefundCollectionName
}

type AdvancePaymentRefundDeleteActivity struct {
	models.Identity     `bson:"inline"`
	models.ActivityTime `bson:"inline"`
}

func (AdvancePaymentRefundDeleteActivity) CollectionName() string {
	return advancepaymentrefundCollectionName
}
