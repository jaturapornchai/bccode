package microservice

import (
	"fmt"
	"net/http"
	"smlcloudplatform/pkg/microservice/models"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/labstack/echo/v4"
)

type CustomClaims struct {
	*jwt.RegisteredClaims
	models.UserInfo
}

// func NewJwtService(signKey *rsa.PrivateKey, verifyKey *rsa.PublicKey, expireMinute int) *JwtService {

// 	return &JwtService{
// 		signKey:   signKey,
// 		verifyKey: verifyKey,
// 		duration:  time.Duration(expireMinute) * time.Minute,
// 	}
// }

func NewJwtService(cacher ICacher, jwtSecretKey string, expireHour int) *JwtService {

	return &JwtService{
		cacher:              cacher,
		jwtSecretKey:        jwtSecretKey,
		expire:              time.Duration(expireHour) * time.Hour,
		prefixCacheKey:      "auth-",
		prefixAuthorization: "Bearer",
	}
}

// Service provides a Json-Web-Token authentication implementation
// type JwtService struct {
// 	// Secret key used for signing.
// 	signKey   *rsa.PrivateKey
// 	verifyKey *rsa.PublicKey

// 	// Duration for which the jwt token is valid.
// 	duration time.Duration
// }

type JwtService struct {
	cacher              ICacher
	jwtSecretKey        string
	expire              time.Duration
	prefixCacheKey      string
	prefixAuthorization string
}

// MWFunc makes JWT implement the Middleware interface.
func (jwtService *JwtService) MWFunc() echo.MiddlewareFunc {

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {

			publicPathList := []string{
				"/login",
				"/poslogin",
				"/register",
			}

			currentPath := c.Path()

			for _, publicPath := range publicPathList {
				if currentPath == publicPath {
					return next(c)
				}
			}

			token, err := jwtService.ParseTokenFromContext(c)

			if err != nil || !token.Valid {

				return c.JSON(http.StatusUnauthorized, map[string]interface{}{"success": false, "message": "Unauthorized."})
			}

			claims := token.Claims.(*CustomClaims)

			c.Set("UserInfo", claims.UserInfo)

			return next(c)
		}
	}
}

func (jwtService *JwtService) MWFuncWithRedis(cacher ICacher, publicPath ...string) echo.MiddlewareFunc {

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {

			currentPath := c.Path()

			for _, publicPath := range publicPath {
				if currentPath == publicPath {
					return next(c)
				}
			}

			tokenStr, err := jwtService.GetTokenFromContext(c)

			if err != nil {
				return c.JSON(http.StatusUnauthorized, map[string]interface{}{"success": false, "message": "Token Invalid."})
			}

			cacheKey := jwtService.prefixCacheKey + tokenStr
			tempUserInfo, err := jwtService.cacher.HMGet(cacheKey, []string{"username", "name", "shopid"})

			if err != nil {
				return c.JSON(http.StatusUnauthorized, map[string]interface{}{"success": false, "message": "Token Invalid."})
			}
			tempShopID := ""

			if tempUserInfo[2] != nil {
				tempShopID = fmt.Sprintf("%v", tempUserInfo[2])
			}

			if len(string(tempShopID)) < 1 {
				return c.JSON(http.StatusUnauthorized, map[string]interface{}{"success": false, "message": "Shop not selected."})
			}

			userInfo := models.UserInfo{
				Username: fmt.Sprintf("%v", tempUserInfo[0]),
				Name:     fmt.Sprintf("%v", tempUserInfo[1]),
				ShopID:   fmt.Sprintf("%v", tempUserInfo[2]),
			}

			cacher.Expire("auth-"+tokenStr, jwtService.expire)

			c.Set("UserInfo", userInfo)

			return next(c)
		}
	}
}

