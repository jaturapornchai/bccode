//go:build integration

package mcptoken

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"smlcloudplatform/pkg/microservice"
	"smlcloudplatform/pkg/microservice/models"
)

type managementRequest struct {
	microservice.IContext
	user     models.UserInfo
	request  *http.Request
	recorder *httptest.ResponseRecorder
	id       string
}

func (r *managementRequest) UserInfo() models.UserInfo           { return r.user }
func (r *managementRequest) Request() *http.Request              { return r.request }
func (r *managementRequest) ResponseWriter() http.ResponseWriter { return r.recorder }
func (r *managementRequest) Param(string) string                 { return r.id }
func (r *managementRequest) Response(status int, data interface{}) {
	r.recorder.WriteHeader(status)
	_ = json.NewEncoder(r.recorder).Encode(data)
}

func TestManagementHTTPPostgres(t *testing.T) {
	dsn := os.Getenv("MCP_TOKEN_TEST_DSN")
	if dsn == "" {
		t.Skip("MCP_TOKEN_TEST_DSN required")
	}
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	db.SetMaxOpenConns(1)
	namespace := fmt.Sprintf("mcp_http_test_%d", time.Now().UnixNano())
	if _, err = db.Exec("CREATE SCHEMA " + namespace); err != nil {
		t.Fatal(err)
	}
	defer db.Exec("DROP SCHEMA " + namespace + " CASCADE")
	if _, err = db.Exec("SET search_path TO " + namespace); err != nil {
		t.Fatal(err)
	}
	_, err = db.Exec(`CREATE TABLE users(id text,is_active boolean,username text DEFAULT '');
 CREATE TABLE holdings(code text,is_active boolean,profile jsonb DEFAULT '{}');
 CREATE TABLE holding_members(holding_code text,user_id text,role text,is_active boolean,access_scopes jsonb DEFAULT '[]',permission_sets jsonb DEFAULT '[]',id text DEFAULT md5(random()::text),access_expiry_date date);
 CREATE TABLE companies(holding_code text,code text,is_active boolean,name text DEFAULT 'Company');
 CREATE TABLE branches(holding_code text,company_code text,code text,is_active boolean);
 INSERT INTO users VALUES('admin',true),('staff',true),('other-admin',true);
 INSERT INTO holdings VALUES('H',true),('OTHER',true);
 INSERT INTO holding_members VALUES('H','admin','ADMIN',true),('H','staff','STAFF',true),('OTHER','other-admin','OWNER',true);
 INSERT INTO companies(holding_code,code,is_active) VALUES('H','C',true),('H','C2',true),('H','C3',true),('OTHER','FOREIGN',true);
 INSERT INTO branches VALUES('H','C','B',true),('H','C2','B2',true),('H','C','B3',true);`)
	if err != nil {
		t.Fatal(err)
	}
	h := &Http{connect: func(database string) (*sql.DB, error) {
		if database != ControlDatabase {
			t.Fatal("token management must use the central database")
		}
		return db, nil
	}}
	adminUser := models.UserInfo{UID: "admin", Username: "admin", HoldingCode: "H", BusinessCode: "", BranchUID: ""}
	input := fmt.Sprintf(`{"name":"test","kind":"mcp","companyCodes":["C","C2"],"mode":"readonly","expiresAt":%q}`, time.Now().UTC().Add(time.Hour).Format(time.RFC3339))
	call := func(handler func(microservice.IContext) error, u models.UserInfo, body, id string) *httptest.ResponseRecorder {
		r := &managementRequest{user: u, request: httptest.NewRequest("POST", "/mcp-tokens", strings.NewReader(body)), recorder: httptest.NewRecorder(), id: id}
		if e := handler(r); e != nil {
			t.Fatal(e)
		}
		if r.recorder.Header().Get("Cache-Control") != "no-store" {
			t.Fatal("missing no-store")
		}
		return r.recorder
	}
	staff := adminUser
	staff.UID = "staff"
	staff.Username = "staff"
	staff.Role = 1
	for _, handler := range []func(microservice.IContext) error{h.list, h.create, h.revoke} {
		if r := call(handler, staff, input, strings.Repeat("a", 32)); r.Code != 403 {
			t.Fatalf("nonadmin accepted: %d", r.Code)
		}
	}
	spoofed := staff
	spoofed.Username = adminUser.Username
	for _, handler := range []func(microservice.IContext) error{h.list, h.companies, h.create, h.revoke} {
		if r := call(handler, spoofed, input, strings.Repeat("a", 32)); r.Code != 403 {
			t.Fatal("another user's admin username grants authority")
		}
	}
	if r := call(h.create, adminUser, strings.Replace(input, `"kind":"mcp",`, "", 1), ""); r.Code != 400 {
		t.Fatal("missing audience accepted")
	}
	r := call(h.create, adminUser, input, "")
	if r.Code != 200 {
		t.Fatalf("create: %s", r.Body.String())
	}
	var created struct {
		Data struct {
			Token string `json:"token"`
			ID    string `json:"id"`
		}
	}
	if json.Unmarshal(r.Body.Bytes(), &created) != nil || created.Data.Token == "" || created.Data.ID == "" {
		t.Fatal("missing one-time credential")
	}
	r = call(h.list, adminUser, "", "")
	if r.Code != 200 {
		t.Fatal("list failed")
	}
	var listed struct{ Data []map[string]interface{} }
	if json.Unmarshal(r.Body.Bytes(), &listed) != nil || len(listed.Data) != 1 {
		t.Fatal("missing metadata")
	}
	if strings.Contains(r.Body.String(), created.Data.Token) || strings.Contains(r.Body.String(), "hash") || listed.Data[0]["token"] != nil {
		t.Fatal("list exposes credentials")
	}
	// Management belongs to the holding, independent of selected company/branch.
	selected := adminUser
	selected.BusinessCode = "C3"
	selected.BranchUID = "B3"
	r = call(h.list, selected, "", "")
	if json.Unmarshal(r.Body.Bytes(), &listed) != nil || len(listed.Data) != 1 {
		t.Fatal("holding admin cannot list across companies")
	}
	for _, codes := range []string{`[]`, `["FOREIGN"]`, `["missing"]`, `["C","C"]`} {
		invalid := strings.Replace(input, `["C","C2"]`, codes, 1)
		if r = call(h.create, adminUser, invalid, ""); r.Code != 400 {
			t.Fatalf("invalid company list accepted: %s", codes)
		}
	}
	foreign := adminUser
	foreign.HoldingCode = "OTHER"
	if r = call(h.revoke, foreign, "", created.Data.ID); r.Code != 403 {
		t.Fatal("foreign holding revoke accepted")
	}
	// An authorized administrator of another Holding still cannot see or revoke this token.
	foreign.UID = "other-admin"
	foreign.Username = "other-admin"
	foreignInput := strings.Replace(input, `["C","C2"]`, `["FOREIGN"]`, 1)
	if r = call(h.create, foreign, foreignInput, ""); r.Code != 200 {
		t.Fatal("second Holding token create failed")
	}
	r = call(h.list, foreign, "", "")
	if r.Code != 200 || json.Unmarshal(r.Body.Bytes(), &listed) != nil || len(listed.Data) != 1 || listed.Data[0]["holdingCode"] != "OTHER" {
		t.Fatal("metadata escaped the selected Holding")
	}
	if r = call(h.revoke, foreign, "", created.Data.ID); r.Code != 404 {
		t.Fatal("another Holding admin can revoke foreign token")
	}
	r = call(h.list, adminUser, "", "")
	if r.Code != 200 || json.Unmarshal(r.Body.Bytes(), &listed) != nil || len(listed.Data) != 1 || listed.Data[0]["holdingCode"] != "H" {
		t.Fatal("first Holding sees second Holding metadata")
	}
	r = call(h.companies, adminUser, "", "")
	if r.Code != 200 || strings.Contains(r.Body.String(), "FOREIGN") {
		t.Fatal("invalid company options")
	}
	principal, e := authenticateAudience(context.Background(), created.Data.Token, "mcp", h.connect, time.Now().UTC())
	if e != nil || len(principal.CompanyCodes) != 2 || principal.User.BusinessCode != "" || principal.User.BranchUID != "" {
		t.Fatal("multi company token incorrect")
	}
	for _, code := range []string{"C", "C2"} {
		resolved, e := resolveCompany(context.Background(), principal, code, h.connect)
		if e != nil || resolved.User.BusinessCode != code {
			t.Fatal("allowed company denied")
		}
	}
	for _, code := range []string{"", "C3", "FOREIGN"} {
		if _, e := resolveCompany(context.Background(), principal, code, h.connect); e == nil {
			t.Fatal("unselected company accepted")
		}
	}
	for range 2 {
		if r = call(h.revoke, adminUser, "", created.Data.ID); r.Code != 200 {
			t.Fatal("revoke failed")
		}
	}
	var count int
	if e := db.QueryRow(`SELECT count(*) FROM mcp_token_audit WHERE action='revoke'`).Scan(&count); e != nil || count != 1 {
		t.Fatal("revoke not idempotent")
	}
	var revoked sql.NullTime
	if e := db.QueryRow(`SELECT revoked_at FROM mcp_access_tokens WHERE id=$1`, created.Data.ID).Scan(&revoked); e != nil || !revoked.Valid {
		t.Fatal("revoke not persisted")
	}
	// A token never reaches further than its issuer: an ADMIN limited to company C can
	// neither see nor allow C2, a branch-only ADMIN cannot mint a company-wide token, and a
	// holding-wide ADMIN may allow every active company.
	if _, err = db.Exec(`INSERT INTO users VALUES('limited',true),('branch-admin',true),('group-admin',true);
 INSERT INTO holding_members VALUES('H','limited','ADMIN',true,'[{"scopetype":"company","companyuid":"C"}]'),
 ('H','branch-admin','ADMIN',true,'[{"scopetype":"branch","companyuid":"C","branchuid":"B"}]'),
 ('H','group-admin','ADMIN',true,'[{"scopetype":"holding"}]')`); err != nil {
		t.Fatal(err)
	}
	as := func(uid string) models.UserInfo {
		u := adminUser
		u.UID, u.Username = uid, uid
		return u
	}
	companyCodes := func(uid string) []string {
		r := call(h.companies, as(uid), "", "")
		var options struct{ Data []struct{ Code string } }
		if r.Code != 200 || json.Unmarshal(r.Body.Bytes(), &options) != nil {
			t.Fatalf("%s company options: %d", uid, r.Code)
		}
		codes := []string{}
		for _, option := range options.Data {
			codes = append(codes, option.Code)
		}
		return codes
	}
	if got := fmt.Sprint(companyCodes("limited")); got != "[C]" {
		t.Fatalf("limited ADMIN offered %s", got)
	}
	if got := fmt.Sprint(companyCodes("branch-admin")); got != "[]" {
		t.Fatalf("branch-only ADMIN offered %s", got)
	}
	if got := fmt.Sprint(companyCodes("group-admin")); got != "[C C2 C3]" {
		t.Fatalf("holding-wide ADMIN offered %s", got)
	}
	message := func(r *httptest.ResponseRecorder) string {
		var body struct{ Message string }
		_ = json.Unmarshal(r.Body.Bytes(), &body)
		return body.Message
	}
	for _, tc := range []struct{ uid, codes string }{{"limited", `["C2"]`}, {"limited", `["C","C2"]`}, {"branch-admin", `["C"]`}} {
		r = call(h.create, as(tc.uid), strings.Replace(input, `["C","C2"]`, tc.codes, 1), "")
		if r.Code != 400 || !strings.Contains(message(r), "เลือกได้เฉพาะบริษัทที่คุณใช้งานได้เอง") {
			t.Fatalf("%s minted %s beyond own scope: %d %s", tc.uid, tc.codes, r.Code, r.Body.String())
		}
	}
	english := &managementRequest{user: as("limited"), request: httptest.NewRequest("POST", "/mcp-tokens", strings.NewReader(strings.Replace(input, `["C","C2"]`, `["C2"]`, 1))), recorder: httptest.NewRecorder()}
	english.request.Header.Set("Accept-Language", "en")
	if e := h.create(english); e != nil || english.recorder.Code != 400 || !strings.HasPrefix(message(english.recorder), "You can allow only companies") {
		t.Fatalf("scope error not localized: %s", english.recorder.Body.String())
	}
	var minted int
	if e := db.QueryRow(`SELECT count(*) FROM mcp_access_tokens WHERE created_by IN ('limited','branch-admin')`).Scan(&minted); e != nil || minted != 0 {
		t.Fatal("rejected token was stored")
	}
	if r = call(h.create, as("limited"), strings.Replace(input, `["C","C2"]`, `["C"]`, 1), ""); r.Code != 200 {
		t.Fatalf("limited ADMIN denied own company: %s", r.Body.String())
	}
	if r = call(h.create, as("group-admin"), strings.Replace(input, `["C","C2"]`, `["C","C2","C3"]`, 1), ""); r.Code != 200 {
		t.Fatalf("holding-wide ADMIN denied: %s", r.Body.String())
	}
	if _, err = db.Exec(`UPDATE users SET is_active=false WHERE id='admin'`); err != nil {
		t.Fatal(err)
	}
	for _, handler := range []func(microservice.IContext) error{h.list, h.companies, h.create, h.revoke} {
		if r = call(handler, adminUser, input, created.Data.ID); r.Code != 403 {
			t.Fatal("inactive user retained administrator access")
		}
	}
	if _, err = db.Exec(`UPDATE users SET is_active=true WHERE id='admin'`); err != nil {
		t.Fatal(err)
	}
	if _, err = db.Exec(`UPDATE holding_members SET is_active=false WHERE user_id='admin'`); err != nil {
		t.Fatal(err)
	}
	for _, handler := range []func(microservice.IContext) error{h.list, h.create, h.revoke} {
		if r = call(handler, adminUser, input, created.Data.ID); r.Code != 403 {
			t.Fatal("inactive admin accepted")
		}
	}
}
