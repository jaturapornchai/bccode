package microservice

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"smlcloudplatform/internal/encrypt"
	"smlcloudplatform/pkg/memorycache"
	"smlcloudplatform/pkg/microservice/models"
	"sort"
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
	CreateSession(userInfo models.UserInfo) (string, string, error)
	SelectShop(tokenType TokenType, tokenStr string, holdingCode string, businessCode string, branchUID string, role uint8) error
	ExpireToken(tokenType TokenType, tokenAuthorizationHeader string) error
	DeleteToken(tokenType TokenType, tokenStr string) error
	RefreshToken(token string) (string, string, bool, error)
	RevokeSession(tokenAuthorizationHeader string) error
	RevokeUserTokens(username string) error
	RevokeUserTokensByUID(userUID string) error
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
	cacheMemoryExpire       time.Duration
	cacheMemory             memorycache.IMemoryCache
	cacher                  ICacher
	expireTimeBearer        time.Duration
	prefixBearerCacheKey    string
	prefixBearerToken       string
	expireXApiKey           time.Duration
	prefixXApiKeyCacheKey   string
	prefixRefreshCacheKey   string
	prefixSessionCacheKey   string
	prefixRevokedSessionKey string
	prefixUsedRefreshKey    string
	expireTimeRefresh       time.Duration
	sessionIdleTimeout      time.Duration
	sessionAbsoluteMax      time.Duration
	allowLegacyToken        bool
	timeNow                 func() time.Time
	encrypt                 encrypt.Encrypt
	authorization           *liveAuthorization
}

const (
	accessTokenMaxAge = 15 * time.Minute
	sessionIdleMaxAge = 12 * time.Hour
	sessionMaxAge     = 12 * time.Hour
)

func cacheString(value interface{}) string {
	if value == nil {
		return ""
	}
	return fmt.Sprintf("%v", value)
}

func NewAuthService(cacher ICacher, expireTimeBearer time.Duration, expireTimeRefresh time.Duration, authorizationFinders ...AuthorizationFinder) *AuthService {
	expireTimeBearer = boundedDuration(expireTimeBearer, accessTokenMaxAge)
	expireTimeRefresh = boundedDuration(expireTimeRefresh, sessionMaxAge)

	authService := &AuthService{
		cacher:                  cacher,
		expireTimeBearer:        expireTimeBearer,
		expireTimeRefresh:       expireTimeRefresh,
		prefixBearerCacheKey:    "auth-",
		prefixBearerToken:       "Bearer",
		prefixXApiKeyCacheKey:   "xapikey-",
		prefixRefreshCacheKey:   "refresh-",
		prefixSessionCacheKey:   "session-",
		prefixRevokedSessionKey: "session-revoked-",
		prefixUsedRefreshKey:    "refresh-used-",
		sessionIdleTimeout:      sessionIdleMaxAge,
		sessionAbsoluteMax:      sessionMaxAge,
		timeNow:                 time.Now,
		encrypt:                 *encrypt.NewEncrypt(),
		cacheMemory:             memorycache.NewMemoryCache(),
		cacheMemoryExpire:       time.Duration(5) * time.Second,
	}
	if len(authorizationFinders) > 0 {
		authService.authorization = newLiveAuthorization(authorizationFinders[0])
	}
	return authService
}

func NewAuthServicePrefix(authPrefixCache string, authRefreshCache string, cacher ICacher, expireTimeBearer time.Duration, expireTimeRefresh time.Duration, authorizationFinders ...AuthorizationFinder) *AuthService {
	expireTimeBearer = boundedDuration(expireTimeBearer, accessTokenMaxAge)
	expireTimeRefresh = boundedDuration(expireTimeRefresh, sessionMaxAge)

	authService := &AuthService{
		cacher:                  cacher,
		expireTimeBearer:        expireTimeBearer,
		expireTimeRefresh:       expireTimeRefresh,
		prefixBearerCacheKey:    authPrefixCache,
		prefixBearerToken:       "Bearer",
		prefixXApiKeyCacheKey:   "xapikey-",
		prefixRefreshCacheKey:   authRefreshCache,
		prefixSessionCacheKey:   "session-",
		prefixRevokedSessionKey: "session-revoked-",
		prefixUsedRefreshKey:    "refresh-used-",
		sessionIdleTimeout:      sessionIdleMaxAge,
		sessionAbsoluteMax:      sessionMaxAge,
		timeNow:                 time.Now,
		encrypt:                 *encrypt.NewEncrypt(),
		cacheMemory:             memorycache.NewMemoryCache(),
		cacheMemoryExpire:       time.Duration(5) * time.Second,
	}
	if len(authorizationFinders) > 0 {
		authService.authorization = newLiveAuthorization(authorizationFinders[0])
	}
	return authService
}

