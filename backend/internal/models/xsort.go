package models

type XSort struct {
	Code string `json:"code" bson:"code"`
	XOrder uint   `json:"xorder" bson:"xorder" validate:"min=0,max=4294967295"`
}

type XSortModifyReqesut struct {
	GUIDFixed string `json:"guid_fixed" bson:"guid_fixed"`
	Code string `json:"code" bson:"code"`
	XOrder uint   `json:"xorder" bson:"xorder" validate:"min=0,max=4294967295"`
}
