package models

import (
	"smlcloudplatform/internal/models"
	transmodels "smlcloudplatform/internal/transaction/models"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

const chequereturnCollectionName = "transactionChequeReturn"

type ChequeReturn struct {
	models.PartitionIdentity     `bson:"inline"`
	transmodels.TransactionMoney `bson:"inline"`
}
type ChequeReturnInfo struct {
	models.DocIdentity `bson:"inline"`
	ChequeReturn       `bson:"inline"`
}

func (ChequeReturnInfo) CollectionName() string {
	return chequereturnCollectionName
}

type ChequeReturnData struct {
	models.HoldingCodeentity `bson:"inline"`
	ChequeReturnInfo         `bson:"inline"`
}

type ChequeReturnDoc struct {
	ID                 primitive.ObjectID `json:"id" bson:"id,omitempty"`
	ChequeReturnData   `bson:"inline"`
	models.ActivityDoc `bson:"inline"`
}

func (ChequeReturnDoc) CollectionName() string {
	return chequereturnCollectionName
}

type ChequeReturnItemGuid struct {
	DocNo string `json:"docno" bson:"docno"`
}

func (ChequeReturnItemGuid) CollectionName() string {
	return chequereturnCollectionName
}

type ChequeReturnActivity struct {
	ChequeReturnData    `bson:"inline"`
	models.ActivityTime `bson:"inline"`
}

func (ChequeReturnActivity) CollectionName() string {
	return chequereturnCollectionName
}

type ChequeReturnDeleteActivity struct {
	models.Identity     `bson:"inline"`
	models.ActivityTime `bson:"inline"`
}

func (ChequeReturnDeleteActivity) CollectionName() string {
	return chequereturnCollectionName
}
