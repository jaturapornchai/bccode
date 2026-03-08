package models

import (
	"encoding/json"
	"smlcloudplatform/internal/models"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

const accountperiodmasterCollectionName = "accountPeriodMaster"

// Request
type AccountPeriodMasterRequest struct {
	Period      int            `json:"period" bson:"period"`
	StartDate   models.ISODate `json:"startdate" bson:"startdate"`
	EndDate     models.ISODate `json:"enddate" bson:"enddate"`
	Description string         `json:"description" bson:"description"`
	IsDisabled  bool           `json:"isdisabled" bson:"isdisabled"`
}

func (apm *AccountPeriodMasterRequest) ToAccountPeriodMaster() AccountPeriodMaster {
	return AccountPeriodMaster{
		Period:      apm.Period,
		StartDate:   apm.StartDate.Time,
		EndDate:     apm.EndDate.Time,
		Description: apm.Description,
		IsDisabled:  apm.IsDisabled,
	}
}

// Account Period Master
type AccountPeriodMaster struct {
	models.PartitionIdentity `bson:"inline"`
	Period                   int       `json:"period" bson:"period"`
	StartDate                time.Time `json:"startdate" bson:"startdate"`
	EndDate                  time.Time `json:"enddate" bson:"enddate"`
	Description              string    `json:"description" bson:"description"`
	IsDisabled               bool      `json:"isdisabled" bson:"isdisabled"`
}

type AccountPeriodMasterInfo struct {
	models.DocIdentity  `bson:"inline"`
	AccountPeriodMaster `bson:"inline"`
}

func (AccountPeriodMasterInfo) CollectionName() string {
	return accountperiodmasterCollectionName
}

func (doc *AccountPeriodMasterInfo) MarshalJSON() ([]byte, error) {
	type Alias AccountPeriodMasterInfo
	return json.Marshal(&struct {
		*Alias
		StartDate string `json:"startdate"`
		EndDate   string `json:"enddate"`
	}{
		Alias:     (*Alias)(doc),
		StartDate: doc.StartDate.Format("2006-01-02"),
		EndDate:   doc.EndDate.Format("2006-01-02"),
	})
}

type AccountPeriodMasterData struct {
	models.ShopIdentity     `bson:"inline"`
	AccountPeriodMasterInfo `bson:"inline"`
}

type AccountPeriodMasterDoc struct {
	ID                      primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	AccountPeriodMasterData `bson:"inline"`
	models.ActivityDoc      `bson:"inline"`
}

func (AccountPeriodMasterDoc) CollectionName() string {
	return accountperiodmasterCollectionName
}

func (doc *AccountPeriodMasterDoc) MarshalJSON() ([]byte, error) {
	type Alias AccountPeriodMasterDoc
	return json.Marshal(&struct {
		*Alias
		StartDate string `json:"startdate"`
		EndDate   string `json:"enddate"`
	}{
		Alias:     (*Alias)(doc),
		StartDate: doc.StartDate.Format("2006-01-02"),
		EndDate:   doc.EndDate.Format("2006-01-02"),
	})
}

type AccountPeriodMasterItemGuid struct {
	DocNo string `json:"docno" bson:"docno"`
}

func (AccountPeriodMasterItemGuid) CollectionName() string {
	return accountperiodmasterCollectionName
}

type AccountPeriodMasterActivity struct {
	AccountPeriodMasterData `bson:"inline"`
	models.ActivityTime     `bson:"inline"`
}

func (AccountPeriodMasterActivity) CollectionName() string {
	return accountperiodmasterCollectionName
}

type AccountPeriodMasterDeleteActivity struct {
	models.Identity     `bson:"inline"`
	models.ActivityTime `bson:"inline"`
}

func (AccountPeriodMasterDeleteActivity) CollectionName() string {
	return accountperiodmasterCollectionName
}

type MapDateAccountPeriodMasterInfo struct {
	Date       string                  `json:"date"`
	PeriodData AccountPeriodMasterInfo `json:"perioddata"`
}
