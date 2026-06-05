package models

type ProcessMongoDebtorModel struct {
	HoldingCode    string                        `json:"holdingcode" bson:"holdingcode"`
	GuidFixed      string                        `json:"guidfixed" bson:"guidfixed"`
	Code           string                        `json:"code" bson:"code"`
	Names          []ProcessMongoDebtorNameModel `json:"names" bson:"names"`
	TaxID          string                        `json:"taxid" bson:"taxid"`
	PersonalType   int8                          `json:"personaltype" bson:"personaltype"`
	CustomerType   int32                         `json:"customertype" bson:"customertype"`
	BranchNumber   string                        `json:"branchnumber" bson:"branchnumber"`
	FundCode       string                        `json:"fundcode" bson:"fundcode"`
	CreditDay      int32                         `json:"creditday" bson:"creditday"`
	Email          string                        `json:"email" bson:"email"`
	AddressBilling map[string]any                `json:"addressforbilling" bson:"addressforbilling"`
}

type ProcessMongoDebtorNameModel struct {
	Name string `json:"name" bson:"name"`
}
