package models

type XSort struct {
	Code   string `json:"code"`
	XOrder uint   `json:"xorder" validate:"min=0,max=4294967295"`
}

type XSortModifyReqesut struct {
	GUIDFixed string `json:"guidfixed"`
	Code      string `json:"code"`
	XOrder    uint   `json:"xorder" validate:"min=0,max=4294967295"`
}
