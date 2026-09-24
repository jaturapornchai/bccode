//go:build integration

package businesstype

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

	_ "github.com/lib/pq"

	common "smlcloudplatform/internal/models"
	"smlcloudplatform/internal/organization/businesstype/models"
	"smlcloudplatform/pkg/microservice"
)

type businessTypeRequest struct {
	microservice.IContext
	status int
	data   interface{}
}

func (r *businessTypeRequest) Request() *http.Request {
	return httptest.NewRequest("POST", "/organization/business-type", nil)
}
func (r *businessTypeRequest) Response(status int, data interface{}) { r.status, r.data = status, data }
func (r *businessTypeRequest) ResponseError(status int, _ string)    { r.status = status }
func (r *businessTypeRequest) QueryParam(string) string              { return "" }
func (r *businessTypeRequest) Header(string) string                  { return "th" }

func businessTypeDB(t *testing.T) *sql.DB {
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
	ns := fmt.Sprintf("businesstype_rows_%d", time.Now().UnixNano())
	if _, err = db.Exec("CREATE SCHEMA " + ns + "; SET search_path TO " + ns); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Exec("DROP SCHEMA " + ns + " CASCADE"); db.Close() })
	// Same shape as centraldb.go business_types; "retail"/"RETAIL" is the legacy duplicate.
	_, err = db.Exec(`CREATE TABLE business_types (id TEXT PRIMARY KEY, holding_code TEXT NOT NULL, code TEXT NOT NULL,
		names JSONB NOT NULL DEFAULT '[]'::jsonb, is_default BOOLEAN NOT NULL DEFAULT false, is_active BOOLEAN NOT NULL DEFAULT true,
		created_at TIMESTAMPTZ NOT NULL DEFAULT now(), updated_at TIMESTAMPTZ NOT NULL DEFAULT now(), UNIQUE (holding_code, code));
	INSERT INTO business_types (id, holding_code, code, names) VALUES
		('bt-retail-lower', 'H', 'retail', '[{"code":"th","name":"ขายปลีก"}]'),
		('bt-retail-upper', 'H', 'RETAIL', '[{"code":"th","name":"ขายปลีก"}]'),
		('bt-wholesale', 'H', 'WHOLESALE', '[{"code":"th","name":"ขายส่ง"}]');
	INSERT INTO business_types (id, holding_code, code, names, is_active) VALUES
		('bt-service', 'H', 'SERVICE', '[{"code":"th","name":"บริการ"}]', false);`)
	if err != nil {
		t.Fatal(err)
	}
	return db
}

func businessTypeInput(code, name string) models.BusinessType {
	names := []common.NameX{{Code: strPtr("th"), Name: strPtr(name)}}
	return models.BusinessType{Code: code, Names: &names}
}

func strPtr(value string) *string { return &value }

