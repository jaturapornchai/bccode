package models

import (
	"smlcloudplatform/internal/models"
	transmodels "smlcloudplatform/internal/transaction/models"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

const paidadvanceCollectionName = "transactionpaidadvance"

type PaidAdvance struct {
	models.PartitionIdentity `bson:"inline"`
	transmodels.Transaction  `bson:"inline"`
}
type PaidAdvanceInfo struct {
	models.DocIdentity `bson:"inline"`
	PaidAdvance        `bson:"inline"`
}

func (PaidAdvanceInfo) CollectionName() string {
	return paidadvanceCollectionName
}

type PaidAdvanceData struct {
	models.HoldingCodeentity `bson:"inline"`
	PaidAdvanceInfo          `bson:"inline"`
}

type PaidAdvanceDoc struct {
	ID                 primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	PaidAdvanceData    `bson:"inline"`
	models.ActivityDoc `bson:"inline"`
}

func (PaidAdvanceDoc) CollectionName() string {
	return paidadvanceCollectionName
}

type PaidAdvanceItemGuid struct {
	DocNo string `json:"docno" bson:"docno"`
}

func (PaidAdvanceItemGuid) CollectionName() string {
	return paidadvanceCollectionName
}

type PaidAdvanceActivity struct {
	PaidAdvanceData     `bson:"inline"`
	models.ActivityTime `bson:"inline"`
}

func (PaidAdvanceActivity) CollectionName() string {
	return paidadvanceCollectionName
}

type PaidAdvanceDeleteActivity struct {
	models.Identity     `bson:"inline"`
	models.ActivityTime `bson:"inline"`
}

func (PaidAdvanceDeleteActivity) CollectionName() string {
	return paidadvanceCollectionName
}
