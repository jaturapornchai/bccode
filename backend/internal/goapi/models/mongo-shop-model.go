package models

type MongoShopModel struct {
	HoldingCode string              `json:"guidfixed" bson:"guidfixed"`
	Names       []languageNameModel `json:"name" bson:"names"`
}
