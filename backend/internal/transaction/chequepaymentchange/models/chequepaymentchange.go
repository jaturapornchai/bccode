package models

import (
	"smlcloudplatform/internal/models"
	transmodels "smlcloudplatform/internal/transaction/models"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

const chequepaymentchangeCollectionName = "transactionChequePaymentChange"

type ChequePaymentChange struct {
	models.PartitionIdentity     `bson:"inline"`
	transmodels.TransactionMoney `bson:"inline"`
}
type ChequePaymentChangeInfo struct {
	models.DocIdentity  `bson:"inline"`
	ChequePaymentChange `bson:"inline"`
}

func (ChequePaymentChangeInfo) CollectionName() string {
	return chequepaymentchangeCollectionName
}

type ChequePaymentChangeData struct {
	models.HoldingCodeentity `bson:"inline"`
	ChequePaymentChangeInfo  `bson:"inline"`
}

type ChequePaymentChangeDoc struct {
	ID                      primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	ChequePaymentChangeData `bson:"inline"`
	models.ActivityDoc      `bson:"inline"`
}

func (ChequePaymentChangeDoc) CollectionName() string {
	return chequepaymentchangeCollectionName
}

type ChequePaymentChangeItemGuid struct {
	DocNo string `json:"docno" bson:"docno"`
}

func (ChequePaymentChangeItemGuid) CollectionName() string {
	return chequepaymentchangeCollectionName
}

type ChequePaymentChangeActivity struct {
	ChequePaymentChangeData `bson:"inline"`
	models.ActivityTime     `bson:"inline"`
}

func (ChequePaymentChangeActivity) CollectionName() string {
	return chequepaymentchangeCollectionName
}

type ChequePaymentChangeDeleteActivity struct {
	models.Identity     `bson:"inline"`
	models.ActivityTime `bson:"inline"`
}

func (ChequePaymentChangeDeleteActivity) CollectionName() string {
	return chequepaymentchangeCollectionName
}
