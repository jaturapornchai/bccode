package firebase

type UserInfo struct {
	SignInProvider string `json:"sign_in_provider"`
	Email          string `json:"email"`
	UserId         string `json:"user_id"`
	Name           string `json:"name"`
}
