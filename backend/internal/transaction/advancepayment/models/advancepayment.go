package models

import (
	"smlcloudplatform/internal/models"
	transmodels "smlcloudplatform/internal/transaction/models"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

const advancepaymentCollectionName = "transactionAdvancePayment"

type AdvancePayment struct {
	models.PartitionIdentity `bson:"inline"`
	transmodels.Transaction  `bson:"inline"`
}
type AdvancePaymentInfo struct {
	models.DocIdentity `bson:"inline"`
	AdvancePayment     `bson:"inline"`
}

func (AdvancePaymentInfo) CollectionName() string {
	return advancepaymentCollectionName
}

type AdvancePaymentData struct {
	models.HoldingCodeentity `bson:"inline"`
	AdvancePaymentInfo       `bson:"inline"`
}

type AdvancePaymentDoc struct {
	ID                 primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	AdvancePaymentData `bson:"inline"`
	models.ActivityDoc `bson:"inline"`
}

func (AdvancePaymentDoc) CollectionName() string {
	return advancepaymentCollectionName
}

type AdvancePaymentItemGuid struct {
	DocNo string `json:"docno" bson:"docno"`
}

func (AdvancePaymentItemGuid) CollectionName() string {
	return advancepaymentCollectionName
}

type AdvancePaymentActivity struct {
	AdvancePaymentData  `bson:"inline"`
	models.ActivityTime `bson:"inline"`
}

func (AdvancePaymentActivity) CollectionName() string {
	return advancepaymentCollectionName
}

type AdvancePaymentDeleteActivity struct {
	models.Identity     `bson:"inline"`
	models.ActivityTime `bson:"inline"`
}

func (AdvancePaymentDeleteActivity) CollectionName() string {
	return advancepaymentCollectionName
}
