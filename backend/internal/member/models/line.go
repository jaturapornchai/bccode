package models

type LineVerify struct {
	ClientID  string `json:"clientid"`
	ExpiresIn int    `json:"expiresin"`
	Scope     string `json:"scope"`
}

type LineProfile struct {
	UserID        string `json:"userid" `
	DisplayName   string `json:"displayname" `
	StatusMessage string `json:"statusmessage" `
	PictureUrl    string `json:"pictureurl" `
}

type LineAuthRequest struct {
	HoldingCode     string `json:"holdingcode"`
	LineAccessToken string `json:"lineaccesstoken"`
}
