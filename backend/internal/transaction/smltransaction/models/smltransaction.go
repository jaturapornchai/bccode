package models

import (
	"smlcloudplatform/internal/models"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type DynamicCollection struct {
	Collection string
}

func (d *DynamicCollection) SetCollectionName(collectionName string) {
	d.Collection = collectionName
}

func (d *DynamicCollection) CollectionName() string {
	return d.Collection
}

const smltransactionCollectionName = "smlTransactions"

type SMLTransaction struct {
	models.PartitionIdentity `bson:"inline"`
	DocNo                    string                 `json:"docno" bson:"docno"`
	DynamicData              map[string]interface{} `json:"dynamic_data,omitempty"`
	// MetaData2                map[string]map[string]interface{}
}

type SMLTransactionInfo struct {
	models.DocIdentity `bson:"inline"`
	SMLTransaction     `bson:"inline"`
}

func (SMLTransactionInfo) CollectionName() string {
	return smltransactionCollectionName
}

type SMLTransactionData struct {
	models.ShopIdentity `bson:"inline"`
	SMLTransactionInfo  `bson:"inline"`
}

type SMLTransactionDoc struct {
	ID                 primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	SMLTransactionData `bson:"inline"`
	models.ActivityDoc `bson:"inline"`
}

func (SMLTransactionDoc) CollectionName() string {
	return smltransactionCollectionName
}

type SMLTransactionItemGuid struct {
	DocNo string `json:"docno" bson:"docno"`
}

func (SMLTransactionItemGuid) CollectionName() string {
	return smltransactionCollectionName
}

type SMLTransactionActivity struct {
	SMLTransactionData  `bson:"inline"`
	models.ActivityTime `bson:"inline"`
}

func (SMLTransactionActivity) CollectionName() string {
	return smltransactionCollectionName
}

type SMLTransactionDeleteActivity struct {
	models.Identity     `bson:"inline"`
	models.ActivityTime `bson:"inline"`
}

func (SMLTransactionDeleteActivity) CollectionName() string {
	return smltransactionCollectionName
}
