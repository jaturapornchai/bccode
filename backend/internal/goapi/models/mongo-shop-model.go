package models

type MongoShopModel struct {
	HoldingCode string              `json:"guid_fixed" bson:"guid_fixed"`
	Names       []languageNameModel `json:"name" bson:"names"`
}
