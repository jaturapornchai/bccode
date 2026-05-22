package models

type ProcessMongoDebtorModel struct {
	ShopId string                        `json:"shopid" bson:"shopid"`
	GuidFixed string                        `json:"guid_fixed" bson:"guid_fixed"`
	Code string                        `json:"code" bson:"code"`
	Names []ProcessMongoDebtorNameModel `json:"names" bson:"names"`
	TaxID string                        `json:"tax_id" bson:"tax_id"`
	PersonalType int8                          `json:"personal_type" bson:"personal_type"`
	CustomerType int32                         `json:"customer_type" bson:"customer_type"`
	BranchNumber string                        `json:"branch_number" bson:"branch_number"`
	FundCode string                        `json:"fund_code" bson:"fund_code"`
	CreditDay int32                         `json:"creditday" bson:"creditday"`
	Email string                        `json:"email" bson:"email"`
	AddressBilling map[string]any                `json:"addressforbilling" bson:"addressforbilling"`
}

type ProcessMongoDebtorNameModel struct {
	Name string `json:"name" bson:"name"`
}
