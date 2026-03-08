package models

type ProcessMongoDebtorModel struct {
	ShopId string                        `json:"shopid" bson:"shopid"`
	Code   string                        `json:"code" bson:"code"`
	Names  []ProcessMongoDebtorNameModel `json:"names" bson:"names"`
	TaxID  string                        `json:"taxid" bson:"taxid"`
}

type ProcessMongoDebtorNameModel struct {
	Name string `json:"name" bson:"name"`
}
