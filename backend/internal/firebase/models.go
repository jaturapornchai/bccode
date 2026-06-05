package firebase

type UserInfo struct {
	SignInProvider string `json:"signinprovider"`
	Email          string `json:"email"`
	UserId         string `json:"userid"`
	Name           string `json:"name"`
}
