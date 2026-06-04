package models

import (
	"smlcloudplatform/internal/models"
	transmodels "smlcloudplatform/internal/transaction/models"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

const saleorderCollectionName = "transactionQuotation"

type Quotation struct {
	models.PartitionIdentity `bson:"inline"`
	transmodels.Transaction  `bson:"inline"`
}
type QuotationInfo struct {
	models.DocIdentity `bson:"inline"`
	Quotation          `bson:"inline"`
}

func (QuotationInfo) CollectionName() string {
	return saleorderCollectionName
}

type QuotationData struct {
	models.HoldingCodeentity `bson:"inline"`
	QuotationInfo            `bson:"inline"`
}

type QuotationDoc struct {
	ID                 primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	QuotationData      `bson:"inline"`
	models.ActivityDoc `bson:"inline"`
}

func (QuotationDoc) CollectionName() string {
	return saleorderCollectionName
}

type QuotationItemGuid struct {
	DocNo string `json:"docno" bson:"docno"`
}

func (QuotationItemGuid) CollectionName() string {
	return saleorderCollectionName
}

type QuotationActivity struct {
	QuotationData       `bson:"inline"`
	models.ActivityTime `bson:"inline"`
}

func (QuotationActivity) CollectionName() string {
	return saleorderCollectionName
}

type QuotationDeleteActivity struct {
	models.Identity     `bson:"inline"`
	models.ActivityTime `bson:"inline"`
}

func (QuotationDeleteActivity) CollectionName() string {
	return saleorderCollectionName
}
