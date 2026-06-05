package models

type DebtorProcessRequest struct {
	HoldingCode string `json:"holdingcode" `
	DebtorCode  string `json:"debtorcode"  `
}
