package microservice

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"smlcloudplatform/internal/config"
	"smlcloudplatform/pkg/microservice/models"

	"github.com/labstack/echo/v4"
)

type sessionTestCacherConfig struct {
	endpoint string
}

func (cfg *sessionTestCacherConfig) Endpoint() string { return cfg.endpoint }
func (cfg *sessionTestCacherConfig) Password() string { return "" }
func (cfg *sessionTestCacherConfig) DB() int          { return 0 }
func (cfg *sessionTestCacherConfig) UserName() string { return "" }
func (cfg *sessionTestCacherConfig) TLS() bool        { return false }
func (cfg *sessionTestCacherConfig) ConnectionSettings() config.ICacherConnectionSettings {
	return config.NewDefaultCacherConnectionSettings()
}

func TestGetTokenFromContextRejectsXAPIKey(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("x-api-key", "legacy-user-key")
	ctx := e.NewContext(req, httptest.NewRecorder())
	authService := &AuthService{prefixBearerToken: "Bearer"}

	if _, err := authService.GetTokenFromContext(ctx); err == nil {
		t.Fatal("user middleware must reject x-api-key authentication")
	}
}

func TestSessionRefreshRotationRejectsReplay(t *testing.T) {
	endpoint := os.Getenv("TEST_REDIS_ADDR")
	if endpoint == "" {
		t.Skip("TEST_REDIS_ADDR not set")
	}

	cacher := NewCacher(&sessionTestCacherConfig{endpoint: endpoint})
	if err := cacher.Healthcheck(); err != nil {
		t.Fatalf("redis healthcheck: %v", err)
	}
	defer cacher.Close()

	prefix := "test-auth-" + NewUUID() + "-"
	refreshPrefix := "test-refresh-" + NewUUID() + "-"
	authService := NewAuthServicePrefix(prefix, refreshPrefix, cacher, time.Hour, 24*time.Hour)

	accessToken, refreshToken, err := authService.CreateSession(models.UserInfo{
		Username: "session_test_user",
		UID:      "session-test-user-uid",
	})
	if err != nil {
		t.Fatalf("create session: %v", err)
	}

	accessKey := prefix + accessToken
	sessionUID, err := cacher.HGet(accessKey, "sessionuid")
	if err != nil || sessionUID == "" {
		t.Fatalf("read session uid: %v", err)
	}
	t.Cleanup(func() {
		_ = cacher.Del(
			accessKey,
			refreshPrefix+refreshToken,
			"session-"+sessionUID,
			"session-revoked-"+sessionUID,
			"refresh-used-"+refreshToken,
		)
	})

	if err := authService.validateAccessSession(AUTHTYPE_BEARER, accessKey, sessionUID); err != nil {
		t.Fatalf("new access token must be active: %v", err)
	}
	redisClient, err := cacher.getClient()
	if err != nil {
		t.Fatalf("get redis client: %v", err)
	}
	accessTTL, err := redisClient.TTL(context.Background(), accessKey).Result()
	if err != nil || accessTTL <= 0 || accessTTL > accessTokenMaxAge {
		t.Fatalf("access token TTL = %v, want (0,%v]; err=%v", accessTTL, accessTokenMaxAge, err)
	}
	refreshTTL, err := redisClient.TTL(context.Background(), refreshPrefix+refreshToken).Result()
	if err != nil || refreshTTL <= 0 || refreshTTL > sessionMaxAge {
		t.Fatalf("refresh token TTL = %v, want (0,%v]; err=%v", refreshTTL, sessionMaxAge, err)
	}

	rotatedAccess, rotatedRefresh, _, err := authService.RefreshToken(refreshToken)
	if err != nil {
		t.Fatalf("rotate refresh token: %v", err)
	}
	rotatedAccessKey := prefix + rotatedAccess
	rotatedRefreshKey := refreshPrefix + rotatedRefresh
	t.Cleanup(func() {
		_ = cacher.Del(rotatedAccessKey, rotatedRefreshKey, "refresh-used-"+rotatedRefresh)
	})

	if exists, err := cacher.Exists(accessKey); err != nil || !exists {
		t.Fatalf("old access token must remain valid until its short TTL expires; exists=%v err=%v", exists, err)
	}
	if exists, err := cacher.Exists(refreshPrefix + refreshToken); err != nil || exists {
		t.Fatalf("old refresh token must be consumed; exists=%v err=%v", exists, err)
	}
	if err := authService.validateAccessSession(AUTHTYPE_BEARER, rotatedAccessKey, sessionUID); err != nil {
		t.Fatalf("rotated access token must be active: %v", err)
	}
	if err := authService.validateAccessSession(AUTHTYPE_BEARER, accessKey, sessionUID); err != nil {
		t.Fatalf("previous access token must stay active during multi-tab rotation: %v", err)
	}

	if _, _, _, err := authService.RefreshToken(refreshToken); err == nil {
		t.Fatal("replayed refresh token must be rejected")
	}
	if exists, err := cacher.Exists(rotatedAccessKey); err != nil || exists {
		t.Fatalf("refresh replay must revoke the token family; exists=%v err=%v", exists, err)
	}
	if exists, err := cacher.Exists(rotatedRefreshKey); err != nil || exists {
		t.Fatalf("refresh replay must revoke the rotated refresh token; exists=%v err=%v", exists, err)
	}
	if err := authService.validateAccessSession(AUTHTYPE_BEARER, accessKey, sessionUID); err == nil {
		t.Fatal("refresh replay must invalidate every access token through the revoked session")
	}
}

