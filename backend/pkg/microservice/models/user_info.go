package models

type UserInfo struct {
	Username           string `json:"username" `
	Name               string `json:"name"`
	HoldingCode        string `json:"holdingcode" `
	BusinessCode       string `json:"businesscode"`
	Role               uint8  `json:"role"`
	UID                string `json:"uid"`
	MustChangePassword bool   `json:"mustchangepassword"`
}
