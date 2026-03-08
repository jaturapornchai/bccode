package models

type ProcessMongoCustomerModel struct {
	ShopId       string                          `json:"shopid" bson:"shopid"`
	Code         string                          `json:"code" bson:"code"`
	PersonalType int                             `json:"personaltype" bson:"personaltype"`
	Names        []ProcessMongoCustomerNameModel `json:"names" bson:"names"`
	TaxID        string                          `json:"taxid" bson:"taxid"`
	CustomerType int                             `json:"customertype" bson:"customertype"`
}

type ProcessMongoCustomerNameModel struct {
	Name string `json:"name" bson:"name"`
}
