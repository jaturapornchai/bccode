package httpapi

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"io"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"smlcloudplatform/pkg/microservice"
	"smlcloudplatform/pkg/microservice/models"
)

type scopeContext struct {
	microservice.IContext
	user models.UserInfo
}

func (c scopeContext) UserInfo() models.UserInfo { return c.user }

type scopeDriver struct {
	role, sets           string
	scopes               string
	memberErr, branchErr error
	expiry               driver.Value // holding_members.access_expiry_date (nil = ไม่มีกำหนด)
	grants               map[string]string
	grantErrors          map[string]error
}

func (d *scopeDriver) Open(string) (driver.Conn, error) { return &scopeConn{d}, nil }

type scopeConn struct{ d *scopeDriver }

func (c *scopeConn) Prepare(string) (driver.Stmt, error) {
	return nil, errors.New("unexpected prepare")
}
func (c *scopeConn) Close() error              { return nil }
func (c *scopeConn) Begin() (driver.Tx, error) { return nil, errors.New("unexpected transaction") }
func (c *scopeConn) QueryContext(_ context.Context, q string, args []driver.NamedValue) (driver.Rows, error) {
	switch {
	case strings.Contains(q, "FROM companies"):
		return &scopeRows{cols: []string{"code"}, data: []driver.Value{"C"}}, nil
	case strings.Contains(q, "FROM branches"):
		if c.d.branchErr != nil {
			return nil, c.d.branchErr
		}
		return &scopeRows{cols: []string{"code"}, data: []driver.Value{"B"}}, nil
	case strings.Contains(q, "FROM holding_members"):
		if c.d.memberErr != nil {
			return nil, c.d.memberErr
		}
		return &scopeRows{cols: []string{"role", "permission_sets", "access_scopes", "access_expiry_date", "timezone"}, data: []driver.Value{c.d.role, []byte(c.d.sets), []byte(c.d.scopes), c.d.expiry, "Asia/Bangkok"}}, nil
	case strings.Contains(q, "FROM role_permissions"):
		code := fmt.Sprint(args[len(args)-1].Value)
		if err := c.d.grantErrors[code]; err != nil {
			return nil, err
		}
		grants, ok := c.d.grants[code]
		if !ok {
			return &scopeRows{cols: []string{"permissions"}}, nil
		}
		return &scopeRows{cols: []string{"permissions"}, data: []driver.Value{[]byte(grants)}}, nil
	}
	return nil, fmt.Errorf("unexpected query: %s", q)
}

type scopeRows struct {
	cols []string
	data []driver.Value
}

func (r *scopeRows) Columns() []string { return r.cols }
func (r *scopeRows) Close() error      { return nil }
func (r *scopeRows) Next(values []driver.Value) error {
	if r.data == nil {
		return io.EOF
	}
	copy(values, r.data)
	r.data = nil
	return nil
}

var scopeDriverID atomic.Int64

func scopeDB(t *testing.T, d *scopeDriver) *sql.DB {
	t.Helper()
	name := fmt.Sprintf("scope-fail-closed-%d", scopeDriverID.Add(1))
	sql.Register(name, d)
	db, err := sql.Open(name, "")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	return db
}
func TestScopeFailsClosed(t *testing.T) {
	lookupErr := errors.New("lookup unavailable")
	for _, tc := range []struct {
		name       string
		db         scopeDriver
		branch     string
		connectErr bool
	}{
		{name: "connection failure", connectErr: true},
		{name: "missing or inactive member", db: scopeDriver{memberErr: sql.ErrNoRows}},
		{name: "member query failure", db: scopeDriver{memberErr: lookupErr}},
		{name: "invalid permission sets", db: scopeDriver{role: "STAFF", sets: "broken"}},
		{name: "partial grants then query failure", db: scopeDriver{role: "STAFF", sets: `["EXTRA"]`, grants: map[string]string{"STAFF": `["gl-journals","gl-journals:update"]`}, grantErrors: map[string]error{"EXTRA": lookupErr}}},
		{name: "invalid grant JSON", db: scopeDriver{role: "STAFF", sets: `[]`, grants: map[string]string{"STAFF": "broken"}}},
		{name: "branch query failure", branch: "B", db: scopeDriver{role: "OWNER", sets: `[]`, branchErr: lookupErr}},
		{name: "inactive branch", branch: "B", db: scopeDriver{role: "OWNER", sets: `[]`, branchErr: sql.ErrNoRows}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			tc.db.scopes = `[{"scopetype":"company","companyuid":"C","allbranches":true}]`
			db := scopeDB(t, &tc.db)
			request := scopeContext{user: models.UserInfo{HoldingCode: "H", BusinessCode: "C", UID: "U", BranchUID: tc.branch}}
			scope, err := resolveScope(context.Background(), request, func(string) (*sql.DB, error) {
				if tc.connectErr {
					return nil, lookupErr
				}
				return db, nil
			})
			if err == nil {
				t.Fatal("lookup failure accepted")
			}
			if status, _ := errorPayloadFor(err); status != 403 {
				t.Fatalf("scope failure HTTP status=%d", status)
			}
			if len(scope.Permissions) != 0 {
				t.Fatalf("permissions leaked on failure: %+v", scope.Permissions)
			}
		})
	}
}

