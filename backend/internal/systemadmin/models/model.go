package models

type RequestReSyncTenant struct {
	HoldingCode string `json:"holding_code"`
}

type RequestReSyncTenantByDate struct {
	HoldingCode string `json:"holding_code"`
	Date        string `json:"date"`
}
