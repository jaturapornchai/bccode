package microservice

import (
	"fmt"
	"net/http"
	"smlcloudplatform/internal/encrypt"
	"smlcloudplatform/pkg/memorycache"
	"smlcloudplatform/pkg/microservice/models"
	"strconv"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
)

type IAuthService interface {
	MWFuncWithRedisMixShop(cacher ICacher, shopPath []string, publicPath ...string) echo.MiddlewareFunc
	MWFuncWithRedis(cacher ICacher, publicPath ...string) echo.MiddlewareFunc
	MWFuncWithShop(cacher ICacher, publicPath ...string) echo.MiddlewareFunc
	GetPrefixCacheKey(tokenType TokenType) string
	GetTokenFromContext(c echo.Context) (*TokenContext, error)
	GetTokenFromAuthorizationHeader(tokenType TokenType, tokenAuthorization string) (string, error)
	GenerateTokenWithRedis(tokenType TokenType, userInfo models.UserInfo) (string, error)
	GenerateTokenWithRedisExpire(tokenType TokenType, userInfo models.UserInfo, expireTime time.Duration) (string, error)
	SelectShop(tokenType TokenType, tokenStr string, shopID string, role uint8) error
	ExpireToken(tokenType TokenType, tokenAuthorizationHeader string) error
	DeleteToken(tokenType TokenType, tokenStr string) error
	RefreshToken(token string) (string, string, error)
}

type TokenType = int

const (
	AUTHTYPE_BEARER TokenType = iota
	AUTHTYPE_WEBSOCKET
	AUTHTYPE_XAPIKEY
	AUTHTYPE_REFRESH
)

type TokenContext struct {
	token     string
	tokenType TokenType
}

type AuthService struct {
	cacheMemoryExpire     time.Duration
	cacheMemory           memorycache.IMemoryCache
	cacher                ICacher
	expireTimeBearer      time.Duration
	prefixBearerCacheKey  string
	prefixBearerToken     string
	expireXApiKey         time.Duration
	prefixXApiKeyCacheKey string
	prefixRefreshCacheKey string
	expireTimeRefresh     time.Duration
	encrypt               encrypt.Encrypt
}

func NewAuthService(cacher ICacher, expireTimeBearer time.Duration, expireTimeRefresh time.Duration) *AuthService {

	return &AuthService{
		cacher:                cacher,
		expireTimeBearer:      expireTimeBearer,
		expireTimeRefresh:     expireTimeRefresh,
		prefixBearerCacheKey:  "auth-",
		prefixBearerToken:     "Bearer",
		prefixXApiKeyCacheKey: "xapikey-",
		prefixRefreshCacheKey: "refresh-",
		encrypt:               *encrypt.NewEncrypt(),
		cacheMemory:           memorycache.NewMemoryCache(),
		cacheMemoryExpire:     time.Duration(5) * time.Second,
	}
}

func NewAuthServicePrefix(authPrefixCache string, authRefreshCache string, cacher ICacher, expireTimeBearer time.Duration, expireTimeRefresh time.Duration) *AuthService {

	return &AuthService{
		cacher:                cacher,
		expireTimeBearer:      expireTimeBearer,
		expireTimeRefresh:     expireTimeRefresh,
		prefixBearerCacheKey:  authPrefixCache,
		prefixBearerToken:     "Bearer",
		prefixXApiKeyCacheKey: "xapikey-",
		prefixRefreshCacheKey: "refresh-",
		encrypt:               *encrypt.NewEncrypt(),
		cacheMemory:           memorycache.NewMemoryCache(),
		cacheMemoryExpire:     time.Duration(5) * time.Second,
	}
}