// NewLegacyAuthServicePrefix is limited to the separate LINE member service,
// whose existing client receives only a long-lived access token and has no
// refresh endpoint. Core user authentication must use NewAuthService so every
// bearer belongs to a revocable server-side session.
func NewLegacyAuthServicePrefix(authPrefixCache string, authRefreshCache string, cacher ICacher, expireTimeBearer time.Duration, expireTimeRefresh time.Duration) *AuthService {
	authService := NewAuthServicePrefix(authPrefixCache, authRefreshCache, cacher, expireTimeBearer, expireTimeRefresh)
	authService.allowLegacyToken = true
	return authService
}

func boundedDuration(value time.Duration, maximum time.Duration) time.Duration {
	if value <= 0 || value > maximum {
		return maximum
	}
	return value
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

				tempUserInfoRaw, err := authService.cacher.HMGet(cacheKey, []string{"username", "name", "uid", "holdingcode", "role", "businesscode", "sessionuid"})

				if err != nil || len(tempUserInfoRaw) < 7 {
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
				if tempUserInfoRaw[5] != nil {
					tempUserInfo.BusinessCode = cacheString(tempUserInfoRaw[5])
				}
				tempUserInfo.SessionUID = cacheString(tempUserInfoRaw[6])

			}

			if strings.TrimSpace(tempUserInfo.UID) == "" {
				return c.JSON(http.StatusUnauthorized, map[string]interface{}{"success": false, "message": "Token Invalid."})
			}

			// check accept shop path
			thisPathExceptShopSelected := false
			for _, publicPath := range shopPath {
				if currentPath == publicPath {
					thisPathExceptShopSelected = true
				}
			}
			authorized, authErr := authService.authenticateCachedAccess(c.Request().Context(), tokenCtx.tokenType, cacheKey, tempUserInfo)
			if authErr != nil {
				if errors.Is(authErr, ErrLiveWorkspaceAccess) && thisPathExceptShopSelected {
					authorized = loginOnlyUserInfo(tempUserInfo)
				} else if errors.Is(authErr, ErrLiveWorkspaceAccess) {
					return c.JSON(http.StatusForbidden, map[string]interface{}{"success": false, "message": "Workspace access changed."})
				} else {
					return c.JSON(http.StatusUnauthorized, map[string]interface{}{"success": false, "message": "Session expired."})
				}
			}

			if !thisPathExceptShopSelected && strings.TrimSpace(authorized.HoldingCode) == "" {
				return c.JSON(http.StatusUnauthorized, map[string]interface{}{"success": false, "message": "Shop not selected."})
			}

			userInfo := authorized
			if thisPathExceptShopSelected {
				userInfo = loginOnlyUserInfo(authorized)
			}

			go func() {
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

			tempUserInfo, err := authService.cacher.HMGet(cacheKey, []string{"username", "name", "uid", "holdingcode", "role", "businesscode", "sessionuid"})

			if err != nil || len(tempUserInfo) < 7 || strings.TrimSpace(cacheString(tempUserInfo[2])) == "" {
				return c.JSON(http.StatusUnauthorized, map[string]interface{}{"success": false, "message": "Token Invalid."})
			}
			identity := models.UserInfo{
				Username:   cacheString(tempUserInfo[0]),
				Name:       cacheString(tempUserInfo[1]),
				UID:        cacheString(tempUserInfo[2]),
				SessionUID: cacheString(tempUserInfo[6]),
			}
			userInfo, authErr := authService.authenticateCachedAccess(c.Request().Context(), tokenCtx.tokenType, cacheKey, identity)
			if errors.Is(authErr, ErrLiveWorkspaceAccess) {
				return c.JSON(http.StatusForbidden, map[string]interface{}{"success": false, "message": "Workspace access changed."})
			}
			if authErr != nil {
				return c.JSON(http.StatusUnauthorized, map[string]interface{}{"success": false, "message": "Session expired."})
			}
			if strings.TrimSpace(userInfo.HoldingCode) == "" {
				return c.JSON(http.StatusUnauthorized, map[string]interface{}{"success": false, "message": "Shop not selected."})
			}

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

			tempUserInfo, err := authService.cacher.HMGet(cacheKey, []string{"username", "name", "uid", "sessionuid"})

			if err != nil || len(tempUserInfo) < 4 {
				return c.JSON(http.StatusUnauthorized, map[string]interface{}{"success": false, "message": "Token Invalid."})
			}

			if strings.TrimSpace(cacheString(tempUserInfo[2])) == "" {
				return c.JSON(http.StatusUnauthorized, map[string]interface{}{"success": false, "message": "Token Invalid."})
			}
			identity := models.UserInfo{
				Username:   fmt.Sprintf("%v", tempUserInfo[0]),
				Name:       fmt.Sprintf("%v", tempUserInfo[1]),
				UID:        cacheString(tempUserInfo[2]),
				SessionUID: cacheString(tempUserInfo[3]),
			}
			userInfo, authErr := authService.authenticateCachedAccess(c.Request().Context(), tokenCtx.tokenType, cacheKey, identity)
			if errors.Is(authErr, ErrLiveWorkspaceAccess) {
				userInfo = loginOnlyUserInfo(identity)
			} else if authErr != nil {
				return c.JSON(http.StatusUnauthorized, map[string]interface{}{"success": false, "message": "Session expired."})
			} else {
				userInfo = loginOnlyUserInfo(userInfo)
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
		"uid":      userInfo.UID,
	})
	authService.SetTokenExpire(tokenType, cacheKey)

	return tokenStr, nil
}

