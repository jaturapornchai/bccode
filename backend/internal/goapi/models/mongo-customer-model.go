package models

type ProcessMongoCustomerModel struct {
	ShopId string                          `json:"shopid" bson:"shopid"`
	Code string                          `json:"code" bson:"code"`
	PersonalType int                             `json:"personal_type" bson:"personal_type"`
	Names []ProcessMongoCustomerNameModel `json:"names" bson:"names"`
	TaxID string                          `json:"tax_id" bson:"tax_id"`
	CustomerType int                             `json:"customer_type" bson:"customer_type"`
}

type ProcessMongoCustomerNameModel struct {
	Name string `json:"name" bson:"name"`
}
