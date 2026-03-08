package models

import (
	"smlcloudplatform/internal/models"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

const accountGroupCollectionName = "accountGroups"
const accountGroupTableName = "account_group"

type AccountGroup struct {
	Code        string `json:"code" bson:"code" validate:"required"`
	models.Name `bson:"inline"`
}

type AccountGroupInfo struct {
	models.DocIdentity `bson:"inline"`
	AccountGroup       `bson:"inline"`
}

func (AccountGroupInfo) CollectionName() string {
	return accountGroupCollectionName
}

type AccountGroupData struct {
	models.ShopIdentity `bson:"inline"`
	AccountGroupInfo    `bson:"inline"`
}

type AccountGroupDoc struct {
	ID                 primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	AccountGroupData   `bson:"inline"`
	models.ActivityDoc `bson:"inline"`
}

func (AccountGroupDoc) CollectionName() string {
	return accountGroupCollectionName
}

type AccountGroupIdentifier struct {
	DocNo string `json:"code" bson:"code" gorm:"code"`
}

func (AccountGroupIdentifier) CollectionName() string {
	return accountGroupCollectionName
}

type AccountGroupActivity struct {
	AccountGroupData    `bson:"inline"`
	models.ActivityTime `bson:"inline"`
}

func (AccountGroupActivity) CollectionName() string {
	return accountGroupCollectionName
}

type AccountGroupDeleteActivity struct {
	models.Identity     `bson:"inline"`
	models.ActivityTime `bson:"inline"`
}

func (AccountGroupDeleteActivity) CollectionName() string {
	return accountGroupCollectionName
}

// Postgresql model
type AccountGroupPg struct {
	models.Identity          `gorm:"embedded;"`
	models.PartitionIdentity `gorm:"embedded;"`
	Code                     string `json:"code" gorm:"column:code;primaryKey"`
	Name1                    string `json:"name1" gorm:"column:name1"`
}

func (AccountGroup) TableName() string {
	return accountGroupTableName
}

type AccountGroupInfoResponse struct {
	Success bool             `json:"success"`
	Data    AccountGroupInfo `json:"data,omitempty"`
}

type AccountGroupPageResponse struct {
	Success    bool                          `json:"success"`
	Data       []AccountGroupInfo            `json:"data,omitempty"`
	Pagination models.PaginationDataResponse `json:"pagination,omitempty"`
}