func (authService *AuthService) GenerateTokenWithRedisExpire(tokenType TokenType, userInfo models.UserInfo, expireTime time.Duration) (string, error) {

	tokenStr := authService.encrypt.GenerateSHA256Hash(NewUUID())
	cacheKey := authService.GetPrefixCacheKey(tokenType) + tokenStr

	authService.cacher.HMSet(cacheKey, map[string]interface{}{
		"username":     userInfo.Username,
		"name":         userInfo.Name,
		"uid":          userInfo.UID,
		"holdingcode":  userInfo.HoldingCode,
		"businesscode": userInfo.BusinessCode,
		"role":         userInfo.Role,
	})
	authService.cacher.Expire(cacheKey, expireTime)

	return tokenStr, nil
}

func (authService *AuthService) CreateSession(userInfo models.UserInfo) (string, string, error) {
	if authService.authorization != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		resolved, err := authService.authorization.Authorize(ctx, loginOnlyUserInfo(userInfo))
		if err != nil {
			return "", "", err
		}
		userInfo = resolved
	}
	now := authService.timeNow().UTC()
	sessionUID := authService.encrypt.GenerateSHA256Hash(NewUUID())
	userInfo.SessionUID = sessionUID

	accessToken, refreshToken, accessKey, refreshKey, err := authService.issueSessionTokens(userInfo, authService.sessionAbsoluteMax)
	if err != nil {
		return "", "", err
	}

	sessionKey := authService.prefixSessionCacheKey + sessionUID
	err = authService.cacher.HMSet(sessionKey, map[string]interface{}{
		"sessionuid":        sessionUID,
		"username":          userInfo.Username,
		"name":              userInfo.Name,
		"createdat":         now.UnixMilli(),
		"lastseenat":        now.UnixMilli(),
		"accesskey":         accessKey,
		"refreshkey":        refreshKey,
		"holdingcode":       userInfo.HoldingCode,
		"businesscode":      userInfo.BusinessCode,
		"role":              userInfo.Role,
		"membershipuid":     userInfo.MembershipUID,
		"holdinguid":        userInfo.HoldingUID,
		"companyuid":        userInfo.CompanyUID,
		"branchuid":         userInfo.BranchUID,
		"permissionversion": userInfo.PermissionVersion,
		"revoked":           false,
	})
	if err == nil {
		err = authService.cacher.Expire(sessionKey, authService.sessionAbsoluteMax)
	}
	if err != nil {
		_ = authService.cacher.Del(accessKey, refreshKey, sessionKey)
		return "", "", err
	}

	return accessToken, refreshToken, nil
}

// SessionStat สรุปเซสชันที่ยังมีชีวิตของหนึ่งกลุ่มกิจการ
type SessionStat struct {
	HoldingCode string `json:"holdingcode"`
	Sessions    int    `json:"sessions"`
	Active      int    `json:"active"`
	LastSeenAt  int64  `json:"lastseenat"`
}

// SessionEntry — ผู้ใช้หนึ่งคน (distinct ตามชื่อ) รวมทุกเซสชันที่ยังไม่หมดอายุ
// — Sessions = จำนวนเซสชันที่รวมมา, CreatedAt = เข้าใช้ล่าสุด,
// LastSeenAt = ใช้งานล่าสุดของทุกเซสชันที่รวม
type SessionEntry struct {
	Username    string `json:"username"`
	Name        string `json:"name"`
	HoldingCode string `json:"holdingcode"`
	Role        string `json:"role"`
	CreatedAt   int64  `json:"createdat"`
	LastSeenAt  int64  `json:"lastseenat"`
	Sessions    int    `json:"sessions"`
	Active      bool   `json:"active"`
}