func TestLogoutWithPreviousAccessRevokesRotatedSession(t *testing.T) {
	endpoint := os.Getenv("TEST_REDIS_ADDR")
	if endpoint == "" {
		t.Skip("TEST_REDIS_ADDR not set")
	}

	cacher := NewCacher(&sessionTestCacherConfig{endpoint: endpoint})
	if err := cacher.Healthcheck(); err != nil {
		t.Fatalf("redis healthcheck: %v", err)
	}
	defer cacher.Close()

	prefix := "test-auth-" + NewUUID() + "-"
	refreshPrefix := "test-refresh-" + NewUUID() + "-"
	authService := NewAuthServicePrefix(prefix, refreshPrefix, cacher, time.Hour, 24*time.Hour)
	accessToken, refreshToken, err := authService.CreateSession(models.UserInfo{
		Username: "multi_tab_logout_user",
		UID:      "multi-tab-logout-user-uid",
	})
	if err != nil {
		t.Fatalf("create session: %v", err)
	}
	accessKey := prefix + accessToken
	sessionUID, err := cacher.HGet(accessKey, "sessionuid")
	if err != nil || sessionUID == "" {
		t.Fatalf("read session uid: %v", err)
	}

	rotatedAccess, rotatedRefresh, _, err := authService.RefreshToken(refreshToken)
	if err != nil {
		t.Fatalf("rotate refresh token: %v", err)
	}
	rotatedAccessKey := prefix + rotatedAccess
	rotatedRefreshKey := refreshPrefix + rotatedRefresh
	t.Cleanup(func() {
		_ = cacher.Del(
			accessKey,
			rotatedAccessKey,
			rotatedRefreshKey,
			"session-"+sessionUID,
			"session-revoked-"+sessionUID,
			"refresh-used-"+refreshToken,
			"refresh-used-"+rotatedRefresh,
		)
	})

	if err := authService.RevokeSession("Bearer " + accessToken); err != nil {
		t.Fatalf("logout with previous access token: %v", err)
	}
	if err := authService.validateAccessSession(AUTHTYPE_BEARER, rotatedAccessKey, sessionUID); err == nil {
		t.Fatal("logout must invalidate a rotated access token")
	}
	if exists, err := cacher.Exists(rotatedRefreshKey); err != nil || exists {
		t.Fatalf("logout must revoke the rotated refresh token; exists=%v err=%v", exists, err)
	}
}

