package models

import (
	"smlcloudplatform/internal/models"
	transmodels "smlcloudplatform/internal/transaction/models"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

const chequedisqualifiedCollectionName = "transactionChequeDisqualified"

type ChequeDisqualified struct {
	models.PartitionIdentity     `bson:"inline"`
	transmodels.TransactionMoney `bson:"inline"`
}
type ChequeDisqualifiedInfo struct {
	models.DocIdentity `bson:"inline"`
	ChequeDisqualified `bson:"inline"`
}

func (ChequeDisqualifiedInfo) CollectionName() string {
	return chequedisqualifiedCollectionName
}

type ChequeDisqualifiedData struct {
	models.ShopIdentity    `bson:"inline"`
	ChequeDisqualifiedInfo `bson:"inline"`
}

type ChequeDisqualifiedDoc struct {
	ID primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	ChequeDisqualifiedData `bson:"inline"`
	models.ActivityDoc     `bson:"inline"`
}

func (ChequeDisqualifiedDoc) CollectionName() string {
	return chequedisqualifiedCollectionName
}

type ChequeDisqualifiedItemGuid struct {
	DocNo string `json:"docno" bson:"docno"`
}

func (ChequeDisqualifiedItemGuid) CollectionName() string {
	return chequedisqualifiedCollectionName
}

type ChequeDisqualifiedActivity struct {
	ChequeDisqualifiedData `bson:"inline"`
	models.ActivityTime    `bson:"inline"`
}

func (ChequeDisqualifiedActivity) CollectionName() string {
	return chequedisqualifiedCollectionName
}

type ChequeDisqualifiedDeleteActivity struct {
	models.Identity     `bson:"inline"`
	models.ActivityTime `bson:"inline"`
}

func (ChequeDisqualifiedDeleteActivity) CollectionName() string {
	return chequedisqualifiedCollectionName
}
