package models

import (
	"smlcloudplatform/internal/models"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

const journalBookCollectionName = "journalBooks"
const journalBookTableName = "journal_book"

type JournalBook struct {
	Code        string `json:"code" bson:"code"`
	models.Name `bson:"inline"`
}

type JournalBookInfo struct {
	models.DocIdentity `bson:"inline"`
	JournalBook        `bson:"inline"`
}

func (JournalBookInfo) CollectionName() string {
	return journalBookCollectionName
}

type JournalBookData struct {
	models.ShopIdentity `bson:"inline"`
	JournalBookInfo     `bson:"inline"`
}

type JournalBookDoc struct {
	ID                 primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	JournalBookData    `bson:"inline"`
	models.ActivityDoc `bson:"inline"`
}

func (JournalBookDoc) CollectionName() string {
	return journalBookCollectionName
}

type JournalBookIdentifier struct {
	DocNo string `json:"code" bson:"code" gorm:"code"`
}

func (JournalBookIdentifier) CollectionName() string {
	return journalBookCollectionName
}

type JournalBookActivity struct {
	JournalBookData     `bson:"inline"`
	models.ActivityTime `bson:"inline"`
}

func (JournalBookActivity) CollectionName() string {
	return journalBookCollectionName
}

type JournalBookDeleteActivity struct {
	models.Identity     `bson:"inline"`
	models.ActivityTime `bson:"inline"`
}

func (JournalBookDeleteActivity) CollectionName() string {
	return journalBookCollectionName
}

// Postgresql model
type JournalPg struct {
	models.Identity          `gorm:"embedded;"`
	models.PartitionIdentity `gorm:"embedded;"`
	Code                     string `json:"code" gorm:"column:code;primaryKey"`
	Name1                    string `json:"name1" gorm:"column:name1"`
}

func (JournalPg) TableName() string {
	return journalBookTableName
}

type JournalBookInfoResponse struct {
	Success bool            `json:"success"`
	Data    JournalBookInfo `json:"data,omitempty"`
}

type JournalBookPageResponse struct {
	Success    bool                          `json:"success"`
	Data       []JournalBookInfo             `json:"data,omitempty"`
	Pagination models.PaginationDataResponse `json:"pagination,omitempty"`
}
