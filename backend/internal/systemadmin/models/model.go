package models

type RequestReSyncTenant struct {
	ShopID string `json:"shopid"`
}

type RequestReSyncTenantByDate struct {
	ShopID string `json:"shopid"`
	Date   string `json:"date"`
}
