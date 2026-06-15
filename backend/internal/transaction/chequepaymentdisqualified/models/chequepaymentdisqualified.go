package models

import (
	"smlcloudplatform/internal/models"
	transmodels "smlcloudplatform/internal/transaction/models"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

const chequepaymentdisqualifiedCollectionName = "transactionchequepaymentdisqualified"

type ChequePaymentDisqualified struct {
	models.PartitionIdentity     `bson:"inline"`
	transmodels.TransactionMoney `bson:"inline"`
}
type ChequePaymentDisqualifiedInfo struct {
	models.DocIdentity        `bson:"inline"`
	ChequePaymentDisqualified `bson:"inline"`
}

func (ChequePaymentDisqualifiedInfo) CollectionName() string {
	return chequepaymentdisqualifiedCollectionName
}

type ChequePaymentDisqualifiedData struct {
	models.HoldingCodeentity      `bson:"inline"`
	ChequePaymentDisqualifiedInfo `bson:"inline"`
}

type ChequePaymentDisqualifiedDoc struct {
	ID                            primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	ChequePaymentDisqualifiedData `bson:"inline"`
	models.ActivityDoc            `bson:"inline"`
}

func (ChequePaymentDisqualifiedDoc) CollectionName() string {
	return chequepaymentdisqualifiedCollectionName
}

type ChequePaymentDisqualifiedItemGuid struct {
	DocNo string `json:"docno" bson:"docno"`
}

func (ChequePaymentDisqualifiedItemGuid) CollectionName() string {
	return chequepaymentdisqualifiedCollectionName
}

type ChequePaymentDisqualifiedActivity struct {
	ChequePaymentDisqualifiedData `bson:"inline"`
	models.ActivityTime           `bson:"inline"`
}

func (ChequePaymentDisqualifiedActivity) CollectionName() string {
	return chequepaymentdisqualifiedCollectionName
}

type ChequePaymentDisqualifiedDeleteActivity struct {
	models.Identity     `bson:"inline"`
	models.ActivityTime `bson:"inline"`
}

func (ChequePaymentDisqualifiedDeleteActivity) CollectionName() string {
	return chequepaymentdisqualifiedCollectionName
}
