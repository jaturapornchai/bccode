package models

type UserInfo struct {
	Username string `json:"username" `
	Name     string `json:"name"`
	ShopID   string `json:"shopid" `
	Role     uint8  `json:"role"`
}
