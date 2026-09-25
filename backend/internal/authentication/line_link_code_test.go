package authentication

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"smlcloudplatform/pkg/microservice"
	msModels "smlcloudplatform/pkg/microservice/models"
	msValidator "smlcloudplatform/pkg/validator"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
)

// fakeLineLinkCodeStore embeds ICacher so it can back a Microservice; only Get/SetS/Del are implemented.
type fakeLineLinkCodeStore struct {
	microservice.ICacher
	values  map[string]string
	ttl     map[string]time.Duration
	getErr  error
	deleted []string
}

func newFakeLineLinkCodeStore() *fakeLineLinkCodeStore {
	return &fakeLineLinkCodeStore{values: map[string]string{}, ttl: map[string]time.Duration{}}
}

func (s *fakeLineLinkCodeStore) Get(key string) (string, error) {
	if s.getErr != nil {
		return "", s.getErr
	}
	return s.values[key], nil
}

func (s *fakeLineLinkCodeStore) SetS(key string, value string, expire time.Duration) error {
	s.values[key] = value
	s.ttl[key] = expire
	return nil
}

func (s *fakeLineLinkCodeStore) Del(keys ...string) error {
	for _, key := range keys {
		delete(s.values, key)
		s.deleted = append(s.deleted, key)
	}
	return nil
}

func TestLineLinkCodeBindsToMinter(t *testing.T) {
	store := newFakeLineLinkCodeStore()

	if err := bindLineLinkCode(store, " 123456 ", "uid-a"); err != nil {
		t.Fatalf("bind: %v", err)
	}
	if got := store.values["linelinkcode:123456"]; got != "uid-a" {
		t.Fatalf("stored owner = %q, want uid-a", got)
	}
	if got := store.ttl["linelinkcode:123456"]; got != lineLinkCodeTTL || got <= 5*time.Minute {
		t.Fatalf("ttl = %v, want %v (longer than the 5-minute polling window)", got, lineLinkCodeTTL)
	}

	if err := checkLineLinkCode(store, "123456", "uid-a"); err != nil {
		t.Fatalf("minter check: %v", err)
	}
	if err := checkLineLinkCode(store, "123456", "uid-b"); !errors.Is(err, errLineLinkCodeNotOwned) {
		t.Fatalf("other user check err = %v, want not owned", err)
	}
}

func TestLineLinkCodeFirstBindingWins(t *testing.T) {
	store := newFakeLineLinkCodeStore()
	if err := bindLineLinkCode(store, "123456", "uid-a"); err != nil {
		t.Fatalf("bind: %v", err)
	}

	if err := bindLineLinkCode(store, "123456", "uid-b"); !errors.Is(err, errLineLinkCodeTaken) {
		t.Fatalf("rebind by another user err = %v, want taken", err)
	}
	if got := store.values["linelinkcode:123456"]; got != "uid-a" {
		t.Fatalf("owner changed to %q", got)
	}
	if err := bindLineLinkCode(store, "123456", "uid-a"); err != nil {
		t.Fatalf("same user retry: %v", err)
	}
}

func TestLineLinkCodeRejectsMissingExpiredOrInvalid(t *testing.T) {
	store := newFakeLineLinkCodeStore()

	for name, code := range map[string]string{
		"never bound or expired": "999999",
		"empty":                  "   ",
		"control character":      "12\n34",
		"too long":               string(make([]byte, lineLinkCodeMaxLength+1)),
	} {
		if err := checkLineLinkCode(store, code, "uid-a"); !errors.Is(err, errLineLinkCodeNotOwned) {
			t.Errorf("%s: err = %v, want not owned", name, err)
		}
	}
	if err := bindLineLinkCode(store, "12 34", "uid-a"); !errors.Is(err, errLineLinkCodeInvalid) {
		t.Errorf("bind with whitespace err = %v, want invalid", err)
	}
	if err := bindLineLinkCode(store, "123456", " "); err == nil {
		t.Error("bind without a session uid must fail")
	}
	if err := checkLineLinkCode(store, "123456", ""); !errors.Is(err, errLineLinkCodeNotOwned) {
		t.Errorf("check without a session uid err = %v, want not owned", err)
	}
	if len(store.values) != 0 {
		t.Fatalf("rejected calls must not store anything: %v", store.values)
	}
}

