package models

import (
	"smlcloudplatform/internal/models"
	transmodels "smlcloudplatform/internal/transaction/models"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

const chequepaymentdepositCollectionName = "transactionChequePaymentDeposit"

type ChequePaymentDeposit struct {
	models.PartitionIdentity     `bson:"inline"`
	transmodels.TransactionMoney `bson:"inline"`
}
type ChequePaymentDepositInfo struct {
	models.DocIdentity   `bson:"inline"`
	ChequePaymentDeposit `bson:"inline"`
}

func (ChequePaymentDepositInfo) CollectionName() string {
	return chequepaymentdepositCollectionName
}

type ChequePaymentDepositData struct {
	models.HoldingCodeentity `bson:"inline"`
	ChequePaymentDepositInfo `bson:"inline"`
}

type ChequePaymentDepositDoc struct {
	ID                       primitive.ObjectID `json:"id" bson:"id,omitempty"`
	ChequePaymentDepositData `bson:"inline"`
	models.ActivityDoc       `bson:"inline"`
}

func (ChequePaymentDepositDoc) CollectionName() string {
	return chequepaymentdepositCollectionName
}

type ChequePaymentDepositItemGuid struct {
	DocNo string `json:"docno" bson:"docno"`
}

func (ChequePaymentDepositItemGuid) CollectionName() string {
	return chequepaymentdepositCollectionName
}

type ChequePaymentDepositActivity struct {
	ChequePaymentDepositData `bson:"inline"`
	models.ActivityTime      `bson:"inline"`
}

func (ChequePaymentDepositActivity) CollectionName() string {
	return chequepaymentdepositCollectionName
}

type ChequePaymentDepositDeleteActivity struct {
	models.Identity     `bson:"inline"`
	models.ActivityTime `bson:"inline"`
}

func (ChequePaymentDepositDeleteActivity) CollectionName() string {
	return chequepaymentdepositCollectionName
}
