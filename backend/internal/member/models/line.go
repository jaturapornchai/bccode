package models

type LineVerify struct {
	ClientID string `json:"client_id"`
	ExpiresIn int    `json:"expires_in"`
	Scope string `json:"scope"`
}

type LineProfile struct {
	UserID string `json:"user_id" `
	DisplayName string `json:"display_name" `
	StatusMessage string `json:"status_message" `
	PictureUrl string `json:"picture_url" `
}

type LineAuthRequest struct {
	ShopID string `json:"shopid"`
	LineAccessToken string `json:"lineaccesstoken"`
}
