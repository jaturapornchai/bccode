package models

type LineVerify struct {
	ClientID  string `json:"client_id"`
	ExpiresIn int    `json:"expires_in"`
	Scope     string `json:"scope"`
}

type LineProfile struct {
	UserID        string `json:"userId" `
	DisplayName   string `json:"displayName" `
	StatusMessage string `json:"statusMessage" `
	PictureUrl    string `json:"pictureUrl" `
}

type LineAuthRequest struct {
	ShopID          string `json:"shopid"`
	LineAccessToken string `json:"lineaccesstoken"`
}
