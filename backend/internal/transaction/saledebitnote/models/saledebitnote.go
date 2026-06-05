package models

import (
	"smlcloudplatform/internal/models"
	transmodels "smlcloudplatform/internal/transaction/models"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

const saledebitnoteCollectionName = "transactionSaleDebitNote"

type SaleDebitNote struct {
	models.PartitionIdentity     `bson:"inline"`
	transmodels.TransactionMoney `bson:"inline"`
}
type SaleDebitNoteInfo struct {
	models.DocIdentity `bson:"inline"`
	SaleDebitNote      `bson:"inline"`
}

func (SaleDebitNoteInfo) CollectionName() string {
	return saledebitnoteCollectionName
}

type SaleDebitNoteData struct {
	models.HoldingCodeentity `bson:"inline"`
	SaleDebitNoteInfo        `bson:"inline"`
}

type SaleDebitNoteDoc struct {
	ID                 primitive.ObjectID `json:"id" bson:"id,omitempty"`
	SaleDebitNoteData  `bson:"inline"`
	models.ActivityDoc `bson:"inline"`
}

func (SaleDebitNoteDoc) CollectionName() string {
	return saledebitnoteCollectionName
}

type SaleDebitNoteItemGuid struct {
	DocNo string `json:"docno" bson:"docno"`
}

func (SaleDebitNoteItemGuid) CollectionName() string {
	return saledebitnoteCollectionName
}

type SaleDebitNoteActivity struct {
	SaleDebitNoteData   `bson:"inline"`
	models.ActivityTime `bson:"inline"`
}

func (SaleDebitNoteActivity) CollectionName() string {
	return saledebitnoteCollectionName
}

type SaleDebitNoteDeleteActivity struct {
	models.Identity     `bson:"inline"`
	models.ActivityTime `bson:"inline"`
}

func (SaleDebitNoteDeleteActivity) CollectionName() string {
	return saledebitnoteCollectionName
}