func (jwtService *JwtService) MWFuncWithShop(cacher ICacher) echo.MiddlewareFunc {

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {

			tokenStr, err := jwtService.GetTokenFromContext(c)

			if err != nil {
				return c.JSON(http.StatusUnauthorized, map[string]interface{}{"success": false, "message": "Token Invalid."})
			}

			cacheKey := jwtService.prefixCacheKey + tokenStr
			tempUserInfo, err := jwtService.cacher.HMGet(cacheKey, []string{"username", "name"})

			if err != nil {
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

func (jwtService *JwtService) GetPrefixCacheKey() string {
	return jwtService.prefixCacheKey
}

func (jwtService *JwtService) GetTokenFromContext(c echo.Context) (string, error) {
	tokenString := c.Request().Header.Get("Authorization")
	if tokenString == "" {
		return "", fmt.Errorf("missing authorization header")
	}

	parts := strings.SplitN(tokenString, " ", 2)
	if !(len(parts) == 2 && parts[0] == jwtService.prefixAuthorization) {
		return "", fmt.Errorf("missing authorization bearer")
	}

	return parts[1], nil
}

func (jwtService *JwtService) GetTokenFromAuthorizationHeader(tokenAuthorization string) (string, error) {

	parts := strings.SplitN(tokenAuthorization, " ", 2)
	if !(len(parts) == 2 && parts[0] == jwtService.prefixAuthorization) {
		return "", fmt.Errorf("missing authorization bearer")
	}

	return parts[1], nil
}

// ParseToken parses token from Authorization header
func (jwtService *JwtService) ParseTokenFromContext(c echo.Context) (*jwt.Token, error) {

	tokenString := c.Request().Header.Get("Authorization")
	if tokenString == "" {
		return nil, fmt.Errorf("missing authorization header")
	}

	parts := strings.SplitN(tokenString, " ", 2)
	if !(len(parts) == 2 && parts[0] == jwtService.prefixAuthorization) {
		return nil, fmt.Errorf("missing authorization bearer")
	}

	return jwtService.ParseToken(parts[1])
}

func (jwtService *JwtService) ParseToken(tokenString string) (*jwt.Token, error) {
	/*
		token, err := jwt.ParseWithClaims(tokenString, &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
			// since we only use the one private key to sign the tokens,
			// we also only use its public counter part to verify
			return jwtService.verifyKey, nil
		})

		if err != nil {
			return nil, err
		}

		return token, nil
	*/

	token, err := jwt.ParseWithClaims(tokenString, &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(jwtService.jwtSecretKey), nil
	})

	if err != nil {
		return nil, err
	}

	return token, nil
}

func (jwtService *JwtService) GenerateToken(userInfo models.UserInfo) (string, error) {
	/*
		t := jwt.New(jwt.GetSigningMethod("RS256"))

		// set claims
		t.Claims = &CustomClaims{
			&jwt.RegisteredClaims{
				// set the expire time
				// ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Minute * 1)),
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(expire)),
			},
			userInfo,
		}

		// Creat token string
		return t.SignedString(jwtService.signKey)
	*/
	claims := &CustomClaims{
		&jwt.RegisteredClaims{
			// set the expire time
			// ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Minute * 1)),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(jwtService.expire)),
		},
		userInfo,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString([]byte(jwtService.jwtSecretKey))
}

func (jwtService *JwtService) GenerateTokenWithRedis(userInfo models.UserInfo) (string, error) {

	claims := &CustomClaims{
		&jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(jwtService.expire)),
		},
		userInfo,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenStr, err := token.SignedString([]byte(jwtService.jwtSecretKey))

	if err != nil {
		return "", err
	}

	cacheKey := jwtService.prefixCacheKey + tokenStr
	jwtService.cacher.HMSet(cacheKey, map[string]interface{}{
		"username": userInfo.Username,
		"name":     userInfo.Name,
	})

	return tokenStr, nil
}

func (jwtService *JwtService) SelectShop(tokenStr string, shopID string) error {
	cacheKey := jwtService.prefixCacheKey + tokenStr
	err := jwtService.cacher.HMSet(cacheKey, map[string]interface{}{
		"shopid": shopID,
	})

	if err != nil {
		return err
	}

	return nil

}

func (jwtService *JwtService) ExpireToken(tokenAuthorizationHeader string) error {
	tokenStr, err := jwtService.GetTokenFromAuthorizationHeader(tokenAuthorizationHeader)
	if err != nil {
		return err
	}
	cacheKey := jwtService.prefixCacheKey + tokenStr
	jwtService.cacher.Expire(cacheKey, -1)
	return nil
}
