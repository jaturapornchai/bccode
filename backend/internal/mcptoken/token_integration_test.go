//go:build integration

package mcptoken

import (
	"context"
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	"os"
	"testing"
	"time"
)

func TestPostgresAuthenticationScopeAndRevocation(t *testing.T) {
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
	namespace := fmt.Sprintf("mcp_token_test_%d", time.Now().UnixNano())
	if _, err = db.Exec("CREATE SCHEMA " + namespace); err != nil {
		t.Fatal(err)
	}
	defer db.Exec("DROP SCHEMA " + namespace + " CASCADE")
	if _, err = db.Exec("SET search_path TO " + namespace); err != nil {
		t.Fatal(err)
	}
	setup := `CREATE TABLE users(id text,is_active boolean);
 CREATE TABLE holdings(code text,is_active boolean);
 CREATE TABLE holding_members(holding_code text,user_id text,role text,is_active boolean);
 CREATE TABLE companies(holding_code text,code text,is_active boolean);
 CREATE TABLE branches(holding_code text,company_code text,code text,is_active boolean);
 INSERT INTO users VALUES('admin',true);
 INSERT INTO holdings VALUES('H',true);
 INSERT INTO holding_members VALUES('H','admin','ADMIN',true);
 INSERT INTO companies VALUES('H','C',true);
 INSERT INTO branches VALUES('H','C','B',true);`
	if _, err = db.Exec(setup + schema); err != nil {
		t.Fatal(err)
	}
	now := time.Now().UTC().Truncate(time.Microsecond)
	id, raw, hash, _ := generate("H", "mcp")
	_, err = db.Exec(`INSERT INTO mcp_access_tokens(id,holding_code,company_code,branch_code,name,kind,mode,token_hash,created_by,creator_username,created_at,expires_at) VALUES($1,'H','C','B','test','mcp','readonly',$2,'admin','admin',$3,$4)`, id, hash, now, now.Add(time.Hour))
	if err != nil {
		t.Fatal(err)
	}
	connect := func(h string) (*sql.DB, error) {
		if h != ControlDatabase {
			t.Fatal("token data must use the central database")
		}
		return db, nil
	}
	p, err := authenticateAudience(context.Background(), raw, "mcp", connect, now)
	if err != nil {
		t.Fatal(err)
	}
	if p.ID != id || p.Mode != "readonly" || p.User.HoldingCode != "H" || p.User.BusinessCode != "C" || p.User.BranchUID != "B" || p.User.UID != "admin" {
		t.Fatal("incorrect bound principal")
	}
	var count int
	if err = db.QueryRow(`SELECT count(*) FROM mcp_token_audit WHERE action='use'`).Scan(&count); err != nil || count != 1 {
		t.Fatal("missing audit")
	}
	// Deployed credentials must work before an admin initializes the join table.
	if _, err = db.Exec(`DROP TABLE mcp_token_companies`); err != nil {
		t.Fatal(err)
	}
	legacy, e := authenticateAudience(context.Background(), raw, "mcp", connect, now)
	if e != nil || len(legacy.CompanyCodes) != 1 || legacy.CompanyCodes[0] != "C" || legacy.User.BranchUID != "B" {
		t.Fatal("legacy scope changed before migration")
	}
	for range 2 {
		if _, err = db.Exec(schema); err != nil {
			t.Fatal(err)
		}
	}
	var migrated int
	if err = db.QueryRow(`SELECT count(*) FROM mcp_token_companies WHERE token_id=$1 AND holding_code='H' AND company_code='C'`, id).Scan(&migrated); err != nil || migrated != 1 {
		t.Fatal("legacy grant migration not idempotent")
	}
	if _, e = resolveCompany(context.Background(), legacy, "other", connect); e == nil {
		t.Fatal("legacy company widened")
	}
	// Both prefix and persisted audience must agree; swapping a prefix cannot
	// turn a credential into a different kind of credential.
	apiID, apiRaw, apiHash, _ := generate("H", "api")
	_, err = db.Exec(`INSERT INTO mcp_access_tokens(id,holding_code,company_code,branch_code,name,kind,mode,token_hash,created_by,creator_username,created_at,expires_at) VALUES($1,'H','C','B','api test','api','readwrite',$2,'admin','admin',$3,$4)`, apiID, apiHash, now, now.Add(time.Hour))
	if err != nil {
		t.Fatal(err)
	}
	if p, e := authenticateAudience(context.Background(), apiRaw, "api", connect, now); e != nil || p.Kind != "api" {
		t.Fatal("API token rejected by API audience")
	}
	if _, e := authenticateAudience(context.Background(), apiRaw, "mcp", connect, now); e == nil {
		t.Fatal("API token accepted by MCP")
	}
	if _, e := authenticateAudience(context.Background(), raw, "api", connect, now); e == nil {
		t.Fatal("MCP token accepted by API")
	}
	if _, err = db.Exec(`UPDATE mcp_access_tokens SET kind='api' WHERE id=$1`, id); err != nil {
		t.Fatal(err)
	}
	if _, e := authenticateAudience(context.Background(), raw, "mcp", connect, now); e == nil {
		t.Fatal("persisted audience mismatch accepted")
	}
	if _, err = db.Exec(`UPDATE mcp_access_tokens SET kind='mcp' WHERE id=$1`, id); err != nil {
		t.Fatal(err)
	}
	cases := []struct{ name, change, restore string }{
		{"revoked", `UPDATE mcp_access_tokens SET revoked_at=now()`, `UPDATE mcp_access_tokens SET revoked_at=NULL`},
		{"user inactive", `UPDATE users SET is_active=false`, `UPDATE users SET is_active=true`},
		{"holding inactive", `UPDATE holdings SET is_active=false`, `UPDATE holdings SET is_active=true`},
		{"issuer inactive", `UPDATE holding_members SET is_active=false`, `UPDATE holding_members SET is_active=true`},
		{"issuer demoted", `UPDATE holding_members SET role='STAFF'`, `UPDATE holding_members SET role='ADMIN'`},
		{"company inactive", `UPDATE companies SET is_active=false`, `UPDATE companies SET is_active=true`},
		{"wrong company branch", `UPDATE branches SET company_code='other'`, `UPDATE branches SET company_code='C'`},
		{"hash mismatch", `UPDATE mcp_access_tokens SET token_hash=decode(repeat('00',32),'hex')`, ""},
	}
	if _, e := authenticateAudience(context.Background(), raw, "mcp", connect, now.Add(time.Hour)); e == nil {
		t.Fatal("expired token accepted")
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if _, e := db.Exec(tc.change); e != nil {
				t.Fatal(e)
			}
			if _, e := authenticateAudience(context.Background(), raw, "mcp", connect, now); e == nil {
				t.Fatal("authentication did not deny")
			}
			if tc.restore != "" {
				if _, e := db.Exec(tc.restore); e != nil {
					t.Fatal(e)
				}
			}
		})
	}
	if err = db.QueryRow(`SELECT count(*) FROM mcp_token_audit`).Scan(&count); err != nil || count != 3 {
		t.Fatal("denied authentication wrote use event")
	}
}