// SessionStats คือ payload ของ GET /sessions/active-count
// Active = lastseenat อยู่ใน session idle window (30 นาที) — เซสชันที่
// "กำลังใช้งาน" จริง ส่วน Total นับทุกเซสชันที่ยังไม่หมดอายุ (≤8 ชม.)
type SessionStats struct {
	TotalSessions  int           `json:"totalsessions"`
	ActiveSessions int           `json:"activesessions"`
	ActiveWindowMs int64         `json:"activewindowms"`
	Holdings       []SessionStat `json:"holdings"`
	Entries        []SessionEntry `json:"entries"`
}

func hmStringValue(value interface{}) string {
	if s, ok := value.(string); ok {
		return s
	}
	return ""
}

func hmInt64Value(value interface{}) int64 {
	raw := hmStringValue(value)
	parsed, err := strconv.ParseInt(strings.TrimSpace(raw), 10, 64)
	if err != nil {
		return 0
	}
	return parsed
}

// ActiveSessionStats นับเซสชันใน Redis แยกตามกลุ่มกิจการ
// (ผู้ดูแลเรียกดูว่า "ขณะนี้มีใคร/กี่เซสชันกำลังใช้ระบบอยู่")
func (authService *AuthService) ActiveSessionStats() (*SessionStats, error) {
	keys, err := authService.cacher.Keys(authService.prefixSessionCacheKey + "*")
	if err != nil {
		return nil, err
	}
	now := authService.timeNow().UnixMilli()
	activeWindowMs := authService.sessionIdleTimeout.Milliseconds()
	stats := &SessionStats{ActiveWindowMs: activeWindowMs, Holdings: []SessionStat{}, Entries: []SessionEntry{}}
	byHolding := map[string]*SessionStat{}
	byUser := map[string]*SessionEntry{}

	for _, key := range keys {
		// Keys("session-*") จับ marker session-revoked-* มาด้วย — เซ็นห์ออก
		if strings.HasPrefix(key, authService.prefixRevokedSessionKey) {
			continue
		}
		values, err := authService.cacher.HMGet(key, []string{"holdingcode", "lastseenat", "username", "name", "role", "createdat", "accesskey"})
		if err != nil || len(values) < 7 {
			continue
		}
		holding := strings.TrimSpace(hmStringValue(values[0]))
		lastSeen := hmInt64Value(values[1])
		username := strings.TrimSpace(hmStringValue(values[2]))
		displayName := strings.TrimSpace(hmStringValue(values[3]))
		role := strings.TrimSpace(hmStringValue(values[4]))
		createdAt := hmInt64Value(values[5])
		accessKey := strings.TrimSpace(hmStringValue(values[6]))

		// เซสชันเก่าก่อนเพิ่มฟิลด์ username/name — ดึงจาก bearer cache ถ้ายังไม่หมดอายุ
		if username == "" && accessKey != "" {
			if bearer, err := authService.cacher.HMGet(accessKey, []string{"username", "name"}); err == nil && len(bearer) >= 2 {
				username = strings.TrimSpace(hmStringValue(bearer[0]))
				displayName = strings.TrimSpace(hmStringValue(bearer[1]))
			}
		}

		stats.TotalSessions++
		active := now-lastSeen <= activeWindowMs
		if active {
			stats.ActiveSessions++
		}
		stat, exists := byHolding[holding]
		if !exists {
			stat = &SessionStat{HoldingCode: holding}
			byHolding[holding] = stat
		}
		stat.Sessions++
		if active {
			stat.Active++
		}
		if lastSeen > stat.LastSeenAt {
			stat.LastSeenAt = lastSeen
		}

		// distinct ตามผู้ใช้: คนเดียวเปิดหลายแท็บ/หลายเครื่อง = 1 แถว
		// ไม่รู้ชื่อ (เซสชันเก่า) จับเป็นกลุ่มเดียวกันต่อ holding เพื่อไม่สับสน
		userKey := username + "|" + displayName + "|" + holding
		entry, exists := byUser[userKey]
		if !exists {
			entry = &SessionEntry{
				Username:    username,
				Name:        displayName,
				HoldingCode: holding,
				Role:        role,
			}
			byUser[userKey] = entry
		}
		entry.Sessions++
		if active {
			entry.Active = true
		}
		if createdAt > entry.CreatedAt {
			entry.CreatedAt = createdAt
		}
		if lastSeen > entry.LastSeenAt {
			entry.LastSeenAt = lastSeen
		}
	}

	stats.Holdings = make([]SessionStat, 0, len(byHolding))
	for _, stat := range byHolding {
		stats.Holdings = append(stats.Holdings, *stat)
	}
	sort.Slice(stats.Holdings, func(i, j int) bool {
		if stats.Holdings[i].Sessions != stats.Holdings[j].Sessions {
			return stats.Holdings[i].Sessions > stats.Holdings[j].Sessions
		}
		return stats.Holdings[i].HoldingCode < stats.Holdings[j].HoldingCode
	})
	stats.Entries = make([]SessionEntry, 0, len(byUser))
	for _, entry := range byUser {
		stats.Entries = append(stats.Entries, *entry)
	}
	// ล่าสุดที่ใช้ก่อน — รายการบนหน้าต่างเห็นใครกำลังออนไลน์ทันที
	sort.Slice(stats.Entries, func(i, j int) bool {
		return stats.Entries[i].LastSeenAt > stats.Entries[j].LastSeenAt
	})
	return stats, nil
}

