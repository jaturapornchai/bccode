package models

type MongoShopModel struct {
	ShopId string              `json:"guidfixed" bson:"guidfixed"`
	Names  []languageNameModel `json:"name" bson:"names"`
}
