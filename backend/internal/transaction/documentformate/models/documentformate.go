package models

import (
	"smlcloudplatform/internal/models"
	chartofaccount_models "smlcloudplatform/internal/vfgl/chartofaccount/models"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

const documentformateCollectionName = "documentformate"

type DocumentFormate struct {
	models.PartitionIdentity `bson:"inline"`
	DocCode                  string                   `json:"doccode" bson:"doccode" validate:"required,min=1"`
	Module                   string                   `json:"module" bson:"module" validate:"min=1"`
	DateFormate              string                   `json:"dateformate" bson:"dateformate"`
	DocNumber                int                      `json:"docnumber" bson:"docnumber"`
	DocFormat                string                   `json:"docformat" bson:"docformat"`
	Description              string                   `json:"description" bson:"description"`
	Details                  *[]DocumentFormateDetail `json:"details" bson:"details"`
	IsAutoFormat             bool                     `json:"isautoformat" bson:"isautoformat"`
	YearType                 int8                     `json:"yeartype" bson:"yeartype"`
	AccountGroup             string                   `json:"accountgroup" bson:"accountgroup"`
	BookCode                 string                   `json:"bookcode" bson:"bookcode"`
	PromptDescription        string                   `json:"promptdescription" bson:"promptdescription"`
}

type DocumentFormateDetail struct {
	AccountCode        string                                   `json:"accountcode,omitempty" bson:"accountcode,omitempty"`
	ActionCode         string                                   `json:"actioncode" bson:"actioncode" validate:"required,min=1"`
	Detail             string                                   `json:"detail" bson:"detail"`
	Debit              string                                   `json:"debit" bson:"debit"`
	Credit             string                                   `json:"credit" bson:"credit"`
	IsEntrySelfAccount bool                                     `json:"isentryselfaccount" bson:"isentryselfaccount"`
	AccountDebit       chartofaccount_models.ChartOfAccountInfo `json:"accountdebit" bson:"accountdebit"`
	AccountCredit      chartofaccount_models.ChartOfAccountInfo `json:"accountcredit" bson:"accountcredit"`
}

type DocumentFormateInfo struct {
	models.DocIdentity `bson:"inline"`
	DocumentFormate    `bson:"inline"`
}

func (DocumentFormateInfo) CollectionName() string {
	return documentformateCollectionName
}

type DocumentFormateData struct {
	models.HoldingCodeentity `bson:"inline"`
	DocumentFormateInfo      `bson:"inline"`
}

type DocumentFormateDoc struct {
	ID                  primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	DocumentFormateData `bson:"inline"`
	models.ActivityDoc  `bson:"inline"`
}

func (DocumentFormateDoc) CollectionName() string {
	return documentformateCollectionName
}

type DocumentFormateItemGuid struct {
	DocCode string `json:"doccode" bson:"doccode"`
}

func (DocumentFormateItemGuid) CollectionName() string {
	return documentformateCollectionName
}

type DocumentFormateActivity struct {
	DocumentFormateData `bson:"inline"`
	models.ActivityTime `bson:"inline"`
}

func (DocumentFormateActivity) CollectionName() string {
	return documentformateCollectionName
}

type DocumentFormateDeleteActivity struct {
	models.Identity     `bson:"inline"`
	models.ActivityTime `bson:"inline"`
}

func (DocumentFormateDeleteActivity) CollectionName() string {
	return documentformateCollectionName
}
