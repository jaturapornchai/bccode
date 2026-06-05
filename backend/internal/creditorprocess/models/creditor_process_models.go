package models

type CreditorProcessRequest struct {
	HoldingCode  string `json:"holdingcode" `
	CreditorCode string `json:"creditorcode"  `
}
