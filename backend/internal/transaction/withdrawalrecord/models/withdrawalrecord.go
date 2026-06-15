package models

import (
	"smlcloudplatform/internal/models"
	transmodels "smlcloudplatform/internal/transaction/models"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

const withdrawalrecordCollectionName = "transactionwithdrawalrecord"

type WithdrawalRecord struct {
	models.PartitionIdentity     `bson:"inline"`
	transmodels.TransactionMoney `bson:"inline"`
}
type WithdrawalRecordInfo struct {
	models.DocIdentity `bson:"inline"`
	WithdrawalRecord   `bson:"inline"`
}

func (WithdrawalRecordInfo) CollectionName() string {
	return withdrawalrecordCollectionName
}

type WithdrawalRecordData struct {
	models.HoldingCodeentity `bson:"inline"`
	WithdrawalRecordInfo     `bson:"inline"`
}

type WithdrawalRecordDoc struct {
	ID                   primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	WithdrawalRecordData `bson:"inline"`
	models.ActivityDoc   `bson:"inline"`
}

func (WithdrawalRecordDoc) CollectionName() string {
	return withdrawalrecordCollectionName
}

type WithdrawalRecordItemGuid struct {
	DocNo string `json:"docno" bson:"docno"`
}

func (WithdrawalRecordItemGuid) CollectionName() string {
	return withdrawalrecordCollectionName
}

type WithdrawalRecordActivity struct {
	WithdrawalRecordData `bson:"inline"`
	models.ActivityTime  `bson:"inline"`
}

func (WithdrawalRecordActivity) CollectionName() string {
	return withdrawalrecordCollectionName
}

type WithdrawalRecordDeleteActivity struct {
	models.Identity     `bson:"inline"`
	models.ActivityTime `bson:"inline"`
}

func (WithdrawalRecordDeleteActivity) CollectionName() string {
	return withdrawalrecordCollectionName
}
