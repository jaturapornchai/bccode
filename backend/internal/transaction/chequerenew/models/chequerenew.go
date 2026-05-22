package models

import (
	"smlcloudplatform/internal/models"
	transmodels "smlcloudplatform/internal/transaction/models"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

const chequerenewCollectionName = "transactionChequeRenew"

type ChequeRenew struct {
	models.PartitionIdentity     `bson:"inline"`
	transmodels.TransactionMoney `bson:"inline"`
}
type ChequeRenewInfo struct {
	models.DocIdentity `bson:"inline"`
	ChequeRenew  `bson:"inline"`
}

func (ChequeRenewInfo) CollectionName() string {
	return chequerenewCollectionName
}

type ChequeRenewData struct {
	models.ShopIdentity `bson:"inline"`
	ChequeRenewInfo  `bson:"inline"`
}

type ChequeRenewDoc struct {
	ID primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	ChequeRenewData  `bson:"inline"`
	models.ActivityDoc `bson:"inline"`
}

func (ChequeRenewDoc) CollectionName() string {
	return chequerenewCollectionName
}

type ChequeRenewItemGuid struct {
	DocNo string `json:"docno" bson:"docno"`
}

func (ChequeRenewItemGuid) CollectionName() string {
	return chequerenewCollectionName
}

type ChequeRenewActivity struct {
	ChequeRenewData  `bson:"inline"`
	models.ActivityTime `bson:"inline"`
}

func (ChequeRenewActivity) CollectionName() string {
	return chequerenewCollectionName
}

type ChequeRenewDeleteActivity struct {
	models.Identity     `bson:"inline"`
	models.ActivityTime `bson:"inline"`
}

func (ChequeRenewDeleteActivity) CollectionName() string {
	return chequerenewCollectionName
}
