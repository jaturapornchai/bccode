package line

type UserInfo struct {
	UserId        string `json:"userid"`
	DisplayName   string `json:"displayname"`
	PictureUrl    string `json:"pictureurl"`
	StatusMessage string `json:"statusmessage"`
}
