package models

type MongoShopModel struct {
	HoldingCode string              `json:"guidfixed" bson:"guidfixed"`
	Code        string              `json:"code" bson:"code"`
	Names       []languageNameModel `json:"name" bson:"names"`
}