func TestSessionIdleTimeoutRevokesCurrentTokens(t *testing.T) {
	endpoint := os.Getenv("TEST_REDIS_ADDR")
	if endpoint == "" {
		t.Skip("TEST_REDIS_ADDR not set")
	}

	cacher := NewCacher(&sessionTestCacherConfig{endpoint: endpoint})
	if err := cacher.Healthcheck(); err != nil {
		t.Fatalf("redis healthcheck: %v", err)
	}
	defer cacher.Close()

	prefix := "test-auth-" + NewUUID() + "-"
	refreshPrefix := "test-refresh-" + NewUUID() + "-"
	authService := NewAuthServicePrefix(prefix, refreshPrefix, cacher, time.Hour, 24*time.Hour)

	accessToken, refreshToken, err := authService.CreateSession(models.UserInfo{
		Username: "idle_session_test_user",
		UID:      "idle-session-test-user-uid",
	})
	if err != nil {
		t.Fatalf("create session: %v", err)
	}
	accessKey := prefix + accessToken
	sessionUID, err := cacher.HGet(accessKey, "sessionuid")
	if err != nil || sessionUID == "" {
		t.Fatalf("read session uid: %v", err)
	}
	sessionKey := "session-" + sessionUID
	t.Cleanup(func() {
		_ = cacher.Del(accessKey, refreshPrefix+refreshToken, sessionKey, "session-revoked-"+sessionUID, "refresh-used-"+refreshToken)
	})

	now := time.Now().UTC()
	authService.timeNow = func() time.Time { return now }
	if err := cacher.HMSet(sessionKey, map[string]interface{}{
		"createdat":  now.Add(-time.Hour).UnixMilli(),
		"lastseenat": now.Add(-sessionIdleMaxAge).UnixMilli(),
	}); err != nil {
		t.Fatalf("age session: %v", err)
	}

	if err := authService.validateAccessSession(AUTHTYPE_BEARER, accessKey, sessionUID); err == nil {
		t.Fatal("session at idle timeout boundary must be rejected")
	}
	if exists, err := cacher.Exists(accessKey); err != nil || exists {
		t.Fatalf("expired session access token must be revoked; exists=%v err=%v", exists, err)
	}
	if exists, err := cacher.Exists(refreshPrefix + refreshToken); err != nil || exists {
		t.Fatalf("expired session refresh token must be revoked; exists=%v err=%v", exists, err)
	}
}

func TestSessionAbsoluteTimeoutRevokesCurrentTokens(t *testing.T) {
	endpoint := os.Getenv("TEST_REDIS_ADDR")
	if endpoint == "" {
		t.Skip("TEST_REDIS_ADDR not set")
	}

	cacher := NewCacher(&sessionTestCacherConfig{endpoint: endpoint})
	if err := cacher.Healthcheck(); err != nil {
		t.Fatalf("redis healthcheck: %v", err)
	}
	defer cacher.Close()

	prefix := "test-auth-" + NewUUID() + "-"
	refreshPrefix := "test-refresh-" + NewUUID() + "-"
	authService := NewAuthServicePrefix(prefix, refreshPrefix, cacher, time.Hour, 24*time.Hour)

	accessToken, refreshToken, err := authService.CreateSession(models.UserInfo{
		Username: "absolute_session_test_user",
		UID:      "absolute-session-test-user-uid",
	})
	if err != nil {
		t.Fatalf("create session: %v", err)
	}
	accessKey := prefix + accessToken
	sessionUID, err := cacher.HGet(accessKey, "sessionuid")
	if err != nil || sessionUID == "" {
		t.Fatalf("read session uid: %v", err)
	}
	sessionKey := "session-" + sessionUID
	t.Cleanup(func() {
		_ = cacher.Del(accessKey, refreshPrefix+refreshToken, sessionKey, "session-revoked-"+sessionUID, "refresh-used-"+refreshToken)
	})

	now := time.Now().UTC()
	authService.timeNow = func() time.Time { return now }
	if err := cacher.HMSet(sessionKey, map[string]interface{}{
		"createdat":  now.Add(-sessionMaxAge).UnixMilli(),
		"lastseenat": now.UnixMilli(),
	}); err != nil {
		t.Fatalf("age session: %v", err)
	}

	if err := authService.validateAccessSession(AUTHTYPE_BEARER, accessKey, sessionUID); err == nil {
		t.Fatal("session at absolute timeout boundary must be rejected")
	}
	if exists, err := cacher.Exists(accessKey); err != nil || exists {
		t.Fatalf("absolute-expired access token must be revoked; exists=%v err=%v", exists, err)
	}
	if exists, err := cacher.Exists(refreshPrefix + refreshToken); err != nil || exists {
		t.Fatalf("absolute-expired refresh token must be revoked; exists=%v err=%v", exists, err)
	}
}

