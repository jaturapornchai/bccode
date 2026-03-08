package models

import (
	"smlcloudplatform/internal/models"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

const smspaymentsettingsCollectionName = "smsPaymentSettings"

type SmsPaymentSettings struct {
	StorefrontGUID   string `json:"storefrontguid" bson:"storefrontguid" validate:"required,max=233"`
	PatternCode      string `json:"patterncode" bson:"patterncode" validate:"required"`
	TimeMinuteBefore int    `json:"timeminutebefore" bson:"timeminutebefore"`
	TimeMinuteAfter  int    `json:"timeminuteafter" bson:"timeminuteafter"`
}

type SmsPaymentSettingsInfo struct {
	models.DocIdentity `bson:"inline"`
	SmsPaymentSettings `bson:"inline"`
}

func (SmsPaymentSettingsInfo) CollectionName() string {
	return smspaymentsettingsCollectionName
}

type SmsPaymentSettingsData struct {
	models.ShopIdentity    `bson:"inline"`
	SmsPaymentSettingsInfo `bson:"inline"`
}

type SmsPaymentSettingsDoc struct {
	ID                     primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	SmsPaymentSettingsData `bson:"inline"`
	models.ActivityDoc     `bson:"inline"`
}

func (SmsPaymentSettingsDoc) CollectionName() string {
	return smspaymentsettingsCollectionName
}

type SmsPaymentSettingsItemGuid struct {
	DocNo string `json:"docno" bson:"docno"`
}

func (SmsPaymentSettingsItemGuid) CollectionName() string {
	return smspaymentsettingsCollectionName
}

type SmsPaymentSettingsActivity struct {
	SmsPaymentSettingsData `bson:"inline"`
	models.ActivityTime    `bson:"inline"`
}

func (SmsPaymentSettingsActivity) CollectionName() string {
	return smspaymentsettingsCollectionName
}

type SmsPaymentSettingsDeleteActivity struct {
	models.Identity     `bson:"inline"`
	models.ActivityTime `bson:"inline"`
}

func (SmsPaymentSettingsDeleteActivity) CollectionName() string {
	return smspaymentsettingsCollectionName
}