func (authService *AuthService) issueSessionTokens(userInfo models.UserInfo, remaining time.Duration) (string, string, string, string, error) {
	if strings.TrimSpace(userInfo.SessionUID) == "" {
		return "", "", "", "", fmt.Errorf("session uid is required")
	}
	if remaining <= 0 {
		return "", "", "", "", fmt.Errorf("session expired")
	}

	accessToken := authService.encrypt.GenerateSHA256Hash(NewUUID())
	refreshToken := authService.encrypt.GenerateSHA256Hash(NewUUID())
	accessKey := authService.prefixBearerCacheKey + accessToken
	refreshKey := authService.prefixRefreshCacheKey + refreshToken
	fields := map[string]interface{}{
		"sessionuid":        userInfo.SessionUID,
		"username":          userInfo.Username,
		"name":              userInfo.Name,
		"uid":               userInfo.UID,
		"holdingcode":       userInfo.HoldingCode,
		"businesscode":      userInfo.BusinessCode,
		"role":              userInfo.Role,
		"membershipuid":     userInfo.MembershipUID,
		"holdinguid":        userInfo.HoldingUID,
		"companyuid":        userInfo.CompanyUID,
		"branchuid":         userInfo.BranchUID,
		"permissionversion": userInfo.PermissionVersion,
	}

	accessTTL := boundedDuration(authService.expireTimeBearer, remaining)
	refreshTTL := boundedDuration(authService.expireTimeRefresh, remaining)
	if err := authService.cacher.HMSet(accessKey, fields); err != nil {
		return "", "", "", "", err
	}
	if err := authService.cacher.Expire(accessKey, accessTTL); err != nil {
		_ = authService.cacher.Del(accessKey)
		return "", "", "", "", err
	}
	if err := authService.cacher.HMSet(refreshKey, fields); err != nil {
		_ = authService.cacher.Del(accessKey)
		return "", "", "", "", err
	}
	if err := authService.cacher.Expire(refreshKey, refreshTTL); err != nil {
		_ = authService.cacher.Del(accessKey, refreshKey)
		return "", "", "", "", err
	}

	return accessToken, refreshToken, accessKey, refreshKey, nil
}

func (authService *AuthService) SelectShop(tokenType TokenType, tokenStr string, holdingCode string, businessCode string, branchUID string, role uint8) error {
	cacheKey := authService.GetPrefixCacheKey(tokenType) + tokenStr
	workspace := models.UserInfo{
		HoldingCode:  strings.TrimSpace(holdingCode),
		BusinessCode: strings.ToUpper(strings.TrimSpace(businessCode)),
		BranchUID:    strings.TrimSpace(branchUID),
		Role:         role,
	}
	identity, err := authService.cacher.HMGet(cacheKey, []string{"username", "name", "uid", "sessionuid"})
	if err != nil || len(identity) < 4 || strings.TrimSpace(cacheString(identity[2])) == "" {
		return fmt.Errorf("session identity invalid")
	}
	workspace.Username = cacheString(identity[0])
	workspace.Name = cacheString(identity[1])
	workspace.UID = cacheString(identity[2])
	workspace.SessionUID = cacheString(identity[3])
	if authService.authorization != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		workspace, err = authService.authorization.SelectWorkspace(ctx, workspace, holdingCode, businessCode, branchUID)
		if err != nil {
			return err
		}
	}

	err = authService.cacher.HMSet(cacheKey, workspaceCacheFields(workspace))
	if err != nil {
		return err
	}
	if tokenType == AUTHTYPE_BEARER || tokenType == AUTHTYPE_WEBSOCKET {
		sessionUID := workspace.SessionUID
		if strings.TrimSpace(sessionUID) == "" {
			return fmt.Errorf("session invalid")
		}
		err = authService.cacher.HMSet(authService.prefixSessionCacheKey+sessionUID, workspaceCacheFields(workspace))
		if err != nil {
			return err
		}
	}

	return nil
}

func workspaceCacheFields(userInfo models.UserInfo) map[string]interface{} {
	return map[string]interface{}{
		"holdingcode":       userInfo.HoldingCode,
		"businesscode":      userInfo.BusinessCode,
		"role":              userInfo.Role,
		"membershipuid":     userInfo.MembershipUID,
		"holdinguid":        userInfo.HoldingUID,
		"companyuid":        userInfo.CompanyUID,
		"branchuid":         userInfo.BranchUID,
		"permissionversion": userInfo.PermissionVersion,
	}
}

