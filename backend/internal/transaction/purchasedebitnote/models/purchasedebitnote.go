package models

import (
	"smlcloudplatform/internal/models"
	transmodels "smlcloudplatform/internal/transaction/models"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

const purchasedebitnoteCollectionName = "transactionPurchaseDebitNote"

type PurchaseDebitNote struct {
	models.PartitionIdentity     `bson:"inline"`
	transmodels.TransactionMoney `bson:"inline"`
}
type PurchaseDebitNoteInfo struct {
	models.DocIdentity `bson:"inline"`
	PurchaseDebitNote  `bson:"inline"`
}

func (PurchaseDebitNoteInfo) CollectionName() string {
	return purchasedebitnoteCollectionName
}

type PurchaseDebitNoteData struct {
	models.HoldingCodeentity `bson:"inline"`
	PurchaseDebitNoteInfo    `bson:"inline"`
}

type PurchaseDebitNoteDoc struct {
	ID                    primitive.ObjectID `json:"id" bson:"id,omitempty"`
	PurchaseDebitNoteData `bson:"inline"`
	models.ActivityDoc    `bson:"inline"`
}

func (PurchaseDebitNoteDoc) CollectionName() string {
	return purchasedebitnoteCollectionName
}

type PurchaseDebitNoteItemGuid struct {
	DocNo string `json:"docno" bson:"docno"`
}

func (PurchaseDebitNoteItemGuid) CollectionName() string {
	return purchasedebitnoteCollectionName
}

type PurchaseDebitNoteActivity struct {
	PurchaseDebitNoteData `bson:"inline"`
	models.ActivityTime   `bson:"inline"`
}

func (PurchaseDebitNoteActivity) CollectionName() string {
	return purchasedebitnoteCollectionName
}

type PurchaseDebitNoteDeleteActivity struct {
	models.Identity     `bson:"inline"`
	models.ActivityTime `bson:"inline"`
}

func (PurchaseDebitNoteDeleteActivity) CollectionName() string {
	return purchasedebitnoteCollectionName
}
