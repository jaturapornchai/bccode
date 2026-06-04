package models

type ProcessMongoCreditorModel struct {
	HoldingCode    string                          `json:"holding_code" bson:"holding_code"`
	GuidFixed      string                          `json:"guid_fixed" bson:"guid_fixed"`
	Code           string                          `json:"code" bson:"code"`
	Names          []ProcessMongoCreditorNameModel `json:"names" bson:"names"`
	TaxID          string                          `json:"tax_id" bson:"tax_id"`
	PersonalType   int8                            `json:"personal_type" bson:"personal_type"`
	CustomerType   int32                           `json:"customer_type" bson:"customer_type"`
	BranchNumber   string                          `json:"branch_number" bson:"branch_number"`
	FundCode       string                          `json:"fund_code" bson:"fund_code"`
	CreditDay      int32                           `json:"creditday" bson:"creditday"`
	Email          string                          `json:"email" bson:"email"`
	AddressBilling map[string]any                  `json:"addressforbilling" bson:"addressforbilling"`
}

type ProcessMongoCreditorNameModel struct {
	Name string `json:"name" bson:"name"`
}
