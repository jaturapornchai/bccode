package models

type ProcessMongoWarehouseModel struct {
	HoldingCode string                               `json:"holdingcode" bson:"holdingcode"`
	Code        string                               `json:"code" bson:"code"`
	Names       []LanguageModel                      `json:"names" bson:"names"`
	Location    []ProcessMongoWarehouseLocationModel `json:"location" bson:"location"`
}

type ProcessMongoWarehouseLocationModel struct {
	Code  string          `json:"code" bson:"code"`
	Names []LanguageModel `json:"names" bson:"names"`
}