func (authService *AuthService) RefreshToken(token string) (string, string, bool, error) {
	cacheKey := authService.GetPrefixCacheKey(AUTHTYPE_REFRESH) + token
	markerKey := authService.prefixUsedRefreshKey + token
	fields := []string{"sessionuid", "username", "name", "uid"}
	values, marker, consumed, err := authService.cacher.ConsumeHash(cacheKey, markerKey, fields, authService.sessionAbsoluteMax)
	if err != nil {
		return "", "", false, err
	}
	if !consumed {
		if strings.TrimSpace(marker) != "" {
			_ = authService.revokeSessionByID(marker)
		}
		return "", "", false, fmt.Errorf("refresh token invalid or reused")
	}
	if len(values) != len(fields) || strings.TrimSpace(cacheString(values[0])) == "" {
		return "", "", false, fmt.Errorf("refresh token invalid")
	}

	userInfo := models.UserInfo{
		SessionUID: cacheString(values[0]),
		Username:   cacheString(values[1]),
		Name:       cacheString(values[2]),
		UID:        cacheString(values[3]),
	}
	state, remaining, err := authService.activeSession(userInfo.SessionUID, authService.timeNow().UTC())
	if err != nil {
		_ = authService.revokeSessionByID(userInfo.SessionUID)
		return "", "", false, err
	}
	userInfo.HoldingCode = state["holdingcode"]
	userInfo.BusinessCode = state["businesscode"]
	userInfo.MembershipUID = state["membershipuid"]
	userInfo.HoldingUID = state["holdinguid"]
	userInfo.CompanyUID = state["companyuid"]
	userInfo.BranchUID = state["branchuid"]
	if role, parseErr := strconv.ParseUint(state["role"], 10, 8); parseErr == nil {
		userInfo.Role = uint8(role)
	}
	if permissionVersion, parseErr := strconv.ParseInt(state["permissionversion"], 10, 64); parseErr == nil {
		userInfo.PermissionVersion = permissionVersion
	}
	if authService.authorization != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		resolved, authErr := authService.authorization.Authorize(ctx, userInfo)
		cancel()
		if errors.Is(authErr, ErrLiveWorkspaceAccess) {
			if clearErr := authService.clearSessionWorkspace(userInfo.SessionUID); clearErr != nil {
				_ = authService.revokeSessionByID(userInfo.SessionUID)
				return "", "", false, clearErr
			}
			userInfo = loginOnlyUserInfo(userInfo)
		} else if authErr != nil {
			_ = authService.revokeSessionByID(userInfo.SessionUID)
			return "", "", false, authErr
		} else {
			userInfo = resolved
		}
	}

	accessToken, refreshToken, accessKey, refreshKey, err := authService.issueSessionTokens(userInfo, remaining)
	if err != nil {
		_ = authService.revokeSessionByID(userInfo.SessionUID)
		return "", "", false, err
	}
	sessionKey := authService.prefixSessionCacheKey + userInfo.SessionUID
	sessionFields := workspaceCacheFields(userInfo)
	sessionFields["accesskey"] = accessKey
	sessionFields["refreshkey"] = refreshKey
	sessionFields["lastseenat"] = authService.timeNow().UTC().UnixMilli()
	err = authService.cacher.HMSet(sessionKey, sessionFields)
	if err == nil {
		err = authService.cacher.Expire(sessionKey, remaining)
	}
	if err != nil {
		_ = authService.cacher.Del(accessKey, refreshKey)
		_ = authService.revokeSessionByID(userInfo.SessionUID)
		return "", "", false, err
	}
	return accessToken, refreshToken, false, nil
}

func (authService *AuthService) clearSessionWorkspace(sessionUID string) error {
	if strings.TrimSpace(sessionUID) == "" {
		return fmt.Errorf("session invalid")
	}
	return authService.cacher.HMSet(authService.prefixSessionCacheKey+sessionUID, workspaceCacheFields(models.UserInfo{}))
}