func requireBusinessTypeStatus(t *testing.T, r *businessTypeRequest, want int, field string) {
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

func businessTypeNames(t *testing.T, db *sql.DB) map[string]string {
	t.Helper()
	rows, err := db.Query(`SELECT id, (names->0->>'name') || CASE WHEN is_active THEN '' ELSE ' off' END FROM business_types`)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	state := map[string]string{}
	for rows.Next() {
		var id, name string
		if err = rows.Scan(&id, &name); err != nil {
			t.Fatal(err)
		}
		state[id] = name
	}
	return state
}

func TestPostgresBusinessTypeCreateNeverOverwrites(t *testing.T) {
	db := businessTypeDB(t)
	h := BusinessTypeHttp{}
	r := &businessTypeRequest{}

	for _, code := range []string{"WHOLESALE", "wholesale", "Retail"} {
		_ = h.createBusinessTypePostgres(r, db, "H", businessTypeInput(code, "ทับชื่อเดิม"))
		requireBusinessTypeStatus(t, r, http.StatusConflict, "code")
	}
	if err := h.createBusinessTypePostgres(r, db, "H", businessTypeInput("ONLINE", "ขายออนไลน์")); err != nil {
		t.Fatal(err)
	}
	requireBusinessTypeStatus(t, r, http.StatusCreated, "")
	newID := r.data.(common.ApiResponse).ID
	var code string
	if err := db.QueryRow(`SELECT code FROM business_types WHERE id=$1`, newID).Scan(&code); err != nil || code != "ONLINE" {
		t.Fatalf("response id %q does not point at the new row: %q (%v)", newID, code, err)
	}
	// A deleted type comes back as the same row, so branches pointing at its id stay valid.
	if err := h.createBusinessTypePostgres(r, db, "H", businessTypeInput("service", "บริการหลังการขาย")); err != nil {
		t.Fatal(err)
	}
	requireBusinessTypeStatus(t, r, http.StatusCreated, "")
	if id := r.data.(common.ApiResponse).ID; id != "bt-service" {
		t.Fatalf("revived id = %q, want bt-service", id)
	}
	state := businessTypeNames(t, db)
	if state["bt-wholesale"] != "ขายส่ง" || state["bt-retail-lower"] != "ขายปลีก" || state["bt-retail-upper"] != "ขายปลีก" ||
		state["bt-service"] != "บริการหลังการขาย" || len(state) != 5 {
		t.Fatalf("unexpected rows: %v", state)
	}
}

func TestPostgresBusinessTypeUpdateDeleteTargetOneRow(t *testing.T) {
	db := businessTypeDB(t)
	h := BusinessTypeHttp{}
	r := &businessTypeRequest{}

	_ = h.updateBusinessTypePostgres(r, db, "H", "Retail", businessTypeInput("", "ชื่อใหม่"))
	requireBusinessTypeStatus(t, r, http.StatusConflict, "code")
	if err := h.updateBusinessTypePostgres(r, db, "H", "RETAIL", businessTypeInput("", "ขายปลีกหน้าร้าน")); err != nil {
		t.Fatal(err)
	}
	requireBusinessTypeStatus(t, r, http.StatusOK, "")
	if err := h.deleteBusinessTypePostgres(r, db, "H", "bt-retail-lower"); err != nil {
		t.Fatal(err)
	}
	requireBusinessTypeStatus(t, r, http.StatusOK, "")
	if err := h.updateBusinessTypePostgres(r, db, "H", "wholesale", businessTypeInput("", "ขายส่งทั่วไป")); err != nil {
		t.Fatal(err)
	}
	requireBusinessTypeStatus(t, r, http.StatusOK, "")
	_ = h.deleteBusinessTypePostgres(r, db, "H", "missing")
	requireBusinessTypeStatus(t, r, http.StatusNotFound, "")
	state := businessTypeNames(t, db)
	if state["bt-retail-upper"] != "ขายปลีกหน้าร้าน" || state["bt-retail-lower"] != "ขายปลีก off" || state["bt-wholesale"] != "ขายส่งทั่วไป" {
		t.Fatalf("writes did not target exact rows: %v", state)
	}
}

// NUL in a business type name is a 400 naming the field, not a 500 (adversarial review 2026-09-24).
func TestPostgresBusinessTypeNULNamesField(t *testing.T) {
	db := businessTypeDB(t)
	h := BusinessTypeHttp{}
	r := &businessTypeRequest{}
	_ = h.createBusinessTypePostgres(r, db, "H", businessTypeInput("CAFE", "ร้าน\x00กาแฟ"))
	requireBusinessTypeStatus(t, r, http.StatusBadRequest, "names[0].name")
	_ = h.updateBusinessTypePostgres(r, db, "H", "bt-wholesale", businessTypeInput("WHOLESALE", "ขาย\x00ส่ง"))
	requireBusinessTypeStatus(t, r, http.StatusBadRequest, "names[0].name")
	if state := businessTypeNames(t, db); len(state) != 4 || state["bt-wholesale"] != "ขายส่ง" {
		t.Fatalf("rows changed: %v", state)
	}
}
