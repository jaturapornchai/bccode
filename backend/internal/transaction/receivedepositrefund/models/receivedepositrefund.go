package models

import (
	"smlcloudplatform/internal/models"
	transmodels "smlcloudplatform/internal/transaction/models"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

const receivedepositrefundCollectionName = "transactionreceivedepositrefund"

type ReceiveDepositRefund struct {
	models.PartitionIdentity `bson:"inline"`
	transmodels.Transaction  `bson:"inline"`
}
type ReceiveDepositRefundInfo struct {
	models.DocIdentity   `bson:"inline"`
	ReceiveDepositRefund `bson:"inline"`
}

func (ReceiveDepositRefundInfo) CollectionName() string {
	return receivedepositrefundCollectionName
}

type ReceiveDepositRefundData struct {
	models.HoldingCodeentity `bson:"inline"`
	ReceiveDepositRefundInfo `bson:"inline"`
}

type ReceiveDepositRefundDoc struct {
	ID                       primitive.ObjectID `json:"id" bson:"id,omitempty"`
	ReceiveDepositRefundData `bson:"inline"`
	models.ActivityDoc       `bson:"inline"`
}

func (ReceiveDepositRefundDoc) CollectionName() string {
	return receivedepositrefundCollectionName
}

type ReceiveDepositRefundItemGuid struct {
	DocNo string `json:"docno" bson:"docno"`
}

func (ReceiveDepositRefundItemGuid) CollectionName() string {
	return receivedepositrefundCollectionName
}

type ReceiveDepositRefundActivity struct {
	ReceiveDepositRefundData `bson:"inline"`
	models.ActivityTime      `bson:"inline"`
}

func (ReceiveDepositRefundActivity) CollectionName() string {
	return receivedepositrefundCollectionName
}

type ReceiveDepositRefundDeleteActivity struct {
	models.Identity     `bson:"inline"`
	models.ActivityTime `bson:"inline"`
}

func (ReceiveDepositRefundDeleteActivity) CollectionName() string {
	return receivedepositrefundCollectionName
}