func TestRevokedSessionCannotBeResurrected(t *testing.T) {
	endpoint := os.Getenv("TEST_REDIS_ADDR")
	if endpoint == "" {
		t.Skip("TEST_REDIS_ADDR not set")
	}

	cacher := NewCacher(&sessionTestCacherConfig{endpoint: endpoint})
	if err := cacher.Healthcheck(); err != nil {
		t.Fatalf("redis healthcheck: %v", err)
	}
	defer cacher.Close()

	prefix := "test-auth-" + NewUUID() + "-"
	refreshPrefix := "test-refresh-" + NewUUID() + "-"
	authService := NewAuthServicePrefix(prefix, refreshPrefix, cacher, time.Hour, 24*time.Hour)
	accessToken, refreshToken, err := authService.CreateSession(models.UserInfo{
		Username: "revoked_session_test_user",
		UID:      "revoked-session-test-user-uid",
	})
	if err != nil {
		t.Fatalf("create session: %v", err)
	}
	accessKey := prefix + accessToken
	sessionUID, err := cacher.HGet(accessKey, "sessionuid")
	if err != nil || sessionUID == "" {
		t.Fatalf("read session uid: %v", err)
	}
	sessionKey := "session-" + sessionUID
	revokedKey := "session-revoked-" + sessionUID
	t.Cleanup(func() {
		_ = cacher.Del(accessKey, refreshPrefix+refreshToken, sessionKey, revokedKey, "refresh-used-"+refreshToken)
	})

	if err := authService.revokeSessionByID(sessionUID); err != nil {
		t.Fatalf("revoke session: %v", err)
	}
	// Model a refresh request that read the old state immediately before the
	// replay revocation and then attempted to recreate the session hash.
	now := time.Now().UTC()
	if err := cacher.HMSet(sessionKey, map[string]interface{}{
		"sessionuid": sessionUID,
		"username":   "revoked_session_test_user",
		"createdat":  now.UnixMilli(),
		"lastseenat": now.UnixMilli(),
		"accesskey":  accessKey,
		"refreshkey": refreshPrefix + refreshToken,
	}); err != nil {
		t.Fatalf("simulate stale session write: %v", err)
	}
	if err := authService.validateAccessSession(AUTHTYPE_BEARER, accessKey, sessionUID); err == nil {
		t.Fatal("revocation tombstone must reject a recreated session hash")
	}
}

func TestPermissionChangeClearsWorkspaceButKeepsLoginSession(t *testing.T) {
	endpoint := os.Getenv("TEST_REDIS_ADDR")
	if endpoint == "" {
		t.Skip("TEST_REDIS_ADDR not set")
	}

	cacher := NewCacher(&sessionTestCacherConfig{endpoint: endpoint})
	if err := cacher.Healthcheck(); err != nil {
		t.Fatalf("redis healthcheck: %v", err)
	}
	defer cacher.Close()

	_, finder, _ := activeAuthorizationFixture()
	prefix := "test-auth-" + NewUUID() + "-"
	refreshPrefix := "test-refresh-" + NewUUID() + "-"
	authService := NewAuthServicePrefix(prefix, refreshPrefix, cacher, time.Hour, 24*time.Hour, finder)
	accessToken, refreshToken, err := authService.CreateSession(models.UserInfo{
		Username: "permission_change_user",
		UID:      "user-1",
	})
	if err != nil {
		t.Fatalf("create session: %v", err)
	}
	accessKey := prefix + accessToken
	sessionUID, err := cacher.HGet(accessKey, "sessionuid")
	if err != nil || sessionUID == "" {
		t.Fatalf("read session uid: %v", err)
	}
	t.Cleanup(func() {
		_ = cacher.Del(
			accessKey,
			refreshPrefix+refreshToken,
			"session-"+sessionUID,
			"session-revoked-"+sessionUID,
			"refresh-used-"+refreshToken,
		)
	})

	if err := authService.SelectShop(AUTHTYPE_BEARER, accessToken, "HOLDING-A", "COMP-A", "", 1); err != nil {
		t.Fatalf("select workspace: %v", err)
	}
	selected, err := authService.AuthenticateAccessToken(context.Background(), accessToken)
	if err != nil || selected.CompanyUID != "company-1" {
		t.Fatalf("selected workspace = %#v, err=%v", selected, err)
	}

	finder.membership.PermissionVersion++
	if _, err := authService.AuthenticateAccessToken(context.Background(), accessToken); !errors.Is(err, ErrLiveWorkspaceAccess) {
		t.Fatalf("permission change error = %v, want ErrLiveWorkspaceAccess", err)
	}
	state, err := cacher.HGetAll("session-" + sessionUID)
	if err != nil {
		t.Fatalf("read cleared session: %v", err)
	}
	if state["holdingcode"] != "" || state["businesscode"] != "" || state["membershipuid"] != "" {
		t.Fatalf("workspace was not cleared: %#v", state)
	}

	loginOnly, err := authService.AuthenticateAccessToken(context.Background(), accessToken)
	if err != nil {
		t.Fatalf("login-only session must remain active: %v", err)
	}
	if loginOnly.UID != "user-1" || loginOnly.HoldingCode != "" {
		t.Fatalf("login-only identity = %#v", loginOnly)
	}
}
