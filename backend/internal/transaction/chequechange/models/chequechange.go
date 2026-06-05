package models

import (
	"smlcloudplatform/internal/models"
	transmodels "smlcloudplatform/internal/transaction/models"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

const chequechangeCollectionName = "transactionChequeChange"

type ChequeChange struct {
	models.PartitionIdentity     `bson:"inline"`
	transmodels.TransactionMoney `bson:"inline"`
}
type ChequeChangeInfo struct {
	models.DocIdentity `bson:"inline"`
	ChequeChange       `bson:"inline"`
}

func (ChequeChangeInfo) CollectionName() string {
	return chequechangeCollectionName
}

type ChequeChangeData struct {
	models.HoldingCodeentity `bson:"inline"`
	ChequeChangeInfo         `bson:"inline"`
}

type ChequeChangeDoc struct {
	ID                 primitive.ObjectID `json:"id" bson:"id,omitempty"`
	ChequeChangeData   `bson:"inline"`
	models.ActivityDoc `bson:"inline"`
}

func (ChequeChangeDoc) CollectionName() string {
	return chequechangeCollectionName
}

type ChequeChangeItemGuid struct {
	DocNo string `json:"docno" bson:"docno"`
}

func (ChequeChangeItemGuid) CollectionName() string {
	return chequechangeCollectionName
}

type ChequeChangeActivity struct {
	ChequeChangeData    `bson:"inline"`
	models.ActivityTime `bson:"inline"`
}

func (ChequeChangeActivity) CollectionName() string {
	return chequechangeCollectionName
}

type ChequeChangeDeleteActivity struct {
	models.Identity     `bson:"inline"`
	models.ActivityTime `bson:"inline"`
}

func (ChequeChangeDeleteActivity) CollectionName() string {
	return chequechangeCollectionName
}
