package models

import (
	"smlcloudplatform/internal/models"
	transmodels "smlcloudplatform/internal/transaction/models"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

const chequedepositCollectionName = "transactionChequeDeposit"

type ChequeDeposit struct {
	models.PartitionIdentity     `bson:"inline"`
	transmodels.TransactionMoney `bson:"inline"`
}
type ChequeDepositInfo struct {
	models.DocIdentity `bson:"inline"`
	ChequeDeposit      `bson:"inline"`
}

func (ChequeDepositInfo) CollectionName() string {
	return chequedepositCollectionName
}

type ChequeDepositData struct {
	models.ShopIdentity `bson:"inline"`
	ChequeDepositInfo   `bson:"inline"`
}

type ChequeDepositDoc struct {
	ID                 primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	ChequeDepositData  `bson:"inline"`
	models.ActivityDoc `bson:"inline"`
}

func (ChequeDepositDoc) CollectionName() string {
	return chequedepositCollectionName
}

type ChequeDepositItemGuid struct {
	DocNo string `json:"docno" bson:"docno"`
}

func (ChequeDepositItemGuid) CollectionName() string {
	return chequedepositCollectionName
}

type ChequeDepositActivity struct {
	ChequeDepositData   `bson:"inline"`
	models.ActivityTime `bson:"inline"`
}

func (ChequeDepositActivity) CollectionName() string {
	return chequedepositCollectionName
}

type ChequeDepositDeleteActivity struct {
	models.Identity     `bson:"inline"`
	models.ActivityTime `bson:"inline"`
}

func (ChequeDepositDeleteActivity) CollectionName() string {
	return chequedepositCollectionName
}
