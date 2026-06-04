package models

import (
	"smlcloudplatform/internal/models"
	transmodels "smlcloudplatform/internal/transaction/models"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

const chequepaymentreturnCollectionName = "transactionChequePaymentReturn"

type ChequePaymentReturn struct {
	models.PartitionIdentity     `bson:"inline"`
	transmodels.TransactionMoney `bson:"inline"`
}
type ChequePaymentReturnInfo struct {
	models.DocIdentity  `bson:"inline"`
	ChequePaymentReturn `bson:"inline"`
}

func (ChequePaymentReturnInfo) CollectionName() string {
	return chequepaymentreturnCollectionName
}

type ChequePaymentReturnData struct {
	models.HoldingCodeentity `bson:"inline"`
	ChequePaymentReturnInfo  `bson:"inline"`
}

type ChequePaymentReturnDoc struct {
	ID                      primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	ChequePaymentReturnData `bson:"inline"`
	models.ActivityDoc      `bson:"inline"`
}

func (ChequePaymentReturnDoc) CollectionName() string {
	return chequepaymentreturnCollectionName
}

type ChequePaymentReturnItemGuid struct {
	DocNo string `json:"docno" bson:"docno"`
}

func (ChequePaymentReturnItemGuid) CollectionName() string {
	return chequepaymentreturnCollectionName
}

type ChequePaymentReturnActivity struct {
	ChequePaymentReturnData `bson:"inline"`
	models.ActivityTime     `bson:"inline"`
}

func (ChequePaymentReturnActivity) CollectionName() string {
	return chequepaymentreturnCollectionName
}

type ChequePaymentReturnDeleteActivity struct {
	models.Identity     `bson:"inline"`
	models.ActivityTime `bson:"inline"`
}

func (ChequePaymentReturnDeleteActivity) CollectionName() string {
	return chequepaymentreturnCollectionName
}
