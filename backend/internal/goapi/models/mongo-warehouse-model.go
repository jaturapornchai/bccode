package models

type ProcessMongoWarehouseModel struct {
	Shopid   string                               `json:"shopid" bson:"shopid"`
	Code     string                               `json:"code" bson:"code"`
	Names    []LanguageModel                      `json:"names" bson:"names"`
	Location []ProcessMongoWarehouseLocationModel `json:"location" bson:"location"`
}

type ProcessMongoWarehouseLocationModel struct {
	Code  string          `json:"code" bson:"code"`
	Names []LanguageModel `json:"names" bson:"names"`
}
