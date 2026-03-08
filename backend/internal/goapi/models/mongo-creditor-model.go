package models

type ProcessMongoCreditorModel struct {
	ShopId string                          `json:"shopid" bson:"shopid"`
	Code   string                          `json:"code" bson:"code"`
	Names  []ProcessMongoCreditorNameModel `json:"names" bson:"names"`
	TaxID  string                          `json:"taxid" bson:"taxid"`
}

type ProcessMongoCreditorNameModel struct {
	Name string `json:"name" bson:"name"`
}