func (authService *AuthService) MWFuncWithRedisMixShop(cacher ICacher, shopPath []string, publicPath ...string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {

			currentPath := c.Request().URL.Path

			for _, publicPath := range publicPath {
				checkSuffix := strings.HasSuffix(publicPath, "*")
				checkPrefix := strings.HasPrefix(currentPath, publicPath[:len(publicPath)-1])

				if checkSuffix && checkPrefix {
					return next(c)
				} else if currentPath == publicPath {
					return next(c)
				}
			}

			tokenCtx, err := authService.GetTokenFromContext(c)

			if err != nil {
				return c.JSON(http.StatusUnauthorized, map[string]interface{}{"success": false, "message": "Token Invalid."})
			}

			cacheKey := authService.GetPrefixCacheKey(tokenCtx.tokenType) + tokenCtx.token

			tempUserInfo := models.UserInfo{}

			// memTempUserInfo, memExists := authService.cacheMemory.Get(cacheKey)

			// if memExists {
			// 	tempUserInfo = memTempUserInfo.(models.UserInfo)
			// }

			if len(tempUserInfo.Username) < 1 {

				tempUserInfoRaw, err := authService.cacher.HMGet(cacheKey, []string{"username", "name", "shopid", "role"})

				if err != nil {
					return c.JSON(http.StatusUnauthorized, map[string]interface{}{"success": false, "message": "Token Invalid."})
				}

				if tempUserInfoRaw[0] != nil {
					tempUserInfo.Username = fmt.Sprintf("%v", tempUserInfoRaw[0])
					tempUserInfo.Name = fmt.Sprintf("%v", tempUserInfoRaw[1])

					if tempUserInfoRaw[2] != nil {
						tempUserInfo.ShopID = fmt.Sprintf("%v", tempUserInfoRaw[2])
					}
				}

				if tempUserInfoRaw[3] != nil {
					userRole, err := strconv.Atoi(fmt.Sprintf("%v", tempUserInfoRaw[3]))
					tempUserInfo.Role = uint8(userRole)

					if err != nil {
						return c.JSON(http.StatusUnauthorized, map[string]interface{}{"success": false, "message": "Token Invalid."})
					}
				}

			}

			if tempUserInfo.Username == "" {
				return c.JSON(http.StatusUnauthorized, map[string]interface{}{"success": false, "message": "Token Invalid."})
			}

			tempShopID := ""

			if tempUserInfo.ShopID != "" {
				tempShopID = tempUserInfo.ShopID
			}

			// check accept shop path
			thisPathExceptShopSelected := false
			for _, publicPath := range shopPath {
				if currentPath == publicPath {
					thisPathExceptShopSelected = true
				}
			}

			if !thisPathExceptShopSelected && len(string(tempShopID)) < 1 {
				return c.JSON(http.StatusUnauthorized, map[string]interface{}{"success": false, "message": "Shop not selected."})
			}

			userInfo := models.UserInfo{
				Username: tempUserInfo.Username,
				Name:     tempUserInfo.Name,
			}

			if !thisPathExceptShopSelected {
				if len(tempShopID) < 1 {
					return c.JSON(http.StatusUnauthorized, map[string]interface{}{"success": false, "message": "Shop not selected."})
				}

				userInfo.ShopID = tempUserInfo.ShopID
				userInfo.Role = tempUserInfo.Role
			}

			go func() {
				authService.ReTokenExpire(tokenCtx.tokenType, cacheKey)

				if userInfo.ShopID != "" {
					authService.cacheMemory.Set(cacheKey, userInfo, authService.cacheMemoryExpire)
				}
			}()

			c.Set("UserInfo", userInfo)

			return next(c)
		}
	}
}

func (authService *AuthService) MWFuncWithRedis(cacher ICacher, publicPath ...string) echo.MiddlewareFunc {

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {

			currentPath := c.Path()

			for _, publicPath := range publicPath {
				if strings.HasPrefix(currentPath, publicPath) {
					return next(c)
				} else if currentPath == publicPath {
					return next(c)
				}

			}

			tokenCtx, err := authService.GetTokenFromContext(c)
			if err != nil {
				return c.JSON(http.StatusUnauthorized, map[string]interface{}{"success": false, "message": "Token Invalid."})
			}

			cacheKey := authService.GetPrefixCacheKey(tokenCtx.tokenType) + tokenCtx.token

			tempUserInfo, err := authService.cacher.HMGet(cacheKey, []string{"username", "name", "shopid", "role"})

			if err != nil || tempUserInfo[0] == nil {
				return c.JSON(http.StatusUnauthorized, map[string]interface{}{"success": false, "message": "Token Invalid."})
			}

			tempShopID := ""

			if tempUserInfo[2] != nil {
				tempShopID = fmt.Sprintf("%v", tempUserInfo[2])
			}

			if len(string(tempShopID)) < 1 {
				return c.JSON(http.StatusUnauthorized, map[string]interface{}{"success": false, "message": "Shop not selected."})
			}

			userRole, err := strconv.ParseUint(fmt.Sprintf("%v", tempUserInfo[3]), 10, 8)

			if err != nil {
				fmt.Println(err)
				return c.JSON(http.StatusUnauthorized, map[string]interface{}{"success": false, "message": fmt.Sprintf("User role invalid. %v", tempUserInfo[3])})
			}

			userInfo := models.UserInfo{
				Username: fmt.Sprintf("%v", tempUserInfo[0]),
				Name:     fmt.Sprintf("%v", tempUserInfo[1]),
				ShopID:   fmt.Sprintf("%v", tempUserInfo[2]),
				Role:     uint8(userRole),
			}

			authService.ReTokenExpire(tokenCtx.tokenType, cacheKey)
			c.Set("UserInfo", userInfo)

			return next(c)
		}
	}
}