func TestLineLinkCodeStoreErrorIsNotTreatedAsOwnership(t *testing.T) {
	store := newFakeLineLinkCodeStore()
	store.getErr = errors.New("database down")

	if err := checkLineLinkCode(store, "123456", "uid-a"); err == nil || errors.Is(err, errLineLinkCodeNotOwned) {
		t.Fatalf("check err = %v, want the store error", err)
	}
	if err := bindLineLinkCode(store, "123456", "uid-a"); err == nil {
		t.Fatal("bind must fail when the store cannot be read")
	}
}

func TestLineLinkCodeReleasedAfterLink(t *testing.T) {
	store := newFakeLineLinkCodeStore()
	if err := bindLineLinkCode(store, "123456", "uid-a"); err != nil {
		t.Fatalf("bind: %v", err)
	}

	releaseLineLinkCode(store, "123456")

	if err := checkLineLinkCode(store, "123456", "uid-a"); !errors.Is(err, errLineLinkCodeNotOwned) {
		t.Fatalf("released code check err = %v, want not owned", err)
	}
}

// Real cache_entries (PostgreSQL) — skipped unless BC_GL_TEST_POSTGRES_DSN is set.
func TestLineLinkCodeOnPostgresCache(t *testing.T) {
	dsn := os.Getenv("BC_GL_TEST_POSTGRES_DSN")
	if dsn == "" {
		t.Skip("BC_GL_TEST_POSTGRES_DSN not set")
	}
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatalf("open postgres: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	cacher, err := microservice.NewCacher(db)
	if err != nil {
		t.Fatalf("new cacher: %v", err)
	}
	code := "it-" + strconv.FormatInt(time.Now().UnixNano(), 10)
	key := lineLinkCodeKeyPrefix + code
	t.Cleanup(func() { _, _ = db.Exec(`DELETE FROM cache_entries WHERE cache_key = $1`, key) })

	if err := bindLineLinkCode(cacher, code, "uid-a"); err != nil {
		t.Fatalf("bind: %v", err)
	}
	var owner string
	var liveLongEnough bool
	if err := db.QueryRow(`SELECT value, expires_at > now() + interval '9 minutes' FROM cache_entries WHERE cache_key = $1`, key).Scan(&owner, &liveLongEnough); err != nil {
		t.Fatalf("read binding row: %v", err)
	}
	if owner != "uid-a" || !liveLongEnough {
		t.Fatalf("row = (%q, ttl>9m %v), want (uid-a, true)", owner, liveLongEnough)
	}
	if err := checkLineLinkCode(cacher, code, "uid-a"); err != nil {
		t.Fatalf("minter check: %v", err)
	}
	if err := checkLineLinkCode(cacher, code, "uid-b"); !errors.Is(err, errLineLinkCodeNotOwned) {
		t.Fatalf("other user check err = %v", err)
	}
	if err := bindLineLinkCode(cacher, code, "uid-b"); !errors.Is(err, errLineLinkCodeTaken) {
		t.Fatalf("other user bind err = %v", err)
	}

	if _, err := db.Exec(`UPDATE cache_entries SET expires_at = now() - interval '1 second' WHERE cache_key = $1`, key); err != nil {
		t.Fatalf("expire binding: %v", err)
	}
	if err := checkLineLinkCode(cacher, code, "uid-a"); !errors.Is(err, errLineLinkCodeNotOwned) {
		t.Fatalf("expired code check err = %v, want not owned", err)
	}

	if err := bindLineLinkCode(cacher, code, "uid-a"); err != nil {
		t.Fatalf("rebind after expiry: %v", err)
	}
	releaseLineLinkCode(cacher, code)
	var rows int
	if err := db.QueryRow(`SELECT count(*) FROM cache_entries WHERE cache_key = $1`, key).Scan(&rows); err != nil {
		t.Fatalf("count rows: %v", err)
	}
	if rows != 0 {
		t.Fatalf("released code left %d rows", rows)
	}
}

type lineLinkResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// serveLineLink runs one handler through the real HTTP context as the middleware would leave it.
func serveLineLink(t *testing.T, store *fakeLineLinkCodeStore, handler func(AuthenticationHttp, microservice.IContext) error, uid string, lang string, body string) (int, lineLinkResponse) {
	t.Helper()
	ms := &microservice.Microservice{}
	ms.SetCacher(store)
	e := echo.New()
	e.Validator = msValidator.NewCustomValidator()
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set("Accept-Language", lang)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("UserInfo", msModels.UserInfo{UID: uid, Username: "user-" + uid})
	_ = handler(AuthenticationHttp{ms: ms}, microservice.NewHTTPContext(ms, c))
	var out lineLinkResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
		t.Fatalf("decode response %q: %v", rec.Body.String(), err)
	}
	return rec.Code, out
}

