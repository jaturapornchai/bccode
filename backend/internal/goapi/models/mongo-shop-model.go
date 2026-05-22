package models

type MongoShopModel struct {
	ShopId string              `json:"guid_fixed" bson:"guid_fixed"`
	Names []languageNameModel `json:"name" bson:"names"`
}