func (authService *AuthService) MWFuncWithShop(cacher ICacher, publicPath ...string) echo.MiddlewareFunc {

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {

			currentPath := c.Path()

			for _, publicPath := range publicPath {
				if currentPath == publicPath {
					return next(c)
				}
			}

			tokenCtx, err := authService.GetTokenFromContext(c)

			if err != nil {
				return c.JSON(http.StatusUnauthorized, map[string]interface{}{"success": false, "message": "Token Invalid."})
			}

			cacheKey := authService.GetPrefixCacheKey(tokenCtx.tokenType) + tokenCtx.token

			tempUserInfo, err := authService.cacher.HMGet(cacheKey, []string{"username", "name"})

			if err != nil {
				return c.JSON(http.StatusUnauthorized, map[string]interface{}{"success": false, "message": "Token Invalid."})
			}

			if tempUserInfo[0] == nil {
				return c.JSON(http.StatusUnauthorized, map[string]interface{}{"success": false, "message": "Token Invalid."})
			}

			userInfo := models.UserInfo{
				Username: fmt.Sprintf("%v", tempUserInfo[0]),
				Name:     fmt.Sprintf("%v", tempUserInfo[1]),
			}

			c.Set("UserInfo", userInfo)

			return next(c)
		}
	}
}

func (authService *AuthService) GetTokenFromContext(c echo.Context) (*TokenContext, error) {

	var rawToken string = ""
	var err error
	var tokenType TokenType = AUTHTYPE_BEARER

	// socket
	if c.IsWebSocket() {
		rawToken = authService.getWebSocketApiKey(c.QueryParam)
		tokenType = AUTHTYPE_WEBSOCKET
	} else {

		// bearer token
		rawToken, err = authService.getBearerToken(c.Request().Header.Get)

		if err != nil {
			rawToken, err = authService.getXApiKeyToken(c.Request().Header.Get)
			tokenType = AUTHTYPE_XAPIKEY

			if err == nil {
				err = nil
			}
		}

	}

	return &TokenContext{
		token:     rawToken,
		tokenType: tokenType,
	}, err
}

func (authService *AuthService) GetPrefixCacheKey(tokenType TokenType) string {
	prefixCacheKey := ""
	if tokenType == AUTHTYPE_BEARER || tokenType == AUTHTYPE_WEBSOCKET {
		prefixCacheKey = authService.prefixBearerCacheKey
	} else if tokenType == AUTHTYPE_XAPIKEY {
		prefixCacheKey = authService.prefixXApiKeyCacheKey
	} else if tokenType == AUTHTYPE_REFRESH {
		prefixCacheKey = authService.prefixRefreshCacheKey
	}

	return prefixCacheKey
}

func (authService *AuthService) getBearerToken(fncGetHeader func(string) string) (string, error) {
	tokenString := fncGetHeader("Authorization")

	if tokenString == "" {
		return "", fmt.Errorf("missing authorization header")
	}

	parts := strings.SplitN(tokenString, " ", 2)
	if !(len(parts) == 2 && parts[0] == authService.prefixBearerToken) {
		return "", fmt.Errorf("missing authorization bearer")
	}

	return parts[1], nil
}

func (authService *AuthService) getWebSocketApiKey(fncQueryParam func(string) string) string {
	return fncQueryParam("apikey")
}

func (authService *AuthService) getXApiKeyToken(fncGetHeader func(string) string) (string, error) {
	tokenString := fncGetHeader("x-api-key")

	if tokenString == "" {
		return "", fmt.Errorf("missing authorization header")
	}

	return strings.TrimSpace(tokenString), nil
}

