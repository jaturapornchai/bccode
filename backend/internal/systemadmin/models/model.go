package models

type RequestReSyncTenant struct {
	HoldingCode string `json:"holdingcode"`
}

type RequestReSyncTenantByDate struct {
	HoldingCode string `json:"holdingcode"`
	Date        string `json:"date"`
}
