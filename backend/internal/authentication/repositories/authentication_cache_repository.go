package repositories

type IAuthenticationCacheRepository interface {
	Save(keyCache string) error
}

type AuthenticationCacheRepository struct {
}

func NewAuthenticationCacheRepository() *AuthenticationCacheRepository {
	return &AuthenticationCacheRepository{}
}
