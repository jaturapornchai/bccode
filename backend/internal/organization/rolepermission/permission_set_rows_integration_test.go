//go:build integration

package rolepermission

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
	authmodels "smlcloudplatform/internal/authentication/models"
	common "smlcloudplatform/internal/models"
	rolemodels "smlcloudplatform/internal/organization/rolepermission/models"
	"smlcloudplatform/pkg/microservice/models"
)

// rolePermissionOwner prepares a schema holding H with an active OWNER and the given rows
// (id, role_code, is_active); every row starts with permissions ["seed"].
func rolePermissionOwner(t *testing.T, rows ...[3]string) (*sql.DB, *postgresRoleRequest) {
	t.Helper()
	dsn := os.Getenv("GL_AUTH_TEST_DSN")
	if dsn == "" {
		t.Skip("GL_AUTH_TEST_DSN required")
	}
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatal(err)
	}
	db.SetMaxOpenConns(1)
	ns := fmt.Sprintf("authaudit_rolerows_%d", time.Now().UnixNano())
	if _, err = db.Exec("CREATE SCHEMA " + ns + "; SET search_path TO " + ns); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Exec("DROP SCHEMA " + ns + " CASCADE"); db.Close() })
	uid := uuid.NewString()
	_, err = db.Exec(`CREATE TABLE users(id uuid,is_active boolean);CREATE TABLE holdings(code text,is_active boolean,profile jsonb);
 CREATE TABLE holding_members(holding_code text,user_id uuid,role text,permission_sets jsonb,is_active boolean,access_expiry_date date);
 CREATE TABLE role_permissions(id text,holding_code text,role_code text,names jsonb,permissions jsonb,is_active boolean,created_at timestamptz,updated_at timestamptz,version bigint NOT NULL DEFAULT 0,PRIMARY KEY(holding_code,role_code));
 INSERT INTO holdings VALUES('H',true);`)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = db.Exec(`INSERT INTO users VALUES($1,true)`, uid); err != nil {
		t.Fatal(err)
	}
	if _, err = db.Exec(`INSERT INTO holding_members VALUES('H',$1,'owner','[]',true)`, uid); err != nil {
		t.Fatal(err)
	}
	for _, row := range rows {
		if _, err = db.Exec(`INSERT INTO role_permissions VALUES(NULLIF($1,''),'H',$2,'[{"code":"th","name":"ชุดเดิม"}]','["seed"]',$3::boolean,now(),now())`, row[0], row[1], row[2]); err != nil {
			t.Fatal(err)
		}
	}
	return db, &postgresRoleRequest{user: models.UserInfo{UID: uid, Username: "owner", HoldingCode: "H", Role: 2}}
}

func rolePermissionInput(t *testing.T, code string, permissions ...string) rolemodels.RolePermissionRequest {
	t.Helper()
	active := true
	version := int64(0)
	input := rolemodels.RolePermissionRequest{RoleCode: code, Names: []rolemodels.LocalizedName{{Code: "th", Name: "ชุดสิทธิ์ " + code}}, Permissions: permissions, IsActive: &active, Version: &version}
	if err := rolemodels.NormalizeRequest(&input); err != nil {
		t.Fatal(err)
	}
	return input
}

// rolePermissionRows is role_code → permissions JSON, with " off" when the set is switched off.
func rolePermissionRows(t *testing.T, db *sql.DB) map[string]string {
	t.Helper()
	rows, err := db.Query(`SELECT role_code, permissions::text || CASE WHEN is_active THEN '' ELSE ' off' END FROM role_permissions`)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	state := map[string]string{}
	for rows.Next() {
		var code, value string
		if err = rows.Scan(&code, &value); err != nil {
			t.Fatal(err)
		}
		state[code] = value
	}
	return state
}

func requireRoleStatus(t *testing.T, r *postgresRoleRequest, want int, field string) {
	t.Helper()
	if r.status != want {
		t.Fatalf("status = %d, want %d (%+v)", r.status, want, r.data)
	}
	if field != "" {
		raw, _ := json.Marshal(r.data)
		if !strings.Contains(string(raw), `"field":"`+field+`"`) {
			t.Fatalf("response %s does not name field %q", raw, field)
		}
	}
}

