package models

import (
	"smlcloudplatform/internal/models"
	transmodels "smlcloudplatform/internal/transaction/models"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

const banktransferrecordCollectionName = "transactionBankTransferRecord"

type BankTransferRecord struct {
	models.PartitionIdentity     `bson:"inline"`
	transmodels.TransactionMoney `bson:"inline"`
}
type BankTransferRecordInfo struct {
	models.DocIdentity `bson:"inline"`
	BankTransferRecord `bson:"inline"`
}

func (BankTransferRecordInfo) CollectionName() string {
	return banktransferrecordCollectionName
}

type BankTransferRecordData struct {
	models.HoldingCodeentity `bson:"inline"`
	BankTransferRecordInfo   `bson:"inline"`
}

type BankTransferRecordDoc struct {
	ID                     primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	BankTransferRecordData `bson:"inline"`
	models.ActivityDoc     `bson:"inline"`
}

func (BankTransferRecordDoc) CollectionName() string {
	return banktransferrecordCollectionName
}

type BankTransferRecordItemGuid struct {
	DocNo string `json:"docno" bson:"docno"`
}

func (BankTransferRecordItemGuid) CollectionName() string {
	return banktransferrecordCollectionName
}

type BankTransferRecordActivity struct {
	BankTransferRecordData `bson:"inline"`
	models.ActivityTime    `bson:"inline"`
}

func (BankTransferRecordActivity) CollectionName() string {
	return banktransferrecordCollectionName
}

type BankTransferRecordDeleteActivity struct {
	models.Identity     `bson:"inline"`
	models.ActivityTime `bson:"inline"`
}

func (BankTransferRecordDeleteActivity) CollectionName() string {
	return banktransferrecordCollectionName
}
