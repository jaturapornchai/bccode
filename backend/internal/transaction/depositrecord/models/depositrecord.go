package models

import (
	"smlcloudplatform/internal/models"
	transmodels "smlcloudplatform/internal/transaction/models"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

const depositrecordCollectionName = "transactiondepositrecord"

type DepositRecord struct {
	models.PartitionIdentity     `bson:"inline"`
	transmodels.TransactionMoney `bson:"inline"`
}
type DepositRecordInfo struct {
	models.DocIdentity `bson:"inline"`
	DepositRecord      `bson:"inline"`
}

func (DepositRecordInfo) CollectionName() string {
	return depositrecordCollectionName
}

type DepositRecordData struct {
	models.HoldingCodeentity `bson:"inline"`
	DepositRecordInfo        `bson:"inline"`
}

type DepositRecordDoc struct {
	ID                 primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	DepositRecordData  `bson:"inline"`
	models.ActivityDoc `bson:"inline"`
}

func (DepositRecordDoc) CollectionName() string {
	return depositrecordCollectionName
}

type DepositRecordItemGuid struct {
	DocNo string `json:"docno" bson:"docno"`
}

func (DepositRecordItemGuid) CollectionName() string {
	return depositrecordCollectionName
}

type DepositRecordActivity struct {
	DepositRecordData   `bson:"inline"`
	models.ActivityTime `bson:"inline"`
}

func (DepositRecordActivity) CollectionName() string {
	return depositrecordCollectionName
}

type DepositRecordDeleteActivity struct {
	models.Identity     `bson:"inline"`
	models.ActivityTime `bson:"inline"`
}

func (DepositRecordDeleteActivity) CollectionName() string {
	return depositrecordCollectionName
}
