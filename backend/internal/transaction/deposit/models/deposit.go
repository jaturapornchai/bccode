package models

import (
	"smlcloudplatform/internal/models"
	transmodels "smlcloudplatform/internal/transaction/models"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

const depositCollectionName = "transactionDeposit"

type Deposit struct {
	models.PartitionIdentity `bson:"inline"`
	transmodels.Transaction  `bson:"inline"`
}
type DepositInfo struct {
	models.DocIdentity `bson:"inline"`
	Deposit            `bson:"inline"`
}

func (DepositInfo) CollectionName() string {
	return depositCollectionName
}

type DepositData struct {
	models.HoldingCodeentity `bson:"inline"`
	DepositInfo              `bson:"inline"`
}

type DepositDoc struct {
	ID                 primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	DepositData        `bson:"inline"`
	models.ActivityDoc `bson:"inline"`
}

func (DepositDoc) CollectionName() string {
	return depositCollectionName
}

type DepositItemGuid struct {
	DocNo string `json:"docno" bson:"docno"`
}

func (DepositItemGuid) CollectionName() string {
	return depositCollectionName
}

type DepositActivity struct {
	DepositData         `bson:"inline"`
	models.ActivityTime `bson:"inline"`
}

func (DepositActivity) CollectionName() string {
	return depositCollectionName
}

type DepositDeleteActivity struct {
	models.Identity     `bson:"inline"`
	models.ActivityTime `bson:"inline"`
}

func (DepositDeleteActivity) CollectionName() string {
	return depositCollectionName
}