func TestLineLinkCodeHandlers(t *testing.T) {
	store := newFakeLineLinkCodeStore()
	bind := AuthenticationHttp.BindLineLinkCode
	check := AuthenticationHttp.CheckLineLinkCode

	if status, out := serveLineLink(t, store, bind, "uid-a", "th", `{"code":"123456"}`); status != http.StatusOK || !out.Success {
		t.Fatalf("A bind = %d %+v", status, out)
	}
	if status, out := serveLineLink(t, store, bind, "uid-b", "en", `{"code":"123456"}`); status != http.StatusConflict ||
		out.Message != "This LINE link code is already bound to another account. Please start linking LINE again." {
		t.Fatalf("B bind = %d %+v", status, out)
	}
	if status, out := serveLineLink(t, store, check, "uid-b", "th", `{"code":"123456"}`); status != http.StatusNotFound ||
		!strings.HasPrefix(out.Message, "รหัสเชื่อมต่อ LINE นี้ไม่ใช่ของบัญชีที่เข้าระบบอยู่") {
		t.Fatalf("B check = %d %+v", status, out)
	}
	if status, out := serveLineLink(t, store, check, "uid-a", "th", `{"code":"123456"}`); status != http.StatusOK || !out.Success {
		t.Fatalf("A check = %d %+v", status, out)
	}
	if status, _ := serveLineLink(t, store, bind, "uid-a", "th", `{"code":`); status != http.StatusBadRequest {
		t.Fatalf("broken JSON bind = %d, want 400", status)
	}
	if store.values["linelinkcode:123456"] != "uid-a" {
		t.Fatalf("owner changed: %v", store.values)
	}
}

// LinkLine must refuse another user's code before the service links anything (the service is nil here,
// so reaching it would panic).
func TestLinkLineRequiresOwnCode(t *testing.T) {
	store := newFakeLineLinkCodeStore()
	if err := bindLineLinkCode(store, "123456", "uid-a"); err != nil {
		t.Fatalf("bind: %v", err)
	}

	for name, body := range map[string]string{
		"another user's code": `{"code":"123456","lineuserid":"U123"}`,
		"no code":             `{"lineuserid":"U123"}`,
	} {
		status, out := serveLineLink(t, store, AuthenticationHttp.LinkLine, "uid-b", "th", body)
		if status != http.StatusNotFound || out.Success {
			t.Errorf("%s: LinkLine = %d %+v, want 404", name, status, out)
		}
	}
	if store.values["linelinkcode:123456"] != "uid-a" {
		t.Fatalf("binding must survive a refused link: %v", store.values)
	}
}
