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
	SelectShop(tokenType TokenType, tokenStr string, holdingCode string, businessCode string, role uint8) error
	ExpireToken(tokenType TokenType, tokenAuthorizationHeader string) error
	DeleteToken(tokenType TokenType, tokenStr string) error
	RefreshToken(token string) (string, string, bool, error)
	RevokeUserTokens(username string) error
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

func cacheString(value interface{}) string {
	if value == nil {
		return ""
	}
	return fmt.Sprintf("%v", value)
}

func cacheBool(value interface{}) bool {
	parsed, err := strconv.ParseBool(cacheString(value))
	return err == nil && parsed
}

func passwordChangeAllowed(method string, path string) bool {
	return path == "/profile/password" || path == "/logout" || (method == http.MethodGet && path == "/profile")
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

				tempUserInfoRaw, err := authService.cacher.HMGet(cacheKey, []string{"username", "name", "uid", "holdingcode", "role", "mustchangepassword", "businesscode"})

				if err != nil {
					return c.JSON(http.StatusUnauthorized, map[string]interface{}{"success": false, "message": "Token Invalid."})
				}

				if tempUserInfoRaw[0] != nil {
					tempUserInfo.Username = fmt.Sprintf("%v", tempUserInfoRaw[0])
					tempUserInfo.Name = fmt.Sprintf("%v", tempUserInfoRaw[1])

					if tempUserInfoRaw[2] != nil {
						tempUserInfo.UID = cacheString(tempUserInfoRaw[2])
					}

					if tempUserInfoRaw[3] != nil {
						tempUserInfo.HoldingCode = fmt.Sprintf("%v", tempUserInfoRaw[3])
					}
				}

				if tempUserInfoRaw[4] != nil {
					userRole, err := strconv.Atoi(fmt.Sprintf("%v", tempUserInfoRaw[4]))
					tempUserInfo.Role = uint8(userRole)

					if err != nil {
						return c.JSON(http.StatusUnauthorized, map[string]interface{}{"success": false, "message": "Token Invalid."})
					}
				}
				tempUserInfo.MustChangePassword = cacheBool(tempUserInfoRaw[5])
				if tempUserInfoRaw[6] != nil {
					tempUserInfo.BusinessCode = cacheString(tempUserInfoRaw[6])
				}

			}

			if tempUserInfo.Username == "" {
				return c.JSON(http.StatusUnauthorized, map[string]interface{}{"success": false, "message": "Token Invalid."})
			}
			if tempUserInfo.MustChangePassword && !passwordChangeAllowed(c.Request().Method, currentPath) {
				return c.JSON(http.StatusForbidden, map[string]interface{}{
					"success": false,
					"code":    "password_change_required",
					"message": "กรุณาเปลี่ยนรหัสผ่านเริ่มต้นก่อนใช้งานระบบ",
				})
			}

			tempHoldingCode := ""

			if tempUserInfo.HoldingCode != "" {
				tempHoldingCode = tempUserInfo.HoldingCode
			}

			// check accept shop path
			thisPathExceptShopSelected := false
			for _, publicPath := range shopPath {
				if currentPath == publicPath {
					thisPathExceptShopSelected = true
				}
			}

			if !thisPathExceptShopSelected && len(string(tempHoldingCode)) < 1 {
				return c.JSON(http.StatusUnauthorized, map[string]interface{}{"success": false, "message": "Shop not selected."})
			}

			userInfo := models.UserInfo{
				Username:           tempUserInfo.Username,
				Name:               tempUserInfo.Name,
				UID:                tempUserInfo.UID,
				MustChangePassword: tempUserInfo.MustChangePassword,
			}

			if !thisPathExceptShopSelected {
				if len(tempHoldingCode) < 1 {
					return c.JSON(http.StatusUnauthorized, map[string]interface{}{"success": false, "message": "Shop not selected."})
				}

				userInfo.HoldingCode = tempUserInfo.HoldingCode
				userInfo.BusinessCode = tempUserInfo.BusinessCode
				userInfo.Role = tempUserInfo.Role
			}

			go func() {
				authService.ReTokenExpire(tokenCtx.tokenType, cacheKey)

				if userInfo.HoldingCode != "" {
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

			tempUserInfo, err := authService.cacher.HMGet(cacheKey, []string{"username", "name", "uid", "holdingcode", "role", "mustchangepassword", "businesscode"})

			if err != nil || len(tempUserInfo) < 7 || tempUserInfo[0] == nil {
				return c.JSON(http.StatusUnauthorized, map[string]interface{}{"success": false, "message": "Token Invalid."})
			}
			mustChangePassword := cacheBool(tempUserInfo[5])
			if mustChangePassword && !passwordChangeAllowed(c.Request().Method, currentPath) {
				return c.JSON(http.StatusForbidden, map[string]interface{}{
					"success": false,
					"code":    "password_change_required",
					"message": "กรุณาเปลี่ยนรหัสผ่านเริ่มต้นก่อนใช้งานระบบ",
				})
			}

			tempHoldingCode := ""

			if tempUserInfo[3] != nil {
				tempHoldingCode = fmt.Sprintf("%v", tempUserInfo[3])
			}

			if len(string(tempHoldingCode)) < 1 {
				return c.JSON(http.StatusUnauthorized, map[string]interface{}{"success": false, "message": "Shop not selected."})
			}

			userRole, err := strconv.ParseUint(fmt.Sprintf("%v", tempUserInfo[4]), 10, 8)

			if err != nil {
				fmt.Println(err)
				return c.JSON(http.StatusUnauthorized, map[string]interface{}{"success": false, "message": fmt.Sprintf("User role invalid. %v", tempUserInfo[4])})
			}

			userInfo := models.UserInfo{
				Username:           fmt.Sprintf("%v", tempUserInfo[0]),
				Name:               fmt.Sprintf("%v", tempUserInfo[1]),
				UID:                cacheString(tempUserInfo[2]),
				HoldingCode:        cacheString(tempUserInfo[3]),
				BusinessCode:       cacheString(tempUserInfo[6]),
				Role:               uint8(userRole),
				MustChangePassword: mustChangePassword,
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

			tempUserInfo, err := authService.cacher.HMGet(cacheKey, []string{"username", "name", "uid", "mustchangepassword"})

			if err != nil || len(tempUserInfo) < 4 {
				return c.JSON(http.StatusUnauthorized, map[string]interface{}{"success": false, "message": "Token Invalid."})
			}

			if tempUserInfo[0] == nil {
				return c.JSON(http.StatusUnauthorized, map[string]interface{}{"success": false, "message": "Token Invalid."})
			}
			mustChangePassword := cacheBool(tempUserInfo[3])
			if mustChangePassword && !passwordChangeAllowed(c.Request().Method, currentPath) {
				return c.JSON(http.StatusForbidden, map[string]interface{}{
					"success": false,
					"code":    "password_change_required",
					"message": "กรุณาเปลี่ยนรหัสผ่านเริ่มต้นก่อนใช้งานระบบ",
				})
			}

			userInfo := models.UserInfo{
				Username:           fmt.Sprintf("%v", tempUserInfo[0]),
				Name:               fmt.Sprintf("%v", tempUserInfo[1]),
				UID:                cacheString(tempUserInfo[2]),
				MustChangePassword: mustChangePassword,
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
		"username":           userInfo.Username,
		"name":               userInfo.Name,
		"uid":                userInfo.UID,
		"mustchangepassword": userInfo.MustChangePassword,
	})
	authService.SetTokenExpire(tokenType, cacheKey)

	return tokenStr, nil
}

func (authService *AuthService) GenerateTokenWithRedisExpire(tokenType TokenType, userInfo models.UserInfo, expireTime time.Duration) (string, error) {

	tokenStr := authService.encrypt.GenerateSHA256Hash(NewUUID())
	cacheKey := authService.GetPrefixCacheKey(tokenType) + tokenStr

	authService.cacher.HMSet(cacheKey, map[string]interface{}{
		"username":           userInfo.Username,
		"name":               userInfo.Name,
		"uid":                userInfo.UID,
		"holdingcode":        userInfo.HoldingCode,
		"businesscode":       userInfo.BusinessCode,
		"role":               userInfo.Role,
		"mustchangepassword": userInfo.MustChangePassword,
	})
	authService.cacher.Expire(cacheKey, expireTime)

	return tokenStr, nil
}

func (authService *AuthService) SelectShop(tokenType TokenType, tokenStr string, holdingCode string, businessCode string, role uint8) error {
	cacheKey := authService.GetPrefixCacheKey(tokenType) + tokenStr

	err := authService.cacher.HMSet(cacheKey, map[string]interface{}{
		"holdingcode":  holdingCode,
		"businesscode": businessCode,
		"role":         role,
	})

	if err != nil {
		return err
	}

	return nil
}

func (authService *AuthService) RefreshToken(token string) (string, string, bool, error) {
	cacheKey := authService.GetPrefixCacheKey(AUTHTYPE_REFRESH) + token

	tempUserInfo, err := authService.cacher.HMGet(cacheKey, []string{"username", "name", "uid", "mustchangepassword"})

	if err != nil {
		return "", "", false, err
	}
	if len(tempUserInfo) < 4 || tempUserInfo[0] == nil {
		return "", "", false, fmt.Errorf("refresh token invalid")
	}

	userInfo := models.UserInfo{
		Username:           fmt.Sprintf("%v", tempUserInfo[0]),
		Name:               fmt.Sprintf("%v", tempUserInfo[1]),
		UID:                cacheString(tempUserInfo[2]),
		MustChangePassword: cacheBool(tempUserInfo[3]),
	}

	tokenStr, err := authService.GenerateTokenWithRedis(AUTHTYPE_BEARER, userInfo)

	if err != nil {
		return "", "", false, err
	}

	refreshTokenStr, err := authService.GenerateTokenWithRedis(AUTHTYPE_REFRESH, userInfo)

	if err != nil {
		return "", "", false, err
	}

	return tokenStr, refreshTokenStr, userInfo.MustChangePassword, nil
}

func (authService *AuthService) RevokeUserTokens(username string) error {
	username = strings.TrimSpace(username)
	if username == "" {
		return fmt.Errorf("username invalid")
	}

	keysToDelete := make([]string, 0)
	for _, prefix := range []string{authService.prefixBearerCacheKey, authService.prefixRefreshCacheKey, authService.prefixXApiKeyCacheKey} {
		keys, err := authService.cacher.Keys(prefix + "*")
		if err != nil {
			return err
		}
		for _, key := range keys {
			cachedUsername, err := authService.cacher.HGet(key, "username")
			if err != nil {
				return err
			}
			if !strings.EqualFold(strings.TrimSpace(cachedUsername), username) {
				continue
			}
			// Block immediately even if Redis deletion fails after this write.
			if err := authService.cacher.HMSet(key, map[string]interface{}{"mustchangepassword": true}); err != nil {
				return err
			}
			authService.cacheMemory.Delete(key)
			keysToDelete = append(keysToDelete, key)
		}
	}
	if len(keysToDelete) == 0 {
		return nil
	}
	return authService.cacher.Del(keysToDelete...)
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