// Older data holds "admin" and "ADMIN" (same id) side by side. A create in any spelling is a 409,
// an id that names both rows is a 409, and an exact code touches only its own row.
func TestPostgresRolePermissionLegacyCaseDuplicates(t *testing.T) {
	db, r := rolePermissionOwner(t, [3]string{"rp-admin", "admin", "true"}, [3]string{"rp-admin", "ADMIN", "true"}, [3]string{"", "owner", "true"})
	h := RolePermissionHttp{}

	_ = h.createRolePermissionPostgres(r, db, "H", rolePermissionInput(t, "ADMIN", "*"))
	requireRoleStatus(t, r, http.StatusConflict, "rolecode")
	_ = h.updateRolePermissionPostgres(r, db, "H", "rp-admin", rolePermissionInput(t, "ADMIN", "gl-journal"))
	requireRoleStatus(t, r, http.StatusConflict, "rolecode")
	if state := rolePermissionRows(t, db); state["admin"] != `["seed"]` || state["ADMIN"] != `["seed"]` {
		t.Fatalf("a refused write changed rows: %v", state)
	}

	if err := h.updateRolePermissionPostgres(r, db, "H", "ADMIN", rolePermissionInput(t, "ADMIN", "gl-journal")); err != nil {
		t.Fatal(err)
	}
	requireRoleStatus(t, r, http.StatusOK, "")
	if err := h.deleteRolePermissionPostgres(r, db, "H", "admin", 0); err != nil {
		t.Fatal(err)
	}
	requireRoleStatus(t, r, http.StatusOK, "")
	// Only "owner" exists: "OWNER" reaches it and keeps the stored spelling.
	if err := h.updateRolePermissionPostgres(r, db, "H", "OWNER", rolePermissionInput(t, "OWNER", "gl-report")); err != nil {
		t.Fatal(err)
	}
	requireRoleStatus(t, r, http.StatusOK, "")
	state := rolePermissionRows(t, db)
	if state["ADMIN"] != `["gl-journal"]` || state["admin"] != `["seed"] off` || state["owner"] != `["gl-report"]` || len(state) != 3 {
		t.Fatalf("writes did not target exact rows: %v", state)
	}
	_ = h.updateRolePermissionPostgres(r, db, "H", "missing", rolePermissionInput(t, "MISSING"))
	requireRoleStatus(t, r, http.StatusNotFound, "")
}

// A new set is a plain insert; a deleted set is brought back as the same row; a rename cannot
// take another set's code in any spelling.
func TestPostgresRolePermissionCreateReviveRename(t *testing.T) {
	db, r := rolePermissionOwner(t, [3]string{"rp-accounting", "accounting", "true"}, [3]string{"rp-sales", "SALES", "false"})
	h := RolePermissionHttp{}

	if err := h.createRolePermissionPostgres(r, db, "H", rolePermissionInput(t, "WAREHOUSE", "stock")); err != nil {
		t.Fatal(err)
	}
	requireRoleStatus(t, r, http.StatusCreated, "")
	_ = h.createRolePermissionPostgres(r, db, "H", rolePermissionInput(t, "WAREHOUSE", "*"))
	requireRoleStatus(t, r, http.StatusConflict, "rolecode")
	_ = h.createRolePermissionPostgres(r, db, "H", rolePermissionInput(t, "ACCOUNTING", "*"))
	requireRoleStatus(t, r, http.StatusConflict, "rolecode")

	if err := h.createRolePermissionPostgres(r, db, "H", rolePermissionInput(t, "SALES", "sale-invoice")); err != nil {
		t.Fatal(err)
	}
	requireRoleStatus(t, r, http.StatusCreated, "")
	var id string
	if err := db.QueryRow(`SELECT id FROM role_permissions WHERE role_code='SALES' AND is_active`).Scan(&id); err != nil || id != "rp-sales" {
		t.Fatalf("deleted SALES not brought back as the same row: %q (%v)", id, err)
	}

	_ = h.updateRolePermissionPostgres(r, db, "H", "SALES", rolePermissionInput(t, "ACCOUNTING", "*"))
	requireRoleStatus(t, r, http.StatusConflict, "rolecode")
	state := rolePermissionRows(t, db)
	if state["accounting"] != `["seed"]` || state["SALES"] != `["sale-invoice"]` || state["WAREHOUSE"] != `["stock"]` || len(state) != 3 {
		t.Fatalf("unexpected rows: %v", state)
	}
}

// A member listing "ACCOUNTING" gets the permissions of a set stored as "accounting".
func TestPostgresMyRolePermissionMatchesCodeIgnoringCase(t *testing.T) {
	db, r := rolePermissionOwner(t, [3]string{"", "accounting", "true"})
	if _, err := db.Exec(`UPDATE role_permissions SET permissions='["gl-journal"]'; UPDATE holding_members SET role='user', permission_sets='["ACCOUNTING"]'`); err != nil {
		t.Fatal(err)
	}
	if err := (RolePermissionHttp{}).infoMyRolePermissionPostgres(r, db, "H", "owner"); err != nil || r.status != http.StatusOK {
		t.Fatalf("status=%d err=%v", r.status, err)
	}
	raw, _ := json.Marshal(r.data)
	if !strings.Contains(string(raw), `"permissions":["gl-journal"]`) {
		t.Fatalf("permissions of the lower-case set missing: %s", raw)
	}
}

