package models

import (
	"smlcloudplatform/internal/models"
	transmodels "smlcloudplatform/internal/transaction/models"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

const chequepassCollectionName = "transactionchequepass"

type ChequePass struct {
	models.PartitionIdentity     `bson:"inline"`
	transmodels.TransactionMoney `bson:"inline"`
}
type ChequePassInfo struct {
	models.DocIdentity `bson:"inline"`
	ChequePass         `bson:"inline"`
}

func (ChequePassInfo) CollectionName() string {
	return chequepassCollectionName
}

type ChequePassData struct {
	models.HoldingCodeentity `bson:"inline"`
	ChequePassInfo           `bson:"inline"`
}

type ChequePassDoc struct {
	ID                 primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	ChequePassData     `bson:"inline"`
	models.ActivityDoc `bson:"inline"`
}

func (ChequePassDoc) CollectionName() string {
	return chequepassCollectionName
}

type ChequePassItemGuid struct {
	DocNo string `json:"docno" bson:"docno"`
}

func (ChequePassItemGuid) CollectionName() string {
	return chequepassCollectionName
}

type ChequePassActivity struct {
	ChequePassData      `bson:"inline"`
	models.ActivityTime `bson:"inline"`
}

func (ChequePassActivity) CollectionName() string {
	return chequepassCollectionName
}

type ChequePassDeleteActivity struct {
	models.Identity     `bson:"inline"`
	models.ActivityTime `bson:"inline"`
}

func (ChequePassDeleteActivity) CollectionName() string {
	return chequepassCollectionName
}
