package models

type DebtorProcessRequest struct {
	HoldingCode string `json:"holding_code" `
	DebtorCode  string `json:"debtor_code"  `
}