// NUL in a set name (pasted from a PDF) is a 400 naming the field — PostgreSQL JSONB cannot store it,
// so the save used to fail as a 500 "try again" (adversarial review 2026-09-24).
func TestPostgresRolePermissionNULNamesField(t *testing.T) {
	db, r := rolePermissionOwner(t)
	h := RolePermissionHttp{}
	input := rolePermissionInput(t, "ACCOUNTING", "gl-journals")
	input.Names[0].Name = "ฝ่าย\x00บัญชี"
	_ = h.createRolePermissionPostgres(r, db, "H", input)
	requireRoleStatus(t, r, http.StatusBadRequest, "names[0].name")
	raw, _ := json.Marshal(r.data)
	if !strings.Contains(string(raw), "NUL") || !strings.Contains(string(raw), "ชื่อชุดสิทธิ์") {
		t.Fatalf("message must name the field and the NUL: %s", raw)
	}
	if rows := rolePermissionRows(t, db); len(rows) != 0 {
		t.Fatalf("rejected set stored: %v", rows)
	}
}

// review 2026-09-24: an update carrying an older __v is a 409 (optimistic lock) and changes
// nothing; create answers the stored timestamps and version; the expiry date itself is usable.
func TestPostgresRolePermissionVersionLockAndCreateTimestamps(t *testing.T) {
	db, r := rolePermissionOwner(t)
	h := RolePermissionHttp{}
	if err := h.createRolePermissionPostgres(r, db, "H", rolePermissionInput(t, "ACCOUNTING", "gl-journals")); err != nil {
		t.Fatal(err)
	}
	requireRoleStatus(t, r, http.StatusCreated, "")
	created, ok := r.data.(common.ApiResponse).Data.(RolePermissionItem)
	if !ok || created.CreatedAt.IsZero() || created.UpdatedAt.IsZero() || created.Version != 0 || created.ID != "rp-accounting" {
		t.Fatalf("201 body must carry the stored row: %+v", r.data)
	}

	first := rolePermissionInput(t, "ACCOUNTING", "gl-journals", "gl-report")
	if err := h.updateRolePermissionPostgres(r, db, "H", "rp-accounting", first); err != nil {
		t.Fatal(err)
	}
	requireRoleStatus(t, r, http.StatusOK, "")
	// A second screen still holding __v 0 must not overwrite the save above.
	stale := rolePermissionInput(t, "ACCOUNTING", "*")
	_ = h.updateRolePermissionPostgres(r, db, "H", "rp-accounting", stale)
	requireRoleStatus(t, r, http.StatusConflict, "__v")
	if state := rolePermissionRows(t, db); state["ACCOUNTING"] != `["gl-journals", "gl-report"]` {
		t.Fatalf("stale update changed the row: %v", state)
	}
	fresh := rolePermissionInput(t, "ACCOUNTING", "gl-report")
	next := int64(1)
	fresh.Version = &next
	if err := h.updateRolePermissionPostgres(r, db, "H", "rp-accounting", fresh); err != nil {
		t.Fatal(err)
	}
	requireRoleStatus(t, r, http.StatusOK, "")
	var version int64
	if err := db.QueryRow(`SELECT version FROM role_permissions WHERE role_code = 'ACCOUNTING'`).Scan(&version); err != nil || version != 2 {
		t.Fatalf("version after two saves = %d (%v)", version, err)
	}

	// review 2026-09-24: a delete confirmed on a list still holding __v 1 must not switch off the set
	// saved meanwhile (409 "__v", row stays active); the current __v deletes it.
	_ = h.deleteRolePermissionPostgres(r, db, "H", "rp-accounting", 1)
	requireRoleStatus(t, r, http.StatusConflict, "__v")
	if state := rolePermissionRows(t, db); state["ACCOUNTING"] != `["gl-report"]` {
		t.Fatalf("stale delete changed the row: %v", state)
	}
	if err := h.deleteRolePermissionPostgres(r, db, "H", "rp-accounting", 2); err != nil {
		t.Fatal(err)
	}
	requireRoleStatus(t, r, http.StatusOK, "")
	if state := rolePermissionRows(t, db); state["ACCOUNTING"] != `["gl-report"] off` {
		t.Fatalf("current delete did not switch the set off: %v", state)
	}
	_ = h.deleteRolePermissionPostgres(r, db, "H", "rp-missing", 0)
	requireRoleStatus(t, r, http.StatusNotFound, "")

	// Expiry date = today in the Holding's timezone (no setting → Asia/Bangkok): still usable.
	today := time.Now().In(authmodels.HoldingLocation("")).Format("2006-01-02")
	if _, err := db.Exec(`UPDATE holding_members SET access_expiry_date = $1::date`, today); err != nil {
		t.Fatal(err)
	}
	if err := h.searchRolePermissionsPostgres(r, db, "H", "", 0, 10); err != nil || r.status != http.StatusOK {
		t.Fatalf("member on the last usable day denied: status=%d err=%v", r.status, err)
	}
}