func (authService *AuthService) activeSession(sessionUID string, now time.Time) (map[string]string, time.Duration, error) {
	if strings.TrimSpace(sessionUID) == "" {
		return nil, 0, fmt.Errorf("session invalid")
	}
	revokedKey := authService.prefixRevokedSessionKey + sessionUID
	revoked, err := authService.cacher.Exists(revokedKey)
	if err != nil {
		return nil, 0, err
	}
	if revoked {
		return nil, 0, fmt.Errorf("session revoked")
	}
	state, err := authService.cacher.HGetAll(authService.prefixSessionCacheKey + sessionUID)
	if err != nil {
		return nil, 0, err
	}
	if len(state) == 0 || strings.EqualFold(state["revoked"], "true") {
		return nil, 0, fmt.Errorf("session invalid")
	}
	// Check again after reading state so a concurrent refresh-token replay cannot
	// revoke the family and have another request continue from stale state.
	revoked, err = authService.cacher.Exists(revokedKey)
	if err != nil {
		return nil, 0, err
	}
	if revoked {
		return nil, 0, fmt.Errorf("session revoked")
	}
	createdAt, err := parseCacheTime(state["createdat"])
	if err != nil {
		return nil, 0, fmt.Errorf("session invalid")
	}
	lastSeenAt, err := parseCacheTime(state["lastseenat"])
	if err != nil {
		return nil, 0, fmt.Errorf("session invalid")
	}
	absoluteExpiry := createdAt.Add(authService.sessionAbsoluteMax)
	if !now.Before(absoluteExpiry) || now.Sub(lastSeenAt) >= authService.sessionIdleTimeout {
		return nil, 0, fmt.Errorf("session expired")
	}
	return state, absoluteExpiry.Sub(now), nil
}

// AuthenticateAccessToken resolves identity from the access token and current
// workspace authority from the shared session plus MongoDB.
func (authService *AuthService) AuthenticateAccessToken(ctx context.Context, token string) (models.UserInfo, error) {
	cacheKey := authService.prefixBearerCacheKey + strings.TrimSpace(token)
	raw, err := authService.cacher.HMGet(cacheKey, []string{"username", "name", "uid", "sessionuid"})
	if err != nil || len(raw) < 4 || strings.TrimSpace(cacheString(raw[2])) == "" {
		return models.UserInfo{}, fmt.Errorf("token invalid")
	}
	identity := models.UserInfo{
		Username:   cacheString(raw[0]),
		Name:       cacheString(raw[1]),
		UID:        cacheString(raw[2]),
		SessionUID: cacheString(raw[3]),
	}
	return authService.authenticateCachedAccess(ctx, AUTHTYPE_BEARER, cacheKey, identity)
}

func (authService *AuthService) authenticateCachedAccess(ctx context.Context, tokenType TokenType, accessKey string, identity models.UserInfo) (models.UserInfo, error) {
	if tokenType == AUTHTYPE_XAPIKEY {
		return identity, nil
	}
	if strings.TrimSpace(identity.SessionUID) == "" && authService.allowLegacyToken {
		return identity, nil
	}
	now := authService.timeNow().UTC()
	state, remaining, err := authService.activeSession(identity.SessionUID, now)
	if err != nil {
		_ = authService.revokeSessionByID(identity.SessionUID)
		return models.UserInfo{}, err
	}
	selected := userInfoFromSession(identity, state)
	if authService.authorization != nil {
		resolved, authErr := authService.authorization.Authorize(ctx, selected)
		if errors.Is(authErr, ErrLiveWorkspaceAccess) {
			if clearErr := authService.clearSessionWorkspace(identity.SessionUID); clearErr != nil {
				return models.UserInfo{}, clearErr
			}
			if touchErr := authService.touchSession(identity.SessionUID, now, remaining); touchErr != nil {
				return models.UserInfo{}, touchErr
			}
			return loginOnlyUserInfo(identity), ErrLiveWorkspaceAccess
		}
		if authErr != nil {
			if errors.Is(authErr, ErrLiveUserAccess) {
				_ = authService.revokeSessionByID(identity.SessionUID)
			}
			return models.UserInfo{}, authErr
		}
		selected = resolved
	}
	if err := authService.touchSession(identity.SessionUID, now, remaining); err != nil {
		return models.UserInfo{}, err
	}
	return selected, nil
}

func userInfoFromSession(identity models.UserInfo, state map[string]string) models.UserInfo {
	identity.HoldingCode = strings.TrimSpace(state["holdingcode"])
	identity.BusinessCode = strings.ToUpper(strings.TrimSpace(state["businesscode"]))
	identity.MembershipUID = strings.TrimSpace(state["membershipuid"])
	identity.HoldingUID = strings.TrimSpace(state["holdinguid"])
	identity.CompanyUID = strings.TrimSpace(state["companyuid"])
	identity.BranchUID = strings.TrimSpace(state["branchuid"])
	if role, err := strconv.ParseUint(strings.TrimSpace(state["role"]), 10, 8); err == nil {
		identity.Role = uint8(role)
	}
	if version, err := strconv.ParseInt(strings.TrimSpace(state["permissionversion"]), 10, 64); err == nil {
		identity.PermissionVersion = version
	}
	return identity
}

func (authService *AuthService) touchSession(sessionUID string, now time.Time, remaining time.Duration) error {
	sessionKey := authService.prefixSessionCacheKey + sessionUID
	if err := authService.cacher.HMSet(sessionKey, map[string]interface{}{"lastseenat": now.UnixMilli()}); err != nil {
		return err
	}
	return authService.cacher.Expire(sessionKey, remaining)
}

