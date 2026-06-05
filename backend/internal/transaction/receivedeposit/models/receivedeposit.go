package models

import (
	"smlcloudplatform/internal/models"
	transmodels "smlcloudplatform/internal/transaction/models"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

const receivedepositCollectionName = "transactionreceivedeposit"

type ReceiveDeposit struct {
	models.PartitionIdentity `bson:"inline"`
	transmodels.Transaction  `bson:"inline"`
}
type ReceiveDepositInfo struct {
	models.DocIdentity `bson:"inline"`
	ReceiveDeposit     `bson:"inline"`
}

func (ReceiveDepositInfo) CollectionName() string {
	return receivedepositCollectionName
}

type ReceiveDepositData struct {
	models.HoldingCodeentity `bson:"inline"`
	ReceiveDepositInfo       `bson:"inline"`
}

type ReceiveDepositDoc struct {
	ID                 primitive.ObjectID `json:"id" bson:"id,omitempty"`
	ReceiveDepositData `bson:"inline"`
	models.ActivityDoc `bson:"inline"`
}

func (ReceiveDepositDoc) CollectionName() string {
	return receivedepositCollectionName
}

type ReceiveDepositItemGuid struct {
	DocNo string `json:"docno" bson:"docno"`
}

func (ReceiveDepositItemGuid) CollectionName() string {
	return receivedepositCollectionName
}

type ReceiveDepositActivity struct {
	ReceiveDepositData  `bson:"inline"`
	models.ActivityTime `bson:"inline"`
}

func (ReceiveDepositActivity) CollectionName() string {
	return receivedepositCollectionName
}

type ReceiveDepositDeleteActivity struct {
	models.Identity     `bson:"inline"`
	models.ActivityTime `bson:"inline"`
}

func (ReceiveDepositDeleteActivity) CollectionName() string {
	return receivedepositCollectionName
}