// วันหมดอายุการเข้าใช้งาน: ใช้ได้ถึงสิ้นวันที่กำหนด (เขตเวลากลุ่มกิจการ) — เลยแล้วทุกคำขอ GL ได้ 403 user_access_expired
// ไม่ใช่แค่ตอน login (session/token ที่ออกก่อนหมดอายุต้องถูกปิดด้วย)
func TestScopeAccessExpiry(t *testing.T) {
	bangkok := time.FixedZone("ICT", 7*60*60)
	now := time.Now().In(bangkok)
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC) // lib/pq คืน DATE เป็น 00:00 UTC
	request := scopeContext{user: models.UserInfo{HoldingCode: "H", BusinessCode: "C", UID: "U"}}
	resolve := func(expiry driver.Value) (requestScope, error) {
		db := scopeDB(t, &scopeDriver{role: "OWNER", sets: `[]`, scopes: `{}`, expiry: expiry})
		return resolveScope(context.Background(), request, func(string) (*sql.DB, error) { return db, nil })
	}
	for name, expiry := range map[string]driver.Value{"no expiry": nil, "expires today (usable through end of day)": today, "expires tomorrow": today.AddDate(0, 0, 1)} {
		if _, err := resolve(expiry); err != nil {
			t.Fatalf("%s: denied: %v", name, err)
		}
	}
	scope, err := resolve(today.AddDate(0, 0, -1))
	if !errors.Is(err, errAccessExpired) {
		t.Fatalf("expired yesterday: err=%v", err)
	}
	if len(scope.Permissions) != 0 {
		t.Fatalf("permissions leaked for expired member: %+v", scope.Permissions)
	}
	if status, payload := errorPayloadFor(err, "th"); status != 403 || payload.Code != "user_access_expired" || !strings.Contains(payload.Message, "หมดอายุ") {
		t.Fatalf("expired payload = %d %+v", status, payload)
	}
}

func TestScopeValidPermissions(t *testing.T) {
	for _, role := range []string{"OWNER", "ADMIN", "STAFF"} {
		t.Run(role, func(t *testing.T) {
			db := scopeDB(t, &scopeDriver{role: role, sets: `["EXTRA"]`, scopes: `[{"scopetype":"company","companyuid":"C","allbranches":true}]`, grants: map[string]string{"STAFF": `["gl-journals"]`, "EXTRA": `["gl-journals:update"]`}})
			request := scopeContext{user: models.UserInfo{HoldingCode: "H", BusinessCode: "C", UID: "U", BranchUID: "B"}}
			result, err := resolveScope(context.Background(), request, func(database string) (*sql.DB, error) {
				if database != "bcai_projection" {
					t.Fatalf("authorization must use central metadata, got %s", database)
				}
				return db, nil
			})
			if err != nil {
				t.Fatal(err)
			}
			if result.Scope.Company != "C" || result.Scope.Branch != "B" {
				t.Fatalf("scope lost: %+v", result.Scope)
			}
			if role == "STAFF" {
				if result.Permissions["*"] || !allowed(result.Permissions, "gl-journals", "update") || allowed(result.Permissions, "gl-post", "update") {
					t.Fatalf("unexpected grants: %+v", result.Permissions)
				}
			} else if !result.Permissions["*"] {
				t.Fatal("owner/admin lost wildcard")
			}
		})
	}
}

// A holding-wide scope opens every active company/branch but must not grant GL permissions.
func TestHoldingScopeKeepsRolePermissions(t *testing.T) {
	db := scopeDB(t, &scopeDriver{role: "STAFF", sets: `[]`, scopes: `[{"scopetype":"holding"}]`, grants: map[string]string{"STAFF": `["gl-journals"]`}})
	request := scopeContext{user: models.UserInfo{HoldingCode: "H", BusinessCode: "C", UID: "U", BranchUID: "B"}}
	result, err := resolveScope(context.Background(), request, func(string) (*sql.DB, error) { return db, nil })
	if err != nil {
		t.Fatal(err)
	}
	if !result.CompanyWide || result.Scope.Branch != "B" {
		t.Fatalf("holding scope lost company/branch access: %+v", result)
	}
	if result.Permissions["*"] || !allowed(result.Permissions, "gl-journals", "") || allowed(result.Permissions, "gl-post", "") {
		t.Fatalf("holding scope changed permissions: %+v", result.Permissions)
	}
}

func TestCurrentSessionAccessScopes(t *testing.T) {
	for _, tc := range []struct {
		name, scopes, company, branch string
		manager, allowed              bool
	}{
		{name: "empty staff denied", scopes: "[]", company: "C"},
		{name: "empty owner keeps access", scopes: "{}", company: "C", manager: true, allowed: true},
		{name: "invalid owner scope denied", scopes: "broken", company: "C", manager: true},
		{name: "branch cannot become company", scopes: `[{"scopetype":"branch","companyuid":"C","branchuid":"B"}]`, company: "C"},
		{name: "branch exact", scopes: `[{"scopetype":"branch","companyuid":"C","branchuid":"B"}]`, company: "C", branch: "B", allowed: true},
		{name: "other branch denied", scopes: `[{"scopetype":"branch","companyuid":"C","branchuid":"B"}]`, company: "C", branch: "OTHER"},
		{name: "explicit owner scope respected", scopes: `[{"scopetype":"company","companyuid":"OTHER","allbranches":true}]`, company: "C", manager: true},
		{name: "whole company", scopes: `[{"scopetype":"company","companyuid":"C","allbranches":true}]`, company: "C", branch: "B", allowed: true},
		{name: "holding rule covers company", scopes: `[{"scopetype":"holding"}]`, company: "C", allowed: true},
		{name: "holding rule covers branch", scopes: `[{"scopetype":"holding"}]`, company: "C", branch: "B", allowed: true},
		{name: "unknown rule type denied", scopes: `[{"scopetype":"everything","companyuid":"C"}]`, company: "C"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := sessionScopeAllowed([]byte(tc.scopes), tc.manager, tc.company, tc.branch); got != tc.allowed {
				t.Fatalf("allowed=%v want%v", got, tc.allowed)
			}
		})
	}
}
