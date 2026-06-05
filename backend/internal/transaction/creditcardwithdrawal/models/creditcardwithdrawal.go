package models

import (
	"smlcloudplatform/internal/models"
	transmodels "smlcloudplatform/internal/transaction/models"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

const creditcardwithdrawalCollectionName = "transactioncreditcardwithdrawal"

type CreditCardWithdrawal struct {
	models.PartitionIdentity     `bson:"inline"`
	transmodels.TransactionMoney `bson:"inline"`
}
type CreditCardWithdrawalInfo struct {
	models.DocIdentity   `bson:"inline"`
	CreditCardWithdrawal `bson:"inline"`
}

func (CreditCardWithdrawalInfo) CollectionName() string {
	return creditcardwithdrawalCollectionName
}

type CreditCardWithdrawalData struct {
	models.HoldingCodeentity `bson:"inline"`
	CreditCardWithdrawalInfo `bson:"inline"`
}

type CreditCardWithdrawalDoc struct {
	ID                       primitive.ObjectID `json:"id" bson:"id,omitempty"`
	CreditCardWithdrawalData `bson:"inline"`
	models.ActivityDoc       `bson:"inline"`
}

func (CreditCardWithdrawalDoc) CollectionName() string {
	return creditcardwithdrawalCollectionName
}

type CreditCardWithdrawalItemGuid struct {
	DocNo string `json:"docno" bson:"docno"`
}

func (CreditCardWithdrawalItemGuid) CollectionName() string {
	return creditcardwithdrawalCollectionName
}

type CreditCardWithdrawalActivity struct {
	CreditCardWithdrawalData `bson:"inline"`
	models.ActivityTime      `bson:"inline"`
}

func (CreditCardWithdrawalActivity) CollectionName() string {
	return creditcardwithdrawalCollectionName
}

type CreditCardWithdrawalDeleteActivity struct {
	models.Identity     `bson:"inline"`
	models.ActivityTime `bson:"inline"`
}

func (CreditCardWithdrawalDeleteActivity) CollectionName() string {
	return creditcardwithdrawalCollectionName
}
