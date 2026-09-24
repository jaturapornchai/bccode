//go:build integration

package rolepermission

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
	rolemodels "smlcloudplatform/internal/organization/rolepermission/models"
	"smlcloudplatform/pkg/microservice"
	"smlcloudplatform/pkg/microservice/models"
)

type postgresRoleRequest struct {
	microservice.IContext
	user   models.UserInfo
	status int
	data   interface{}
}

func (r *postgresRoleRequest) UserInfo() models.UserInfo { return r.user }
func (r *postgresRoleRequest) Request() *http.Request {
	return httptest.NewRequest("POST", "/organization/role-permission", nil)
}
func (r *postgresRoleRequest) Response(status int, data interface{}) {
	r.status = status
	r.data = data
}
func (r *postgresRoleRequest) ResponseError(status int, message string) { r.status = status }
func (r *postgresRoleRequest) QueryParam(string) string                 { return "" }
func (r *postgresRoleRequest) Header(string) string                     { return "th" }

func TestPostgresRolePermissionAdministration(t *testing.T) {
	dsn := os.Getenv("GL_AUTH_TEST_DSN")
	if dsn == "" {
		t.Skip("GL_AUTH_TEST_DSN required")
	}
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	db.SetMaxOpenConns(1)
	ns := fmt.Sprintf("authaudit_roles_%d", time.Now().UnixNano())
	if _, err = db.Exec("CREATE SCHEMA " + ns + "; SET search_path TO " + ns); err != nil {
		t.Fatal(err)
	}
	defer db.Exec("DROP SCHEMA " + ns + " CASCADE")
	_, err = db.Exec(`CREATE TABLE users(id uuid,is_active boolean);CREATE TABLE holdings(code text,is_active boolean,profile jsonb);
 CREATE TABLE holding_members(holding_code text,user_id uuid,role text,permission_sets jsonb,is_active boolean,access_expiry_date date);
 CREATE TABLE role_permissions(id text,holding_code text,role_code text,names jsonb,permissions jsonb,is_active boolean,created_at timestamptz,updated_at timestamptz,version bigint NOT NULL DEFAULT 0,UNIQUE(holding_code,role_code));
 INSERT INTO holdings VALUES('H',true);INSERT INTO role_permissions VALUES('rp-accounting','H','ACCOUNTING','[{"code":"th","name":"Accounting"}]','["gl-journals"]',true,now(),now());`)
	if err != nil {
		t.Fatal(err)
	}
	uid := uuid.NewString()
	if _, err = db.Exec(`INSERT INTO users VALUES($1,true)`, uid); err != nil {
		t.Fatal(err)
	}
	if _, err = db.Exec(`INSERT INTO holding_members VALUES('H',$1,'user','["ACCOUNTING"]',true)`, uid); err != nil {
		t.Fatal(err)
	}
	r := &postgresRoleRequest{user: models.UserInfo{UID: uid, Username: "claimed-owner", HoldingCode: "H", Role: 2}}
	h := RolePermissionHttp{}
	active := true
	zeroVersion := int64(0)
	input := rolemodels.RolePermissionRequest{RoleCode: "ACCOUNTING", Names: []rolemodels.LocalizedName{{Code: "th", Name: "Accounting"}}, Permissions: []string{"*"}, IsActive: &active, Version: &zeroVersion}
	if err = rolemodels.NormalizeRequest(&input); err != nil {
		t.Fatal(err)
	}
	actions := []func() error{
		func() error { return h.createRolePermissionPostgres(r, db, "H", input) },
		func() error { return h.updateRolePermissionPostgres(r, db, "H", "rp-accounting", input) },
		func() error { return h.deleteRolePermissionPostgres(r, db, "H", "rp-accounting", 0) },
	}
	denyAll := func() {
		t.Helper()
		for _, action := range actions {
			r.status = 0
			if err := action(); err != nil || r.status != 403 {
				t.Fatalf("unauthorized role mutation status=%d err=%v", r.status, err)
			}
		}
	}
	denyAll()
	if err = h.infoMyRolePermissionPostgres(r, db, "H", "claimed-owner"); err != nil || r.status != 200 {
		t.Fatal("active member cannot read own permissions")
	}
	myRole, _ := json.Marshal(r.data)
	if !strings.Contains(string(myRole), `"rolecode":"USER"`) || strings.Contains(string(myRole), `"*"`) {
		t.Fatal("my role escalated using username/session claims")
	}
	if err = h.searchRolePermissionsPostgres(r, db, "H", "", 0, 10); err != nil || r.status != 403 {
		t.Fatal("USER can browse role administration")
	}
	if err = h.infoRolePermissionPostgres(r, db, "H", "rp-accounting"); err != nil || r.status != 403 {
		t.Fatal("USER can inspect role administration")
	}
	var wildcard bool
	if err = db.QueryRow(`SELECT permissions ? '*' FROM role_permissions WHERE role_code='ACCOUNTING'`).Scan(&wildcard); err != nil || wildcard {
		t.Fatal("self-assigned permission set escalated")
	}
	if _, err = db.Exec(`UPDATE holding_members SET role='admin'`); err != nil {
		t.Fatal(err)
	}
	denyAll()
	if _, err = db.Exec(`UPDATE holding_members SET role='owner'`); err != nil {
		t.Fatal(err)
	}
	// Creating a code that already exists must not overwrite it (was a silent upsert).
	if _ = h.createRolePermissionPostgres(r, db, "H", input); r.status != http.StatusConflict {
		t.Fatalf("create of existing ACCOUNTING status=%d, want 409", r.status)
	}
	sales := input
	sales.RoleCode = "SALES"
	if err = h.createRolePermissionPostgres(r, db, "H", sales); err != nil || r.status != 201 {
		t.Fatal("active OWNER denied")
	}
	if err = h.updateRolePermissionPostgres(r, db, "H", "rp-accounting", input); err != nil || r.status != 200 {
		t.Fatal("active OWNER update denied")
	}
	if err = h.deleteRolePermissionPostgres(r, db, "H", "rp-accounting", 1); err != nil || r.status != 200 {
		t.Fatal("active OWNER delete denied")
	}
	for _, tc := range []struct{ change, restore string }{
		{`UPDATE holding_members SET is_active=false`, `UPDATE holding_members SET is_active=true`},
		{`UPDATE users SET is_active=false`, `UPDATE users SET is_active=true`},
		{`UPDATE holdings SET is_active=false`, `UPDATE holdings SET is_active=true`},
		// access expiry date passed: usable through the END of the date only (review 2026-09-24)
		{`UPDATE holding_members SET access_expiry_date = CURRENT_DATE - 2`, `UPDATE holding_members SET access_expiry_date = NULL`},
	} {
		if _, err = db.Exec(tc.change); err != nil {
			t.Fatal(err)
		}
		denyAll()
		if _, err = db.Exec(tc.restore); err != nil {
			t.Fatal(err)
		}
	}
	r.user.UID = uuid.NewString()
	denyAll()
}