func (authService *AuthService) validateAccessSession(tokenType TokenType, accessKey string, sessionUID string) error {
	if tokenType == AUTHTYPE_XAPIKEY {
		return nil
	}
	if strings.TrimSpace(sessionUID) == "" && authService.allowLegacyToken {
		return nil
	}
	now := authService.timeNow().UTC()
	_, remaining, err := authService.activeSession(sessionUID, now)
	if err != nil {
		_ = authService.revokeSessionByID(sessionUID)
		return err
	}
	return authService.touchSession(sessionUID, now, remaining)
}

func parseCacheTime(raw string) (time.Time, error) {
	millis, err := strconv.ParseInt(strings.TrimSpace(raw), 10, 64)
	if err != nil || millis <= 0 {
		return time.Time{}, fmt.Errorf("invalid cached time")
	}
	return time.UnixMilli(millis).UTC(), nil
}

func (authService *AuthService) RevokeSession(tokenAuthorizationHeader string) error {
	token, err := authService.GetTokenFromAuthorizationHeader(AUTHTYPE_BEARER, tokenAuthorizationHeader)
	if err != nil {
		return err
	}
	accessKey := authService.prefixBearerCacheKey + token
	sessionUID, err := authService.cacher.HGet(accessKey, "sessionuid")
	if err != nil {
		return err
	}
	if strings.TrimSpace(sessionUID) == "" {
		return authService.cacher.Del(accessKey)
	}
	return authService.revokeSessionByID(sessionUID)
}

func (authService *AuthService) revokeSessionByID(sessionUID string) error {
	sessionUID = strings.TrimSpace(sessionUID)
	if sessionUID == "" {
		return nil
	}
	sessionKey := authService.prefixSessionCacheKey + sessionUID
	// Keep a short-lived tombstone before removing the session. Without it, a
	// concurrent refresh that already read the old state could recreate the hash
	// after replay detection revoked the token family.
	if err := authService.cacher.SetS(authService.prefixRevokedSessionKey+sessionUID, "true", authService.sessionAbsoluteMax); err != nil {
		return err
	}
	state, err := authService.cacher.HGetAll(sessionKey)
	if err != nil {
		return err
	}
	if len(state) > 0 {
		if err := authService.cacher.HMSet(sessionKey, map[string]interface{}{"revoked": true}); err != nil {
			return err
		}
	}
	keys := []string{sessionKey}
	for _, field := range []string{"accesskey", "refreshkey"} {
		if key := strings.TrimSpace(state[field]); key != "" {
			keys = append(keys, key)
			authService.cacheMemory.Delete(key)
		}
	}
	return authService.cacher.Del(keys...)
}

func (authService *AuthService) RevokeUserTokens(username string) error {
	username = strings.TrimSpace(username)
	if username == "" {
		return fmt.Errorf("username invalid")
	}
	return authService.revokeUserTokensByField("username", username, true)
}

func (authService *AuthService) RevokeUserTokensByUID(userUID string) error {
	userUID = strings.TrimSpace(userUID)
	if userUID == "" {
		return fmt.Errorf("user identity invalid")
	}
	return authService.revokeUserTokensByField("uid", userUID, false)
}

func (authService *AuthService) revokeUserTokensByField(field string, value string, caseInsensitive bool) error {
	keysToDelete := make([]string, 0)
	sessionUIDs := make(map[string]struct{})
	for _, prefix := range []string{authService.prefixBearerCacheKey, authService.prefixRefreshCacheKey, authService.prefixXApiKeyCacheKey} {
		keys, err := authService.cacher.Keys(prefix + "*")
		if err != nil {
			return err
		}
		for _, key := range keys {
			cachedValue, err := authService.cacher.HGet(key, field)
			if err != nil {
				return err
			}
			matches := strings.TrimSpace(cachedValue) == value
			if caseInsensitive {
				matches = strings.EqualFold(strings.TrimSpace(cachedValue), value)
			}
			if !matches {
				continue
			}
			sessionUID, sessionErr := authService.cacher.HGet(key, "sessionuid")
			if sessionErr != nil {
				return sessionErr
			}
			if strings.TrimSpace(sessionUID) != "" {
				sessionUIDs[sessionUID] = struct{}{}
			} else {
				keysToDelete = append(keysToDelete, key)
			}
			authService.cacheMemory.Delete(key)
		}
	}
	for sessionUID := range sessionUIDs {
		if err := authService.revokeSessionByID(sessionUID); err != nil {
			return err
		}
	}
	if len(keysToDelete) > 0 {
		return authService.cacher.Del(keysToDelete...)
	}
	return nil
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
