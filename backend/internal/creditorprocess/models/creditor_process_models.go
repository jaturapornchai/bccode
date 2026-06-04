package models

type CreditorProcessRequest struct {
	HoldingCode  string `json:"holding_code" `
	CreditorCode string `json:"creditor_code"  `
}
