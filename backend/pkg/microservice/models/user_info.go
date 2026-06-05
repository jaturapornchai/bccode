package models

type UserInfo struct {
	Username    string `json:"username" `
	Name        string `json:"name"`
	HoldingCode string `json:"holdingcode" `
	Role        uint8  `json:"role"`
	UID         string `json:"uid"`
}
