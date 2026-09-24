//go:build integration

package httpapi

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"smlcloudplatform/pkg/microservice/models"
)

func TestPostgresScopeRevocation(t *testing.T) {
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
	ns := fmt.Sprintf("authaudit_scope_%d", time.Now().UnixNano())
	if _, err = db.Exec("CREATE SCHEMA " + ns + "; SET search_path TO " + ns); err != nil {
		t.Fatal(err)
	}
	defer db.Exec("DROP SCHEMA " + ns + " CASCADE")
	uid := uuid.NewString()
	_, err = db.Exec(`CREATE TABLE users(id uuid,is_active boolean);CREATE TABLE holdings(code text,is_active boolean,profile jsonb);
 CREATE TABLE holding_members(holding_code text,user_id uuid,role text,permission_sets jsonb,access_scopes jsonb,is_active boolean,access_expiry_date date);
 CREATE TABLE companies(holding_code text,code text,is_active boolean);CREATE TABLE branches(holding_code text,company_code text,code text,is_active boolean);
 CREATE TABLE role_permissions(holding_code text,role_code text,permissions jsonb,is_active boolean);
 INSERT INTO holdings VALUES('H',true);INSERT INTO companies VALUES('H','C',true);INSERT INTO branches VALUES('H','C','B',true);
 INSERT INTO role_permissions VALUES('H','USER','[]',true),('H','ACCOUNTING','["gl-journals","gl-journals:update"]',true);`)
	if err != nil {
		t.Fatal(err)
	}
	scopes := `[{"scopetype":"branch","companyuid":"C","branchuid":"B"}]`
	if _, err = db.Exec(`INSERT INTO users VALUES($1,true);`, uid); err != nil {
		t.Fatal(err)
	}
	if _, err = db.Exec(`INSERT INTO holding_members VALUES('H',$1,'user','["ACCOUNTING"]',$2,true)`, uid, scopes); err != nil {
		t.Fatal(err)
	}
	request := scopeContext{user: models.UserInfo{UID: uid, HoldingCode: "H", BusinessCode: "C", BranchUID: "B"}}
	connect := func(database string) (*sql.DB, error) {
		if database != "bcai_projection" {
			t.Fatal("scope metadata not central")
		}
		return db, nil
	}
	if got, e := resolveScope(context.Background(), request, connect); e != nil || !allowed(got.Permissions, "gl-journals", "update") {
		t.Fatalf("active scope rejected: %v", e)
	}
	for _, tc := range []struct{ name, change, restore string }{
		{"role disabled", `UPDATE role_permissions SET is_active=false WHERE role_code='ACCOUNTING'`, `UPDATE role_permissions SET is_active=true`},
		{"company disabled", `UPDATE companies SET is_active=false`, `UPDATE companies SET is_active=true`},
		{"branch disabled", `UPDATE branches SET is_active=false`, `UPDATE branches SET is_active=true`},
		{"member disabled", `UPDATE holding_members SET is_active=false`, `UPDATE holding_members SET is_active=true`},
		{"user disabled", `UPDATE users SET is_active=false`, `UPDATE users SET is_active=true`},
		{"scopes revoked", `UPDATE holding_members SET access_scopes='[]'`, `UPDATE holding_members SET access_scopes='` + scopes + `'`},
		// ใช้ได้ถึงสิ้นวันที่กำหนดตามเขตเวลากลุ่มกิจการ — เมื่อวาน (เวลาไทย) = หมดอายุแล้ว
		{"access expired", `UPDATE holding_members SET access_expiry_date=(now() AT TIME ZONE 'Asia/Bangkok')::date - 1`, `UPDATE holding_members SET access_expiry_date=NULL`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if _, e := db.Exec(tc.change); e != nil {
				t.Fatal(e)
			}
			if _, e := resolveScope(context.Background(), request, connect); e == nil {
				t.Fatal("revocation did not apply to next request")
			}
			if _, e := db.Exec(tc.restore); e != nil {
				t.Fatal(e)
			}
		})
	}
	request.user.BranchUID = ""
	if _, err = resolveScope(context.Background(), request, connect); err == nil {
		t.Fatal("branch grant promoted to company scope")
	}
	if _, err = db.Exec(`UPDATE holding_members SET role='owner',access_scopes='{}'`); err != nil {
		t.Fatal(err)
	}
	if got, e := resolveScope(context.Background(), request, connect); e != nil || !got.Permissions["*"] {
		t.Fatal("active owner lost company access")
	}
	if _, err = db.Exec(`UPDATE holding_members SET access_scopes='[{"scopetype":"company","companyuid":"OTHER"}]'`); err != nil {
		t.Fatal(err)
	}
	if _, err = resolveScope(context.Background(), request, connect); err == nil {
		t.Fatal("explicit owner restriction ignored")
	}
	// A token request is bounded by its issuer's current scope like a session: the token's
	// own allow-list cannot reach a company the issuer is restricted away from.
	tokenRequest := &mcpGLContext{IContext: request, tokenKind: "mcp", tokenID: "verified"}
	if _, err = resolveScope(context.Background(), tokenRequest, connect); err == nil {
		t.Fatal("token reached a company outside its issuer's scope")
	}
	if _, err = db.Exec(`UPDATE holding_members SET access_scopes='[{"scopetype":"holding"}]'`); err != nil {
		t.Fatal(err)
	}
	if got, e := resolveScope(context.Background(), tokenRequest, connect); e != nil || !got.CompanyWide || !got.Permissions["*"] {
		t.Fatalf("holding-wide issuer's token denied: %v", e)
	}
}