func (authService *AuthService) GetTokenFromAuthorizationHeader(tokenType TokenType, tokenAuthorization string) (string, error) {

	if tokenType == AUTHTYPE_BEARER {
		if len(tokenAuthorization) < 1 {
			return "", fmt.Errorf("authorization is not empty")
		}

		parts := strings.SplitN(tokenAuthorization, " ", 2)
		if !(len(parts) == 2 && parts[0] == authService.prefixBearerToken) {
			return "", fmt.Errorf("missing authorization bearer")
		}

		return parts[1], nil
	} else {
		return strings.TrimSpace(tokenAuthorization), nil
	}

}

func (authService *AuthService) GenerateTokenWithRedis(tokenType TokenType, userInfo models.UserInfo) (string, error) {

	tokenStr := authService.encrypt.GenerateSHA256Hash(NewUUID())
	cacheKey := authService.GetPrefixCacheKey(tokenType) + tokenStr

	authService.cacher.HMSet(cacheKey, map[string]interface{}{
		"username": userInfo.Username,
		"name":     userInfo.Name,
	})
	authService.SetTokenExpire(tokenType, cacheKey)

	return tokenStr, nil
}

func (authService *AuthService) GenerateTokenWithRedisExpire(tokenType TokenType, userInfo models.UserInfo, expireTime time.Duration) (string, error) {

	tokenStr := authService.encrypt.GenerateSHA256Hash(NewUUID())
	cacheKey := authService.GetPrefixCacheKey(tokenType) + tokenStr

	authService.cacher.HMSet(cacheKey, map[string]interface{}{
		"username": userInfo.Username,
		"name":     userInfo.Name,
		"shopid":   userInfo.ShopID,
		"role":     userInfo.Role,
	})
	authService.cacher.Expire(cacheKey, expireTime)

	return tokenStr, nil
}

func (authService *AuthService) SelectShop(tokenType TokenType, tokenStr string, shopID string, role uint8) error {
	cacheKey := authService.GetPrefixCacheKey(tokenType) + tokenStr

	err := authService.cacher.HMSet(cacheKey, map[string]interface{}{
		"shopid": shopID,
		"role":   role,
	})

	if err != nil {
		return err
	}

	return nil
}

func (authService *AuthService) RefreshToken(token string) (string, string, error) {
	cacheKey := authService.GetPrefixCacheKey(AUTHTYPE_REFRESH) + token

	tempUserInfo, err := authService.cacher.HMGet(cacheKey, []string{"username", "name"})

	if err != nil || tempUserInfo[0] == nil {
		return "", "", err
	}

	userInfo := models.UserInfo{
		Username: fmt.Sprintf("%v", tempUserInfo[0]),
		Name:     fmt.Sprintf("%v", tempUserInfo[1]),
	}

	tokenStr, err := authService.GenerateTokenWithRedis(AUTHTYPE_BEARER, userInfo)

	if err != nil {
		return "", "", err
	}

	refreshTokenStr, err := authService.GenerateTokenWithRedis(AUTHTYPE_REFRESH, userInfo)

	if err != nil {
		return "", "", err
	}

	return tokenStr, refreshTokenStr, err
}

func (authService *AuthService) ReTokenExpire(tokenType TokenType, cacheKey string) {
	if tokenType == AUTHTYPE_BEARER || tokenType == AUTHTYPE_WEBSOCKET {
		authService.cacher.Expire(cacheKey, authService.expireTimeBearer)
	}
}

func (authService *AuthService) SetTokenExpire(tokenType TokenType, cacheKey string) {
	if tokenType == AUTHTYPE_BEARER || tokenType == AUTHTYPE_WEBSOCKET {
		authService.cacher.Expire(cacheKey, authService.expireTimeBearer)
	} else if tokenType == AUTHTYPE_XAPIKEY {
		authService.cacher.Expire(cacheKey, authService.expireXApiKey)
	} else if tokenType == AUTHTYPE_REFRESH {
		authService.cacher.Expire(cacheKey, authService.expireTimeRefresh)
	}
}

func (authService *AuthService) ExpireToken(tokenType TokenType, tokenAuthorizationHeader string) error {
	tokenStr, err := authService.GetTokenFromAuthorizationHeader(tokenType, tokenAuthorizationHeader)
	if err != nil {
		return err
	}
	cacheKey := authService.GetPrefixCacheKey(tokenType) + tokenStr
	authService.cacher.Expire(cacheKey, -1)
	return nil
}

func (authService *AuthService) DeleteToken(tokenType TokenType, tokenStr string) error {
	cacheKey := authService.GetPrefixCacheKey(tokenType) + tokenStr
	authService.cacher.Expire(cacheKey, -1)
	return nil
}
